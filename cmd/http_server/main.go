package main

import (
	"net/http"
	"os"
	"text/template"
	"time"

	"google.golang.org/grpc"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	cfg "github.com/xasai/fragmo/config"
	"github.com/xasai/fragmo/rpc"
)

var (
	recorder  FileRecorder = NewRecorder(cfg.RecordsFilename)
	templates *template.Template
	storage   rpc.StorageServiceClient
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	defer recorder.Close()

	log.Info().Msgf("connecting to storage_server on %s", cfg.StorageServerAddr)
	storageConnection, err := grpc.Dial(
		cfg.StorageServerAddr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Millisecond*1000),
	)
	if err != nil {
		log.Fatal().Err(err).Msgf(cfg.StorageServerAddr)
	}
	defer storageConnection.Close()
	storage = rpc.NewStorageServiceClient(storageConnection)

	templates = template.Must(template.ParseGlob(cfg.TemplatesPath + "*"))

	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/upload", uploadHandler)
	r.HandleFunc("/download", downloadHandler).Queries("file", "{file}")
	r.Use(LoggerMware)

	http.Handle("/", r)

	srv := &http.Server{
		Addr:         cfg.HttpServerAddr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	log.Info().Msgf("%s serve on http://%s", os.Args[0], srv.Addr)
	log.Error().Err(srv.ListenAndServe())
}
