package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/VaudKK/CAS/pkg/data"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.writeJSONError(w, http.StatusInternalServerError, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")

		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			app.writeJSONError(w, http.StatusUnauthorized, errors.New("invalid or missing authentication token"))
			return
		}

		token := headerParts[1]

		userId, err := data.VerifyToken(token)

		if err != nil {
			app.writeJSONError(w, http.StatusUnauthorized, err)
			return
		}

		user, err := app.userModel.GetUserID(userId)

		if err != nil {
			switch {
			case errors.Is(err, data.ErrorNoRecords):
				app.writeUnauthorizedJSON(w, r)
			default:
				app.writeJSONError(w, http.StatusInternalServerError, err)
			}

			return
		}

		r = app.contextSetUser(r, user)

		next.ServeHTTP(w, r)

	})
}

func (app *application) requiresAuthenticatedUser(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the contextGetUser() helper to retrieve the user
		// information from the request context.
		user := app.contextGetUser(r)
		// If the user is anonymous, then call the authenticationRequiredResponse() to
		// inform the client that they should authenticate before trying again.
		if data.IsAnonymous(user) {
			app.writeJSON(w, http.StatusUnauthorized, envelope{"errorMessage": "unauthorized"})
			return
		}
		// If the user is not activated,
		// inform them that they need to verify their account.
		if !user.Verified {
			app.writeJSON(w, http.StatusForbidden, envelope{"errorMessage": "unverified user"})
			return
		}
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}
