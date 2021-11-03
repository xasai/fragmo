package main

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/rs/zerolog/log"
	cfg "github.com/xasai/fragmo/config"
	"github.com/xasai/fragmo/rpc"
)

//downloadHandler takes filename as get 'file' parameter.
//If such file exists in records, try to download it from storage gRPC server.
//Otherwise return 404 not found.
func downloadHandler(w http.ResponseWriter, req *http.Request) {
	const op = "downloadHandler"

	filename, err := url.QueryUnescape(req.URL.Query().Get("file"))

	//Check if file exists
	if err != nil || !recorder.Exist(filename) {
		log.Error().Err(err).Msgf("file not found %s", filename)
		w.WriteHeader(http.StatusNotFound)
		err := templates.ExecuteTemplate(w, "404.html", filename)
		if err != nil {
			log.Error().Err(err).Msg(op + ".ExecuteTemplate")
		}
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", req.Header.Get("Content-Type"))
	w.Header().Set("Content-Lenght", strconv.FormatInt(recorder.GetSize(filename), 10))

	reqRPC := &rpc.DownloadReq{Filename: filename}

	//Creating stream with storage rpc
	s, err := storage.Download(context.Background(), reqRPC)
	if err != nil {
		http.Error(w, "connection refused", http.StatusNotFound)
		log.Error().Err(err).Msg(op)
		return
	}

	f := &rpc.File{
		Data: []byte{},
	}

	//If stream successfully created we can redirect
	//client to index page and start download
	http.RedirectHandler("/", http.StatusMovedPermanently)

	size := int64(0)

	for idx := 0; ; idx++ {
		//read file by file and write it to client writer
		//until err or size > file.Size
		f, err = s.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Error().Err(err).Msg(op)
			return
		}

		log.Info().Str("file", filename).Int("size", len(f.Data)).Int("index", idx).Msgf("received")

		size += int64(len(f.Data))
		if size >= recorder.GetSize(filename) {
			break
		}

		_, err = w.Write(f.Data)
		if err != nil {
			log.Error().Err(err).Msg(op)
			return
		}
	}

	//last chunk size
	size = recorder.GetSize(filename) % cfg.FragmentSize
	f.Data = f.Data[:size]
	_, err = w.Write(f.Data)
	if err != nil {
		log.Error().Err(err).Msg(op)
	}

	err = s.CloseSend()
	if err != nil {
		log.Error().Err(err).Msg(op)
	}
}
