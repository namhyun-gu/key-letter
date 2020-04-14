package util

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

var codeToStr = map[codes.Code]string{
	codes.OK:                 `OK`,
	codes.Canceled:           `CANCELLED`,
	codes.Unknown:            `UNKNOWN`,
	codes.InvalidArgument:    `INVALID_ARGUMENT`,
	codes.DeadlineExceeded:   `DEADLINE_EXCEEDED`,
	codes.NotFound:           `NOT_FOUND`,
	codes.AlreadyExists:      `ALREADY_EXISTS`,
	codes.PermissionDenied:   `PERMISSION_DENIED`,
	codes.ResourceExhausted:  `RESOURCE_EXHAUSTED`,
	codes.FailedPrecondition: `FAILED_PRECONDITION`,
	codes.Aborted:            `ABORTED`,
	codes.OutOfRange:         `OUT_OF_RANGE`,
	codes.Unimplemented:      `UNIMPLEMENTED`,
	codes.Internal:           `INTERNAL`,
	codes.Unavailable:        `UNAVAILABLE`,
	codes.DataLoss:           `DATA_LOSS`,
	codes.Unauthenticated:    `UNAUTHENTICATED`,
}

func UnaryServerInterceptor(logger grpclog.LoggerV2) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		logger.Infof("Called %s (%s)\n", info.FullMethod, codeToStr[status.Code(err)])
		return resp, err
	}
}

func StreamServerInterceptor(logger grpclog.LoggerV2) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		logger.Infof("Called %s[stream] (%s)\n", info.FullMethod, codeToStr[status.Code(err)])
		return err
	}
}
