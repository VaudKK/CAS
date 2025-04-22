package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() http.Handler {
	mx := mux.NewRouter()
	subRouter := mx.PathPrefix("/api/v1").Subrouter()

	// contributions
	subRouter.HandleFunc("/contributions/import",app.upload)
	subRouter.HandleFunc("/contributions", app.getContributions).Methods("GET")
	subRouter.HandleFunc("/contributions/search", app.search).Methods("GET")

	// sign up
	subRouter.HandleFunc("/auth/signup", app.createUser).Methods("POST")

	return subRouter
}