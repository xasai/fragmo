package main

import (
	"net"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	cfg "github.com/xasai/fragmo/config"
	"github.com/xasai/fragmo/rpc"
	"google.golang.org/grpc"
)

type StorageServerHandler struct {
	rpc.UnimplementedStorageServiceServer
}

func main() {
	handler := &StorageServerHandler{}

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", cfg.StorageServerAddr)

	if err != nil {
		log.Error().Err(err)
	}
	rpc.RegisterStorageServiceServer(grpcServer, handler)
	log.Info().Msgf("Starting grpc storage server on %s", cfg.StorageServerAddr)
	log.Fatal().Err(grpcServer.Serve(listener)).Send()
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
