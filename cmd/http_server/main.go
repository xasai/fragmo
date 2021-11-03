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
	storage   rpc.StorageServiceClient
	recorder  = NewRecorder(cfg.RecordsFilename)
	templates *template.Template
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	defer recorder.Close()

	templates = template.Must(template.ParseGlob(cfg.TemplatesPath + "/*"))

	log.Info().Msgf("connecting to storage_server on %s", cfg.StorageServerAddr)
	storageConnection, err := grpc.Dial(
		cfg.StorageServerAddr,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Millisecond*1000),
	)

	if err != nil {
		log.Fatal().Err(err).Msgf("connecting to storage_server %s", cfg.StorageServerAddr)
	}
	defer storageConnection.Close()
	log.Info().Msgf("connected to storage_server on %s", cfg.StorageServerAddr)

	storage = rpc.NewStorageServiceClient(storageConnection)

	r := mux.NewRouter()
	r.HandleFunc("/", listHandler)
	r.HandleFunc("/list", listHandler)
	r.HandleFunc("/download", downloadHandler)
	r.HandleFunc("/upload", uploadHandler)
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
