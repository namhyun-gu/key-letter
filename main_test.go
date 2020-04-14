package main

import (
	"context"
	"github.com/namhyun-gu/key-letter/service"
	"testing"

	"google.golang.org/grpc"
)

var (
	ctx = context.Background()
	guestInfo = &service.GuestInfo{
		Identifier: "Guest",
	}
)

func InitClient() (service.KeyLetterClient, error) {
	conn, err := grpc.Dial(`127.0.0.1:8080`, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return service.NewKeyLetterClient(conn), nil
}

func TestConnection(t *testing.T) {
	_, err := InitClient()
	if err != nil {
		t.Error("Not connected")
	}
}

func TestIssueCode(t *testing.T) {
	client, _ := InitClient()
	code, err := client.IssueCode(ctx, &service.Key{Value: "Test"})
	if err != nil {
		t.Errorf("Can't issue code: %v", err)
	}
	if len(code.Value) == 0 {
		t.Error("Not generated code")
	}
}

func TestVerifyCode_Success(t *testing.T) {
	client, _ := InitClient()

	code, _ := client.IssueCode(ctx, &service.Key{Value: "Test"})

	go func() {
		// Host routine
		stream, _ := client.WaitPermit(ctx)
		_ = stream.Send(&service.WaitPermitRequest{Code: code.Value})
		_, _ = stream.Recv()
		_ = stream.Send(&service.WaitPermitRequest{Permit: true})
	}()

	reply, err := client.VerifyCode(ctx, &service.VerifyRequest{Code: code.Value, GuestInfo: guestInfo})
	if err != nil || reply.Status == service.VerifyStatus_FAILED {
		t.Errorf("Failed verify (reason: %v, err: %v)", reply.Reason, err)
	}
}

func TestVerifyCode_Failed(t *testing.T) {
	client, _ := InitClient()

	code, _ := client.IssueCode(ctx, &service.Key{Value: "Test"})

	go func() {
		// Host routine
		stream, _ := client.WaitPermit(ctx)
		_ = stream.Send(&service.WaitPermitRequest{Code: code.Value})
		_, _ = stream.Recv()
		_ = stream.Send(&service.WaitPermitRequest{Permit: true})
	}()

	reply, err := client.VerifyCode(ctx, &service.VerifyRequest{Code: "0", GuestInfo: guestInfo})
	if err != nil {
		t.Errorf("Error occurred: %v", err)
	}
	if reply.Status == service.VerifyStatus_SUCCESS {
		t.Error("Not correct verify result.")
	}
}

func TestVerifyCode_Reject_Host(t *testing.T) {
	client, _ := InitClient()

	code, _ := client.IssueCode(ctx, &service.Key{Value: "Test"})

	go func() {
		// Host routine
		stream, _ := client.WaitPermit(ctx)
		_ = stream.Send(&service.WaitPermitRequest{Code: code.Value})
		_, _ = stream.Recv()
		_ = stream.Send(&service.WaitPermitRequest{Permit: false})
	}()

	reply, err := client.VerifyCode(ctx, &service.VerifyRequest{Code: code.Value, GuestInfo: guestInfo})
	if err != nil {
		t.Errorf("Error occurred: %v", err)
	}
	if reply.Status == service.VerifyStatus_FAILED && reply.Reason != service.FailedReason_REJECT_HOST {
		t.Error("Not correct verify result.")
	}
}

func TestVerifyCode_No_Host_Waited(t *testing.T) {
	client, _ := InitClient()

	code, _ := client.IssueCode(ctx, &service.Key{Value: "Test"})

	reply, err := client.VerifyCode(ctx, &service.VerifyRequest{Code: code.Value, GuestInfo: guestInfo})
	if err != nil {
		t.Errorf("Error occurred: %v", err)
	}
	if reply.Status == service.VerifyStatus_FAILED && reply.Reason != service.FailedReason_NO_HOST_WAITED {
		t.Error("Not correct verify result.")
	}
}

func TestVerifyCode_Response_Timeout(t *testing.T) {
	client, _ := InitClient()

	code, _ := client.IssueCode(ctx, &service.Key{Value: "Test"})

	go func() {
		// Host routine
		stream, _ := client.WaitPermit(ctx)
		_ = stream.Send(&service.WaitPermitRequest{Code: code.Value})
		_, _ = stream.Recv()
	}()

	reply, err := client.VerifyCode(ctx, &service.VerifyRequest{Code: code.Value, GuestInfo: guestInfo})
	if err != nil {
		t.Errorf("Error occurred: %v", err)
	}
	if reply.Status == service.VerifyStatus_FAILED && reply.Reason != service.FailedReason_RESPONSE_TIMEOUT {
		t.Error("Not correct verify result.")
	}
}
