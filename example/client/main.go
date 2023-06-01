package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	hmac "github.com/yogeshlonkar/go-grpc-hmac"
	"github.com/yogeshlonkar/go-grpc-hmac/example/pb"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, FormatTimestamp: formatter})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	interceptor := hmac.NewClientInterceptor(os.Getenv("key_id"), os.Getenv("secret_key"))
	opts := []grpc.DialOption{
		interceptor.WithUnaryInterceptor(),
		interceptor.WithStreamInterceptor(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial("127.0.0.1:50051", opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewUserServiceClient(conn)
	resp, err := client.GetUser(context.Background(), &pb.GetUserRequest{Name: "unknown"})
	if err != nil {
		log.Error().Err(err).Msg("GetUser failed")
	} else {
		log.Info().Msgf("GetUser resp: %v", resp)
	}
	stream, err := client.ListUsers(context.Background(), &pb.Empty{Empty: "asdsdasdas"})
	for {
		resp, err = stream.Recv()
		if err == io.EOF {
			stream.CloseSend()
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("ListUsers failed")
			break
		}
		log.Info().Msgf("ListUsers resp: %v", resp)
	}
}

func formatter(i interface{}) string {
	if i == nil {
		return ""
	}
	if tt, ok := i.(string); ok {
		ts, err := time.ParseInLocation(time.RFC3339, tt, time.Local)
		if err != nil {
			i = tt
		} else {
			i = ts.Local().Format(time.Kitchen)
		}
	}
	return fmt.Sprintf("\u001B[32m[example-client] \x1b[90m%v\x1b[0m", i)
}
