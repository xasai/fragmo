package main

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog/log"
	cfg "github.com/xasai/fragmo/config"
	"github.com/xasai/fragmo/rpc"
)

func (h *StorageServerHandler) Upload(stream rpc.StorageService_UploadServer) error {
	const op = "StorageServerHandler.Upload"
	log.Info().Msg(op)

	var files = []string{}

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return h.onErrorUpload(err, op, files)
		}

		log.Info().
			Str("filename", req.Filename).
			Int32("idx", req.Index).
			Msgf("received %.2f Mb of data", float32(len(req.Data))/(1<<20))

		filename := cfg.StoragePath + fmt.Sprintf("%s_%d", req.Filename, req.Index)

		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, os.FileMode(0666))
		if err != nil {
			return h.onErrorUpload(err, op, files)
		}
		defer f.Close()

		files = append(files, filename)

		_, err = f.Write(req.Data)
		if err != nil {
			return h.onErrorUpload(err, op, files)
		}
	}
	stream.SendAndClose(&rpc.Empty{})
	return nil
}

func (h *StorageServerHandler) onErrorUpload(err error, op string, files []string) error {
	log.Error().Err(err).Msg(op)
	h.removeFiles(files)
	return err
}

func (h *StorageServerHandler) removeFiles(files []string) {
	for _, f := range files {
		os.Remove(f)
		log.Info().Msgf("removing %s ...", f)
	}
}
