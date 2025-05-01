package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() http.Handler {
	mx := mux.NewRouter()
	subRouter := mx.PathPrefix("/api/v1").Subrouter()

	// contributions
	subRouter.Handle("/contributions/import",app.requiredAuthenticatedUser(app.upload)).Methods("POST")
	subRouter.Handle("/contributions", app.requiredAuthenticatedUser(app.addContribution)).Methods("POST")
	subRouter.Handle("/contributions/search", app.requiredAuthenticatedUser(app.search)).Methods("GET")
	subRouter.Handle("/contributions/stats",app.requiredAuthenticatedUser(app.getMonthlyStats)).Methods("GET")
	subRouter.Handle("/contributions/categories/all",app.requiredAuthenticatedUser(app.getCategories)).Methods("GET")
	subRouter.Handle("/contributions/{id}", app.requiredAuthenticatedUser(app.updateContribution)).Methods("PUT")

	// user
	subRouter.HandleFunc("/auth/signup", app.createUser).Methods("POST")
	subRouter.HandleFunc("/auth/login",app.issueToken).Methods("POST")
	subRouter.HandleFunc("/auth/otp/send",app.sendOtp).Methods("GET")
	subRouter.HandleFunc("/auth/otp/verify",app.verifyOtp).Methods("GET")

	return app.recoverPanic(app.authenticate(subRouter))
}