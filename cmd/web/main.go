package main

import (
	"flag"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Application struct {
	dbUrl    string
	addr     string
	infoLog  *log.Logger
	errorLog *log.Logger
}

func main() {

	application := new(Application)

	flag.StringVar(&application.addr, "addr", ":8080", "Server port")
	flag.StringVar(&application.dbUrl, "dbUrl", "postgres://postgres:root@localhost:5432/casdb?sslmode=disable", "Database Url postgres://{user}:{password}@{hostname}:{port}/{database-name}")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	application.infoLog = infoLog
	application.errorLog = errorLog

	server := NewAPIServer(application)
	server.Run()
}
