package app

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"net/http"

	"hw3/internal"
)

func main() {

	var err error

	internal.Db, err = sql.Open("sqlite3", "../UserDatabase.db")
	defer internal.Db.Close()

	if err != nil {
		panic("Couldn't open database connection")
	}

	internal.InitAuth()

	fmt.Printf("Starting server on 8080\n")
	http.ListenAndServe("localhost:8080", router())
}

func router() http.Handler {
	r := chi.NewRouter()
	// Protected service routes
	r.Group(func(r chi.Router) {
		r.Use(internal.Verifier(internal.TokenAuth))
		r.Use(internal.UserAuthenticator)
		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["username"])))
		})
	})
	//Protected confirmation routes
	r.Group(func(r chi.Router) {
		r.Use(internal.Verifier(internal.TokenAuth))
		r.Use(internal.VerificationAuthenticator)
		r.Post("/api/confirm-register", internal.ConfirmRegistration)
		r.Post("/api/confirm-reset", internal.ConfirmResetPassword)
	})
	// Public routes
	r.Group(func(r chi.Router) {
		r.Post("/api/login", internal.Login)
		r.Post("/api/register", internal.Register)
		r.Post("/api/reset-password", internal.ResetPassword)
	})
	return r
}
