package main

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func LoggerMware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Info().Str("URI", req.RequestURI).Msg(req.Method)
		next.ServeHTTP(w, req)
	})
}
