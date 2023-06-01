package hmac

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClientInterceptor is a grpc client interceptor that adds HMAC authentication to outgoing requests.
type ClientInterceptor interface {
	// StreamClientInterceptor a grpc.StreamClientInterceptor that adds HMAC authentication to outgoing requests.
	StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error)
	// UnaryClientInterceptor a grpc.UnaryClientInterceptor that adds HMAC authentication to outgoing requests.
	UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error
	// WithStreamInterceptor returns a grpc.DialOption that can be passed to grpc.Dial
	WithStreamInterceptor() grpc.DialOption
	// WithUnaryInterceptor returns a grpc.DialOption that can be passed to grpc.Dial
	WithUnaryInterceptor() grpc.DialOption
}

type clientInterceptor struct {
	hmacKeyId, hmacSecret string
}

// NewClientInterceptor returns a new client interceptor that adds HMAC authentication to outgoing requests.
// The hmacKeyId and hmacSecret are used to sign the request.
func NewClientInterceptor(hmacKeyId, hmacSecret string) ClientInterceptor {
	return &clientInterceptor{hmacKeyId, hmacSecret}
}

// StreamClientInterceptor a grpc.StreamClientInterceptor that adds HMAC authentication to outgoing requests.
func (c *clientInterceptor) StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	message, err := NewMessage(nil, method)
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-hmac-key-id", c.hmacKeyId, "x-hmac-signature", String(c.hmacSecret, message))
	return streamer(ctx, desc, cc, method, opts...)
}

// UnaryClientInterceptor a grpc.UnaryClientInterceptor that adds HMAC authentication to outgoing requests.
func (c *clientInterceptor) UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	message, err := NewMessage(req, method)
	if err != nil {
		return err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-hmac-key-id", c.hmacKeyId, "x-hmac-signature", String(c.hmacSecret, message))
	return invoker(ctx, method, req, reply, cc, opts...)
}

// WithStreamInterceptor returns a grpc.DialOption that can be passed to grpc.Dial.
func (c *clientInterceptor) WithStreamInterceptor() grpc.DialOption {
	return grpc.WithStreamInterceptor(c.StreamClientInterceptor)
}

// WithUnaryInterceptor returns a grpc.DialOption that can be passed to grpc.Dial.
func (c *clientInterceptor) WithUnaryInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(c.UnaryClientInterceptor)
}
