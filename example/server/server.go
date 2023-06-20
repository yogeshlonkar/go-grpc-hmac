package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	hmac "github.com/yogeshlonkar/go-grpc-hmac"
	"go-grpc-hmac/example/pb"
)

type Servicer struct {
	pb.UnimplementedUserServiceServer
}

func (s Servicer) GetUser(_ context.Context, request *pb.GetUserRequest) (*pb.User, error) {
	log.Info().Msgf("GetUser req: %v", request)
	return &pb.User{
		Name:  "unknown",
		Email: "test@example.com",
	}, nil
}

func (s Servicer) ListUsers(_ *pb.Empty, server pb.UserService_ListUsersServer) error {
	log.Info().Msg("ListUsers req received")
	_ = server.Send(&pb.User{
		Name:  "unknown",
		Email: "test@example.com",
	})
	return server.Send(&pb.User{
		Name:  "known",
		Email: "test2@example.com",
	})
}

type server struct {
	addr   string
	server *grpc.Server
}

func (s *server) handleShutdown(shutdown <-chan bool, done chan<- bool) {
	<-shutdown
	s.server.GracefulStop()
	log.Info().Msg("server server stopped")
	done <- true
}

func (s *server) start() {
	s.addr = ":50051"
	if os.Getenv("GRPC_PORT") != "" {
		s.addr = fmt.Sprintf(":%s", os.Getenv("GRPC_PORT"))
	}
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	if err = s.server.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve server")
	}
}

func (s *server) getSecrets(_ context.Context, keyId string) (string, error) {
	secrets := map[string]string{
		os.Getenv("key_id"): os.Getenv("secret_key"),
	}
	if secret, ok := secrets[keyId]; ok {
		return secret, nil
	}
	return "", fmt.Errorf("key not found")
}

func (s *server) setup() {
	interceptor := hmac.NewServerInterceptor(s.getSecrets)
	opts := []grpc.ServerOption{
		interceptor.UnaryInterceptor(),
		interceptor.StreamInterceptor(),
		grpc.Creds(insecure.NewCredentials()),
	}
	s.server = grpc.NewServer(opts...)
}
