package main

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func indexHandler(w http.ResponseWriter, req *http.Request) {
	const op = "indexHandler"

	err := templates.ExecuteTemplate(w, "index.html", recorder.GetRecords())
	if err != nil {
		log.Error().Err(err).Msg(op + ".ExecuteTemplate")
	}
}
