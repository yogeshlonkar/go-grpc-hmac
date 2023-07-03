package hmac

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNewMessage(t *testing.T) {
	type args struct {
		req    interface{}
		method string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "NoRequest",
			args: args{nil, "method1"},
			want: "method=method1",
		},
		{
			name: "EmtpyRequest",
			args: args{&struct{ field int }{}, "method2"},
			want: "method=method2",
		},
		{
			name: "RequestWithFields",
			args: args{&struct {
				Field1 int `json:"field1"`
			}{1}, "method3"},
			want: `request={"field1":1};method=method3`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMessage(tt.args.req, tt.args.method)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NewMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authForSecrets(t *testing.T) {
	type args struct {
		getSecret func(context.Context, string) (string, error)
		ctx       context.Context //nolint:containedctx
		message   string
	}
	tests := []struct {
		name string
		args
		want error
	}{
		{
			"NoMetadata",
			args{
				ctx: context.Background(),
			},
			ErrMissingMetadata,
		},
		{
			"NoHmacSignature",
			args{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{}),
			},
			ErrMissingHmac,
		},
		{
			"NoHmacKeyID",
			args{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{"x-hmac-signature": []string{"signature"}}),
			},
			ErrMissingHmacKeyID,
		},
		{
			"FailedToGetSecret",
			args{
				getSecret: func(context.Context, string) (string, error) { return "", errors.New("something went wrong") },
				ctx:       metadata.NewIncomingContext(context.Background(), metadata.MD{"x-hmac-signature": []string{"signature"}, "x-hmac-key-id": []string{"key-id"}}),
			},
			status.Errorf(codes.Internal, "something went wrong"),
		},
		{
			"InvalidHmacKeyID",
			args{
				getSecret: func(context.Context, string) (string, error) { return "", nil },
				ctx:       metadata.NewIncomingContext(context.Background(), metadata.MD{"x-hmac-signature": []string{"signature"}, "x-hmac-key-id": []string{"key-id"}}),
			},
			ErrInvalidHmacKeyID,
		},
		{
			"InvalidHmacSignature",
			args{
				getSecret: func(context.Context, string) (string, error) { return "secret", nil },
				ctx:       metadata.NewIncomingContext(context.Background(), metadata.MD{"x-hmac-signature": []string{"signature"}, "x-hmac-key-id": []string{"key-id"}}),
				message:   "plain-text",
			},
			ErrInvalidHmacSignature,
		},
		{
			"ValidHmacSignature",
			args{
				getSecret: func(context.Context, string) (string, error) { return "secret", nil },
				ctx:       metadata.NewIncomingContext(context.Background(), metadata.MD{"x-hmac-signature": []string{"10UnPiUX0BMx6XS+VrOwCo0S8L/K58ySRb+VUT/xuvU="}, "x-hmac-key-id": []string{"key-id"}}),
				message:   "plain-text",
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := authForSecrets(tt.args.getSecret)
			if got := auth(tt.ctx, tt.message); tt.want != got && !errors.Is(got, tt.want) { //nolint:errorlint
				t.Errorf("NewMessage() return got = %v, want %v", got, tt.want)
			}
		})
	}
}
