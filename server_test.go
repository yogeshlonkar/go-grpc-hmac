package hmac

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

type mockServerStream struct {
	grpc.ServerStream
}

func (m *mockServerStream) Context() context.Context {
	return context.Background()
}

func TestStreamServerInterceptor(t *testing.T) {
	handlerCalled := false
	handler := func(srv interface{}, ss grpc.ServerStream) error {
		handlerCalled = true
		return nil
	}
	mockCalled := false
	mockedAuth := func(ctx context.Context, message string) error {
		mockCalled = true
		expectedMessage, _ := NewMessage(nil, "method1")
		if message != expectedMessage {
			t.Errorf("StreamServerInterceptor() expected message to be %v got %v", expectedMessage, message)
		}
		return nil
	}
	s := &serverInterceptor{
		auth: mockedAuth,
	}
	err := s.StreamServerInterceptor(nil, &mockServerStream{}, &grpc.StreamServerInfo{FullMethod: "method1"}, handler)
	if err != nil {
		t.Fatalf("StreamServerInterceptor() expected error to be nil got error = %v", err)
	}
	if !handlerCalled {
		t.Errorf("StreamServerInterceptor() expected handler to be called")
	}
	if !mockCalled {
		t.Errorf("StreamServerInterceptor() expected auth to be called")
	}
}

func TestUnaryServerInterceptor(t *testing.T) {
	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return nil, nil
	}
	req := &struct{ field string }{field: "value"}
	mockCalled := false
	mockedAuth := func(ctx context.Context, message string) error {
		mockCalled = true
		expectedMessage, _ := NewMessage(req, "method1")
		if message != expectedMessage {
			t.Errorf("UnaryServerInterceptor() expected message to be %v got %v", expectedMessage, message)
		}
		return nil
	}
	s := &serverInterceptor{
		auth: mockedAuth,
	}
	_, err := s.UnaryServerInterceptor(context.Background(), req, &grpc.UnaryServerInfo{FullMethod: "method1"}, handler)
	if err != nil {
		t.Fatalf("UnaryServerInterceptor() expected error to be nil got error = %v", err)
	}
	if !handlerCalled {
		t.Errorf("UnaryServerInterceptor() expected handler to be called")
	}
	if !mockCalled {
		t.Errorf("UnaryServerInterceptor() expected auth to be called")
	}
}
