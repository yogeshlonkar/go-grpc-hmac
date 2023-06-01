package hmac

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestStreamClientInterceptor(t *testing.T) {
	handlerCalled := false
	handler := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		handlerCalled = true
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Errorf("StreamClientInterceptor() expected metadata to be present")
		}
		hmacSign := md.Get("x-hmac-signature")
		message, _ := NewMessage(nil, "method1")
		if len(hmacSign) < 1 || hmacSign[0] != String("secret1", message) {
			t.Errorf("StreamClientInterceptor() expected signature to match")
		}
		hmacKeyID := md.Get("x-hmac-key-id")
		if len(hmacKeyID) < 1 || hmacKeyID[0] != "key1" {
			t.Errorf("StreamClientInterceptor() expected key id to match")
		}
		return nil, nil
	}
	c := &clientInterceptor{
		hmacKeyId:  "key1",
		hmacSecret: "secret1",
	}
	_, err := c.StreamClientInterceptor(context.Background(), &grpc.StreamDesc{}, nil, "method1", handler)
	if err != nil {
		t.Fatalf("StreamClientInterceptor() expected error to be nil got error = %v", err)
	}
	if !handlerCalled {
		t.Errorf("StreamClientInterceptor() expected handler to be called")
	}
}

func TestUnaryClientInterceptor(t *testing.T) {
	req := &struct{ field string }{field: "value"}
	handlerCalled := false
	handler := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		handlerCalled = true
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Errorf("UnaryClientInterceptor() expected metadata to be present")
		}
		hmacSign := md.Get("x-hmac-signature")
		message, _ := NewMessage(req, "method1")
		if len(hmacSign) < 1 || hmacSign[0] != String("secret1", message) {
			t.Errorf("UnaryClientInterceptor() expected signature to match")
		}
		hmacKeyID := md.Get("x-hmac-key-id")
		if len(hmacKeyID) < 1 || hmacKeyID[0] != "key1" {
			t.Errorf("UnaryClientInterceptor() expected key id to match")
		}
		return nil
	}
	c := &clientInterceptor{
		hmacKeyId:  "key1",
		hmacSecret: "secret1",
	}
	err := c.UnaryClientInterceptor(context.Background(), "method1", &grpc.StreamDesc{}, req, nil, handler)
	if err != nil {
		t.Fatalf("UnaryClientInterceptor() expected error to be nil got error = %v", err)
	}
	if !handlerCalled {
		t.Errorf("UnaryClientInterceptor() expected handler to be called")
	}
}
