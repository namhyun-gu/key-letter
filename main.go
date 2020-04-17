package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/go-redis/redis"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	"github.com/namhyun-gu/key-letter/app"
	"github.com/namhyun-gu/key-letter/proto"
	"github.com/namhyun-gu/key-letter/util"
)

var (
	port          = flag.Int("port", 8000, "Server port")
	certFilePath = flag.String("cert", "", "Cert file path")
	keyFilePath = flag.String("cert-key", "", "Cert key file path")
	redisAddr     = flag.String("redis-addr", "", "Redis endpoint")
	redisPassword = flag.String("redis-password", "", "Redis password")
	optsIssuer    = flag.String("opts-issuer", "", "Issuer for generate code")
	optsPeriod    = flag.Uint("opts-period", 30, "Code period")
	optsDigits    = flag.Int("opts-digits", 6, "Code digits")
	optsAlgorithm = flag.String("opts-algorithm", "", "Algorithm for generate code")
)

func main() {
	logger := grpclog.NewLoggerV2(os.Stdout, os.Stderr, os.Stderr)
	grpclog.SetLoggerV2(logger)

	config := util.GetConfig()
	if config == nil {
		flag.Parse()
		config = buildConfig()
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		logger.Fatalf("Failed to connect redis: %v", err)
	}

	unaryServerMiddleware := grpc_middleware.WithUnaryServerChain(
		util.UnaryServerInterceptor(logger),
		grpc_recovery.UnaryServerInterceptor(),
	)

	streamServerMiddleware := grpc_middleware.WithStreamServerChain(
		util.StreamServerInterceptor(logger),
		grpc_recovery.StreamServerInterceptor(),
	)

	creds, err := credentials.NewServerTLSFromFile(config.CertFile, config.CertKeyFile)

	var grpcServer *grpc.Server

	if creds != nil {
		logger.Infoln("Enabled SSL/TLS")
		grpcServer = grpc.NewServer(grpc.Creds(creds), unaryServerMiddleware, streamServerMiddleware)
	} else {
		logger.Infoln("Disabled SSL/TLS")
		grpcServer = grpc.NewServer(unaryServerMiddleware, streamServerMiddleware)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", config.Port))
	if err != nil {
		logger.Fatalf("Failed to listen: %v", err)
	}

	proto.RegisterKeyLetterServer(grpcServer, &app.Server{
		Config: config,
		Database: &app.RedisDatabase{
			Client: redisClient,
		},
		DatabaseChannel: &app.RedisDatabaseChannel{
			Client: redisClient,
		},
	})

	logger.Infof("Started to serve (:%d)\n", *port)

	err = grpcServer.Serve(lis)
	if err != nil {
		logger.Fatal("Failed to serve", err)
	}
}

func buildConfig() *util.Config {
	return &util.Config{
		Port:        *port,
		CertFile:    *certFilePath,
		CertKeyFile: *keyFilePath,
		Redis: struct {
			Addr     string
			Password string
		}{
			Addr:     *redisAddr,
			Password: *redisPassword,
		},
		Opts: struct {
			Issuer    string
			Period    uint
			Digits    int
			Algorithm string
		}{
			Issuer:    *optsIssuer,
			Period:    *optsPeriod,
			Digits:    *optsDigits,
			Algorithm: *optsAlgorithm,
		},
	}
}