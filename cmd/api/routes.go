package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() http.Handler {
	mx := mux.NewRouter()
	subRouter := mx.PathPrefix("/api/v1").Subrouter()

	// contributions
	subRouter.Handle("/contributions/import", app.requiresAuthenticatedUser(app.upload)).Methods("POST")
	subRouter.Handle("/contributions", app.requiresAuthenticatedUser(app.addContribution)).Methods("POST")
	subRouter.Handle("/contributions", app.requiresAuthenticatedUser(app.getContributions)).Methods("GET")
	subRouter.Handle("/contributions/search", app.requiresAuthenticatedUser(app.search)).Methods("GET")
	subRouter.Handle("/contributions/stats", app.requiresAuthenticatedUser(app.getMonthlyStats)).Methods("GET")
	subRouter.Handle("/contributions/variance", app.requiresAuthenticatedUser(app.getStatisticalVariance)).Methods("GET")
	subRouter.Handle("/contributions/categories/all", app.requiresAuthenticatedUser(app.getCategories)).Methods("GET")
	subRouter.Handle("/contributions/{id}", app.requiresAuthenticatedUser(app.updateContribution)).Methods("PUT")
	subRouter.Handle("/contributions/summary", app.requiresAuthenticatedUser(app.getSummary)).Methods("GET")

	// user
	subRouter.HandleFunc("/auth/signup", app.createUser).Methods("POST")
	subRouter.HandleFunc("/auth/login", app.issueToken).Methods("POST")
	subRouter.HandleFunc("/auth/otp/send", app.sendOtp).Methods("GET")
	subRouter.HandleFunc("/auth/otp/verify", app.verifyOtp).Methods("GET")
	subRouter.HandleFunc("/auth/reset", app.resetPassword).Methods("GET")
	subRouter.HandleFunc("/auth/update", app.changePassword).Methods("POST")

	return app.recoverPanic(app.authenticate(subRouter))
}
