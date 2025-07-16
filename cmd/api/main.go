package main

import (
	"database/sql"
	"flag"
	"net/http"

	"github.com/VaudKK/CAS/utils"

	excel_exports "github.com/VaudKK/CAS/pkg/exports/excel"
	pdf_exports "github.com/VaudKK/CAS/pkg/exports/pdf"
	"github.com/VaudKK/CAS/pkg/mailer"
	"github.com/VaudKK/CAS/pkg/models/postgres"
	_ "github.com/lib/pq"
)

type config struct {
	env   string
	dbUrl string
	port  string
	smtp  struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
	configuration *config
	fundsModel    *postgres.FundsModel
	userModel     *postgres.UserModel
	otpModel      *postgres.OtpModel
	mailer        mailer.Mailer
}

func main() {

	cfg := new(config)

	flag.StringVar(&cfg.port, "addr", ":8080", "Server port")
	flag.StringVar(&cfg.dbUrl, "dbUrl", "postgres://postgres:root@localhost:5432/casdb?sslmode=disable", "Database Url postgres://{user}:{password}@{hostname}:{port}/{database-name}")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|production)")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "live.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "api", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "ddc03a96f5541dbd293c4e6a7371212a", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "KCSDA <noreply@kcsda.or.ke>", "SMTP sender")

	// flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	// flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	// flag.StringVar(&cfg.smtp.username, "smtp-username", "5f7aeca5180126", "SMTP username")
	// flag.StringVar(&cfg.smtp.password, "smtp-password", "dfa110cb72fb74", "SMTP password")
	// flag.StringVar(&cfg.smtp.sender, "smtp-sender", "CAS <no-reply@churchaccountingsystem>", "SMTP sender")

	flag.Parse()

	application := &application{
		configuration: cfg,
		mailer:        mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
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
		ExcelExporter: &excel_exports.ExcelExport{},
		PdfExporter:   &pdf_exports.PdfExport{
			Logger: utils.GetLoggerInstance(),
		},
		Logger: 	utils.GetLoggerInstance(),
	}

	application.userModel = &postgres.UserModel{
		DB: db,
		Mailer: &application.mailer,
	}

	application.otpModel = &postgres.OtpModel{
		DB:     db,
		Mailer: &application.mailer,
		User:   application.userModel,
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
