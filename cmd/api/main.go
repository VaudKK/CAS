package main

import (
	"database/sql"
	"flag"
	"net/http"

	"github.com/VaudKK/CAS/utils"

	"github.com/VaudKK/CAS/pkg/models/postgres"
	_ "github.com/lib/pq"
)

type config struct {
	env   string
	dbUrl string
	port  string
}

type application struct {
	configuration *config
	fundsModel    *postgres.FundsModel
}

func main() {

	cfg := new(config)

	flag.StringVar(&cfg.port, "addr", ":8080", "Server port")
	flag.StringVar(&cfg.dbUrl, "dbUrl", "postgres://postgres:root@localhost:5432/casdb?sslmode=disable", "Database Url postgres://{user}:{password}@{hostname}:{port}/{database-name}")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|production)")
	flag.Parse()

	application := &application{
		configuration: cfg,
	}

	run(application)
}

func run(application *application) {

	server := &http.Server{
		Addr:     application.configuration.port,
		ErrorLog: utils.GetLoggerInstance().ErrorLog,
		Handler:  application.routes(),
	}

	utils.GetLoggerInstance().InfoLog.Printf("Starting server on %s", application.configuration.port)

	db, err := openDB(application.configuration.dbUrl)

	if err != nil {
		utils.GetLoggerInstance().ErrorLog.Fatal(err)
	}

	defer db.Close()

	application.fundsModel = &postgres.FundsModel{
		DB: db,
	}

	err = server.ListenAndServe()
	utils.GetLoggerInstance().ErrorLog.Fatal(err)
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
