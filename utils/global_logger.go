package utils

import (
	"log"
	"os"
	"sync"
)

var logger *CLogger
var once sync.Once

type CLogger struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func GetLoggerInstance() *CLogger {
	// Use sync.Once to ensure that the logger is initialized only once
	once.Do(func() {
		logger = &CLogger{
			InfoLog:  log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
			ErrorLog: log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		}
	})

	return logger
}