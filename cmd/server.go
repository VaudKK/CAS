package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	server *http.Server
	config *Config
}

func NewAPIServer(config *Config) *APIServer {
	router := mux.NewRouter()
	//subRouter := router.PathPrefix("/api/v1").Subrouter()

	srv := &http.Server{
		Addr:     config.addr,
		ErrorLog: config.errorLog,
		Handler:  router,
	}

	return &APIServer{
		server: srv,
		config: config,
	}
}

func (srv *APIServer) Run() {
	srv.config.infoLog.Printf("Starting server on %s", srv.config.addr)
	err := srv.server.ListenAndServe()
	srv.config.errorLog.Fatal(err)
}
