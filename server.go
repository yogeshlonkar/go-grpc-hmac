package hmac

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrUnauthorized is returned when the request is not authorized for any reason.
var ErrUnauthorized = status.Errorf(codes.Unauthenticated, "Unauthenticated")

// ServerInterceptor that implements HMAC authentication for gRPC servers.
type ServerInterceptor interface {
	// StreamInterceptor a grpc.ServerOption that can be passed to grpc.NewServer
	StreamInterceptor() grpc.ServerOption
	// StreamServerInterceptor a grpc.StreamInterceptor that authenticates methods with client or server stream requests
	StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error
	// UnaryInterceptor a grpc.ServerOption that can be passed to grpc.NewServer
	UnaryInterceptor() grpc.ServerOption
	// UnaryServerInterceptor a grpc.UnaryServerInterceptor that authenticates methods with unary (proto message) requests
	UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
	IgnoredMethods(methods ...string)
	ClearIgnores()
}

type serverInterceptor struct {
	auth   func(ctx context.Context, message string) error
	ignore []string
}

// GetSecret is a function that returns the secret for a given keyId.
// Returns an empty string in case the keyId is not found instead of an error.
// If the function returns an error, the request is rejected.
type GetSecret func(ctx context.Context, keyId string) (secret string, err error)

// NewServerInterceptor returns a new server interceptor that authenticates requests using GetSecret.
func NewServerInterceptor(getSecret GetSecret) ServerInterceptor {
	return &serverInterceptor{authForSecrets(getSecret), make([]string, 0)}
}

// StreamInterceptor a grpc.ServerOption that can be passed to grpc.NewServer.
func (s *serverInterceptor) StreamInterceptor() grpc.ServerOption {
	return grpc.StreamInterceptor(s.StreamServerInterceptor)
}

// StreamServerInterceptor a grpc.StreamInterceptor that authenticates methods with client or server stream requests.
func (s *serverInterceptor) StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if s.ignored(info.FullMethod) {
		logger.Printf("ignoring steaming method %s", info.FullMethod)
		return handler(srv, ss)
	}
	message, err := NewMessage(nil, info.FullMethod)
	if err != nil {
		return err
	}
	if err = s.auth(ss.Context(), message); err != nil {
		logger.Printf("auth error on streaming method %s: %q", info.FullMethod, err)
		return ErrUnauthorized
	}
	return handler(srv, ss)
}

// UnaryInterceptor a grpc.ServerOption that can be passed to grpc.NewServer.
func (s *serverInterceptor) UnaryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(s.UnaryServerInterceptor)
}

// UnaryServerInterceptor a grpc.UnaryServerInterceptor that authenticates methods with unary (proto message) requests.
func (s *serverInterceptor) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if s.ignored(info.FullMethod) {
		logger.Printf("ignoring unary method %s", info.FullMethod)
		return handler(ctx, req)
	}
	message, err := NewMessage(req, info.FullMethod)
	if err != nil {
		return nil, err
	}
	if err = s.auth(ctx, message); err != nil {
		logger.Printf("auth error on unary method %s: %q", info.FullMethod, err)
		return nil, ErrUnauthorized
	}
	return handler(ctx, req)
}

// IgnoredMethods from authentication.
func (s *serverInterceptor) IgnoredMethods(methods ...string) {
	s.ignore = append(s.ignore, methods...)
}

// ClearIgnores clears the ignored methods.
func (s *serverInterceptor) ClearIgnores() {
	s.ignore = make([]string, 0)
}

func (s *serverInterceptor) ignored(method string) bool {
	for _, m := range s.ignore {
		if m == method {
			return true
		}
	}
	return false
}
