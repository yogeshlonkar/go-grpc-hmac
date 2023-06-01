# go-grcp-hmac

HMAC Client and Server Interceptor for golang grpc

## 💻 Install

```shell
go get github.com/travix/protoc-gen-gotf
```

## ✏️ [Example]

## 🧑‍💻 Usage

### Server

Add required interceptors to grpc server options

```go
// getSecrets implements hmac.GetSecret func type that returns secret key for given keyId
interceptor := hmac.NewServerInterceptor(getSecrets)
opts := []grpc.ServerOption{
    interceptor.UnaryInterceptor(),
    interceptor.StreamInterceptor(),
    // ... other options
}
server := grpc.NewServer(opts...)
```

### Client

Add required interceptors to grpc client options

```go
// keyId for which secret_key is returned by hmac.GetSecret func type on server side
interceptor := hmac.NewClientInterceptor(keyId, secret_key)
opts := []grpc.DialOption{
    interceptor.WithUnaryInterceptor(),
    interceptor.WithStreamInterceptor(),
	// ... other options
}
conn, err := grpc.Dial(addr, opts...)
```

## 🔐 HMAC Authentication

HMAC is generated using
 
 - Request payload encoded using [gob encoder], full method name concatenated with `;` as separator
 - If request payload is empty, then only full method name is used.
 - Generated message is encrypted with given secret using [SHA512_256]

Authentication flow

 - Client interceptor adds `x-hmac-key-id` and `x-hmac-signature` to outgoing request context.
 - Server interceptor reads `x-hmac-key-id` and `x-hmac-signature` from incoming request context and verifies the signature using secret independently fetched on server using given key id.
 - If signature is valid, request is processed, otherwise `Unauthenticated` error is returned.

[Example]: ./example/README.md
[gob encoder]: https://pkg.go.dev/encoding/gob#Encoder.Encode
[SHA512_256]: https://pkg.go.dev/crypto/sha512#New512_256
