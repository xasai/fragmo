package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	cfg "github.com/xasai/fragmo/config"
	"github.com/xasai/fragmo/rpc"
)

func (h *StorageServerHandler) chunkName(name string, idx int) string {
	return cfg.StoragePath + fmt.Sprintf("%s_%d", name, idx)
}

func (h *StorageServerHandler) Download(req *rpc.DownloadReq, stream rpc.StorageService_DownloadServer) (err error) {
	const op = "StorageServerHandler.Download"
	log.Info().Msg(op)

	res := &rpc.File{
		Data: make([]byte, cfg.FragmentSize),
	}

	fileIdx := 0

	//Read all files until some file chunk not exists
	for {
		fileIdx++
		filename := h.chunkName(req.Filename, fileIdx)
		f, err := os.OpenFile(filename, os.O_RDONLY, 0)
		if err != nil {
			break
		}

		_, err = f.Read(res.Data)
		if err != nil {
			break
		}

		err = stream.Send(res)
		if err != nil {
			break
		}
	}

	if err != nil {
		log.Error().Err(err).Send()
	}

	return err
}
