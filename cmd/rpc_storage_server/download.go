package main

import (
	"github.com/rs/zerolog/log"
	"github.com/xasai/fragmo/rpc"
)

func (h *StorageServerHandler) Download(req *rpc.DownloadReq, stream rpc.StorageService_DownloadServer) error {
	const op = "StorageServerHandler.Download"
	log.Info().Msg(op)

	return nil
}
