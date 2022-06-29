package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/api/users", app.userAPI)
	mux.HandleFunc("/users", app.user)
	mux.HandleFunc("/", app.home)

	//creates a middleware chain
	standard := alice.New(app.recoverPanic, app.logRequest, securityHeaders)

	return standard.Then(mux)
}
