module go-grpc-hmac/example

go 1.20

require (
	github.com/rs/zerolog v1.29.1
	github.com/yogeshlonkar/go-grpc-hmac v0.1.0
	google.golang.org/grpc v1.56.0
	google.golang.org/protobuf v1.30.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/net v0.11.0 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/text v0.10.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230530153820-e85fd2cbaebc // indirect
)

replace github.com/yogeshlonkar/go-grpc-hmac => ../
