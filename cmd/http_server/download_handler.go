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

func downloadHandler(w http.ResponseWriter, req *http.Request) {
	const op = "downloadHandler"

	filename, err := url.QueryUnescape(req.URL.Query().Get("file"))

	if err != nil || !recorder.Exist(filename) {
		log.Error().Err(err).Msgf("file not found %s", filename)
		w.WriteHeader(http.StatusNotFound)
		err := templates.ExecuteTemplate(w, "404.html", filename)
		if err != nil {
			log.Error().Err(err).Msg(op + ".ExecuteTemplate")
		}
		return
	}

	reqRPC := &rpc.DownloadReq{
		Filename: filename,
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", req.Header.Get("Content-Type"))
	w.Header().Set("Content-Lenght", strconv.FormatInt(recorder.GetSize(filename), 10))

	s, err := storage.Download(context.Background(), reqRPC)
	if err != nil {
		http.Error(w, "connection refused", http.StatusNotFound)
		log.Error().Err(err).Msg(op)
		return
	}

	f := &rpc.File{
		Data: []byte{},
	}

	http.RedirectHandler("/", http.StatusMovedPermanently)

	size := int64(0)
	for idx := 0; ; idx++ {
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
		return
	}

	err = s.CloseSend()
	if err != nil {
		log.Error().Err(err).Msg(op)
	}
}
