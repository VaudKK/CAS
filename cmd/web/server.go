package main

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	server *http.Server
	application *Application
}

func NewAPIServer(application *Application) *APIServer {
	router := mux.NewRouter()
	//subRouter := router.PathPrefix("/api/v1").Subrouter()

	srv := &http.Server{
		Addr:     application.addr,
		ErrorLog: application.errorLog,
		Handler:  router,
	}

	return &APIServer{
		server: srv,
		application: application,
	}
}

func (srv *APIServer) Run() {
	srv.application.infoLog.Printf("Starting server on %s", srv.application.addr)

	db,err := openDB(srv.application.dbUrl)

	if err != nil {
		srv.application.errorLog.Fatal(err)
	}

	defer db.Close()

	err = srv.server.ListenAndServe()
	srv.application.errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB,error){
	db,err := sql.Open("postgres",dsn)

	if err != nil  {
		return nil,err
	}

	if err = db.Ping(); err != nil {
		return nil,err
	}

	return db,err
}
