package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/namhyun-gu/key-letter/app"
	"github.com/namhyun-gu/key-letter/proto"
	"github.com/namhyun-gu/key-letter/util"
)

var (
	port          = flag.Int("port", 8000, "Server port")
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
		if os.Getenv("USE_ENV") != "" {
			parseEnv()
		}
		config = buildConfig()
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", config.Port))
	if err != nil {
		logger.Fatalf("Failed to listen: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
	})

	_, err = redisClient.Ping().Result()
	if err != nil {
		logger.Fatalf("Failed to connect redis: %v", err)
	}

	grpcServer := grpc.NewServer(grpc_middleware.WithUnaryServerChain(
		util.UnaryServerInterceptor(logger),
		grpc_recovery.UnaryServerInterceptor(),
	), grpc_middleware.WithStreamServerChain(
		util.StreamServerInterceptor(logger),
		grpc_recovery.StreamServerInterceptor(),
	))
	proto.RegisterKeyLetterServer(grpcServer, &app.Server{
		Config: config,
		Database: &app.RedisDatabase{
			Client: redisClient,
		},
		DatabaseChannel: &app.RedisDatabaseChannel{
			Client: redisClient,
		},
	})
	logger.Infoln("Start to serve")
	err = grpcServer.Serve(lis)
	if err != nil {
		logger.Fatal("Failed to serve", err)
	}
}

func buildConfig() *util.Config {
	return &util.Config{
		Port: *port,
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

func parseEnv() {
	envPort, _ := strconv.Atoi(os.Getenv("PORT"))
	port = &envPort

	envRedisAddr := os.Getenv("REDIS_ADDR")
	redisAddr = &envRedisAddr

	envRedisPassword := os.Getenv("REDIS_PASSWORD")
	redisPassword = &envRedisPassword

	envOptsIssuer := os.Getenv("OPTS_ISSUER")
	optsIssuer = &envOptsIssuer

	envOptsPeriod := os.Getenv("OPTS_PERIOD")
	if envOptsPeriod != "" {
		period, _ := strconv.Atoi(envOptsPeriod)
		convertPeriod := uint(period)
		optsPeriod = &convertPeriod
	}

	envOptsDigits := os.Getenv("OPTS_DIGITS")
	if envOptsDigits != "" {
		digits, _ := strconv.Atoi(envOptsDigits)
		optsDigits = &digits
	}

	envOptsAlgorithm := os.Getenv("OPTS_ALGORITHM")
	optsAlgorithm = &envOptsAlgorithm
}
