package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() http.Handler {
	mx := mux.NewRouter()
	subRouter := mx.PathPrefix("/api/v1").Subrouter()

	subRouter.HandleFunc("/contributions/import",app.upload)
	subRouter.HandleFunc("/contributions", app.getContributions).Methods("GET")

	return subRouter
}