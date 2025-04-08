package main

import (
	"database/sql"
	"net/http"
)

type APIServer struct {
	server      *http.Server
	application *application
}

func NewAPIServer(application *application) *APIServer {

	srv := &http.Server{
		Addr:     application.addr,
		ErrorLog: application.errorLog,
		Handler:  application.routes(),
	}

	return &APIServer{
		server:      srv,
		application: application,
	}
}

func (srv *APIServer) Run() {
	srv.application.infoLog.Printf("Starting server on %s", srv.application.addr)

	db, err := openDB(srv.application.dbUrl)

	if err != nil {
		srv.application.errorLog.Fatal(err)
	}

	defer db.Close()

	err = srv.server.ListenAndServe()
	srv.application.errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}
