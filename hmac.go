package hmac

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const emptyBracketLength = 2

var (
	logger                  = log.New(io.Discard, "[go-grpc-hmac] ", log.LstdFlags|log.LUTC)
	ErrInvalidHmacKeyID     = status.Errorf(codes.Unauthenticated, "invalid x-hmac-key-id")
	ErrInvalidHmacSignature = status.Errorf(codes.Unauthenticated, "invalid x-hmac-signature")
	ErrMissingHmac          = status.Errorf(codes.Unauthenticated, "missing x-hmac-signature metadata")
	ErrMissingHmacKeyID     = status.Errorf(codes.Unauthenticated, "missing x-hmac-key-id metadata")
	ErrMissingMetadata      = status.Errorf(codes.Unauthenticated, "missing hmac metadata")
)

func init() {
	lvl, _ := os.LookupEnv("GO_GRPC_HMAC_LOG")
	if strings.ToLower(lvl) == "true" {
		EnableLogging()
	}
}

// EnableLogging for this module.
func EnableLogging() {
	logger.SetOutput(os.Stderr)
}

// DisableLogging for this module.
func DisableLogging() {
	logger.SetOutput(io.Discard)
}

// NewMessage returns a string representation of the request and method.
func NewMessage(req interface{}, method string) (string, error) {
	buf := new(bytes.Buffer)
	if req == nil {
		logger.Println("warning: no request, using only method name as message")
		return "method=" + method, nil
	}
	reqBuf := new(bytes.Buffer)
	if err := json.NewEncoder(reqBuf).Encode(req); err != nil {
		if strings.Contains(err.Error(), "has no exported fields") {
			logger.Println("warning: no exported fields in request, using only method name as message")
			goto ADDMETHOD
		}
		return "", fmt.Errorf("failed to encode request: %w", err)
	}
	reqBuf.Truncate(reqBuf.Len() - 1) // remove trailing newline
	if reqBuf.Len() > emptyBracketLength {
		buf.WriteString("request=")
		buf.Write(reqBuf.Bytes())
		buf.WriteString(";")
	}
ADDMETHOD:
	buf.WriteString("method=" + method)
	return buf.String(), nil
}

// Bytes generate a HMAC signature and return it as a base64 encoded []byte.
func Bytes(secretKey string, message string) []byte {
	logger.Printf("generating signature for message %q", message)
	mac := hmac.New(sha512.New512_256, []byte(secretKey))
	mac.Write([]byte(message))
	in := mac.Sum(nil)
	data := make([]byte, base64.StdEncoding.EncodedLen(len(in)))
	base64.StdEncoding.Encode(data, in)
	return data
}

// String generates a HMAC signature and returns it as a base64 encoded string.
func String(secretKey string, message string) string {
	return string(Bytes(secretKey, message))
}

func authForSecrets(getSecret GetSecret) func(ctx context.Context, message string) error {
	return func(ctx context.Context, message string) error {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return ErrMissingMetadata
		}
		hmacSign := getFirst(md, "x-hmac-signature")
		if hmacSign == "" {
			return ErrMissingHmac
		}
		hmacKeyID := getFirst(md, "x-hmac-key-id")
		if hmacKeyID == "" {
			return ErrMissingHmacKeyID
		}
		secretKey, err := getSecret(ctx, hmacKeyID)
		if err != nil {
			log.Printf("internal error getting secret for keyID %s: %q", hmacKeyID, err)
			return status.Error(codes.Internal, err.Error())
		}
		if secretKey == "" {
			logger.Printf("no secret found for keyID %s", hmacKeyID)
			return ErrInvalidHmacKeyID
		}
		if !hmac.Equal([]byte(hmacSign), Bytes(secretKey, message)) {
			return ErrInvalidHmacSignature
		}
		return nil
	}
}

func getFirst(md metadata.MD, key string) string {
	if len(md[key]) > 0 {
		return md[key][0]
	}
	return ""
}
