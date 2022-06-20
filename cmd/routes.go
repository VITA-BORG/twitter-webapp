package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/api/user", app.userAPI)
	mux.HandleFunc("/user", app.user)
	mux.HandleFunc("/", app.home)

	return mux
}
