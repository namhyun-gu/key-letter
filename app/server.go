package app

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/namhyun-gu/key-letter/proto"
	"github.com/namhyun-gu/key-letter/util"
)

type Server struct {
	Config          *util.Config
	Database        Database
	DatabaseChannel DatabaseChannel
}

func (server *Server) IssueCode(ctx context.Context, key *proto.Key) (*proto.Code, error) {
	opts := server.Config.Opts

	var digits otp.Digits
	if opts.Digits == 8 {
		digits = otp.DigitsEight
	} else {
		digits = otp.DigitsSix
	}

	var algorithm otp.Algorithm
	if opts.Algorithm == "SHA256" {
		algorithm = otp.AlgorithmSHA256
	} else if opts.Algorithm == "SHA512" {
		algorithm = otp.AlgorithmSHA512
	} else if opts.Algorithm == "MD5" {
		algorithm = otp.AlgorithmMD5
	} else {
		algorithm = otp.AlgorithmSHA1
	}

	generateKey, err := totp.Generate(totp.GenerateOpts{
		Issuer:      opts.Issuer,
		AccountName: key.Value,
		Period:      opts.Period,
		Digits:      digits,
		Algorithm:   algorithm,
	})
	if generateKey == nil {
		return nil, err
	}

	code, err := totp.GenerateCode(generateKey.Secret(), time.Now())
	if code == "" {
		return nil, err
	}

	err = server.Database.CreateStore(code, Store{Key: key.Value, Secret: generateKey.Secret()}, time.Duration(opts.Period))
	if err != nil {
		return nil, err
	}
	return &proto.Code{Value: code}, nil
}

func (server *Server) VerifyCode(ctx context.Context, request *proto.VerifyRequest) (*proto.VerifyReply, error) {
	store, err := server.Database.ReadStore(request.Code)
	if err != nil {
		return &proto.VerifyReply{
			Status: proto.VerifyStatus_FAILED,
			Reason: proto.FailedReason_INTERNAL_ERR,
		}, err
	}

	if store == nil {
		return &proto.VerifyReply{
			Status: proto.VerifyStatus_FAILED,
			Reason: proto.FailedReason_AUTH_FAILED,
		}, nil
	}

	if !totp.Validate(request.Code, store.Secret) {
		return &proto.VerifyReply{
			Status: proto.VerifyStatus_FAILED,
			Reason: proto.FailedReason_AUTH_FAILED,
		}, nil
	}

	_ = server.Database.DeleteStore(request.Code)

	// guest wait while permitted exchange by host
	subscriber := server.DatabaseChannel.Subscribe(request.GuestInfo.Identifier)
	defer subscriber.Close()

	guestInfoJson, err := json.Marshal(request.GuestInfo)
	if err != nil {
		return &proto.VerifyReply{
			Status: proto.VerifyStatus_FAILED,
			Reason: proto.FailedReason_INTERNAL_ERR,
		}, err
	}

	// send guest information to host for permit
	subs, err := server.DatabaseChannel.Publish(request.Code, string(guestInfoJson))
	if err != nil {
		return &proto.VerifyReply{
			Status: proto.VerifyStatus_FAILED,
			Reason: proto.FailedReason_INTERNAL_ERR,
		}, err
	}

	if subs == 0 {
		return &proto.VerifyReply{
			Status: proto.VerifyStatus_FAILED,
			Reason: proto.FailedReason_NO_HOST_WAITED,
		}, nil
	}

	time.AfterFunc(time.Duration(server.Config.Timeout)*time.Second, func() {
		subscriber.Close()
	})

	channel := subscriber.Channel()
	for message := range channel {
		permit, err := strconv.ParseBool(message.Payload)
		if err != nil {
			return &proto.VerifyReply{
				Status: proto.VerifyStatus_FAILED,
				Reason: proto.FailedReason_INTERNAL_ERR,
			}, err
		}
		if !permit {
			return &proto.VerifyReply{
				Status: proto.VerifyStatus_FAILED,
				Reason: proto.FailedReason_REJECT_HOST,
			}, nil
		}

		return &proto.VerifyReply{
			Status: proto.VerifyStatus_SUCCESS,
			Key:    store.Key,
		}, nil
	}

	return &proto.VerifyReply{
		Status: proto.VerifyStatus_FAILED,
		Reason: proto.FailedReason_RESPONSE_TIMEOUT,
	}, nil
}

func (server *Server) WaitPermit(permitServer proto.KeyLetter_WaitPermitServer) error {
	request, err := permitServer.Recv()
	if err != nil {
		return err
	}

	// host wait while access guest
	subscriber := server.DatabaseChannel.Subscribe(request.Code)
	defer subscriber.Close()

	channel := subscriber.Channel()
	for message := range channel {
		var guestInfo *proto.GuestInfo
		err := json.Unmarshal([]byte(message.Payload), &guestInfo)
		if err != nil {
			return err
		}

		err = permitServer.Send(guestInfo)
		if err != nil {
			return err
		}

		permitResponse, err := permitServer.Recv()
		if permitResponse == nil {
			continue
		}

		_, err = server.DatabaseChannel.Publish(guestInfo.Identifier, permitResponse.Permit)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
