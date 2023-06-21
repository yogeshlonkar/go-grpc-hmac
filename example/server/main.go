package main

import (
	"fmt"
	graceful "github.com/yogeshlonkar/go-shutdown-graceful"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"go-grpc-hmac/example/pb"
)

const shutdownDelay = 5 * time.Second

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, FormatTimestamp: formatter})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	s := &server{}
	s.setup()
	pb.RegisterUserServiceServer(s.server, &Servicer{})
	go s.start()
	go s.handleShutdown()
	time.Sleep(1 * time.Second)
	log.Info().Msgf("Listening and serving server on %s", s.addr)
	if err := graceful.HandleSignals(shutdownDelay); err != nil {
		log.Error().Err(err).Msg("failed to handle signals")
	}
	log.Info().Msg("shutdown complete")
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
	return fmt.Sprintf("\u001B[32m[example-server] \x1b[90m%v\x1b[0m", i)
}
