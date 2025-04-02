package main

import (
	"flag"
	"log"
	"os"
)

type Config struct {
	dbUrl    string
	addr     string
	infoLog  *log.Logger
	errorLog *log.Logger
}

func main() {

	cfg := new(Config)

	flag.StringVar(&cfg.addr, "addr", ":8080", "Server port")
	flag.StringVar(&cfg.dbUrl, "dbUrl", "postgres://postgres:root@locahost:5432/casdb", "Database Url postgres://{user}:{password}@{hostname}:{port}/{database-name}")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	cfg.infoLog = infoLog
	cfg.errorLog = errorLog

	server := NewAPIServer(cfg)
	server.Run()
}
