package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/user", app.userAPI)
	mux.HandleFunc("/user", app.user)

	return mux
}
