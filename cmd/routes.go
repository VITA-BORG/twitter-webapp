package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {

	router := httprouter.New()

	//todo: tell router how to handle certain errors

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.Handler(http.MethodGet, "/static/", http.StripPrefix("/static", fileServer))

	router.HandlerFunc(http.MethodGet, "/api/users", app.userAPI)
	router.HandlerFunc(http.MethodGet, "/users", app.user)
	router.HandlerFunc(http.MethodGet, "/", app.home)

	//creates a middleware chain
	standard := alice.New(app.recoverPanic, app.logRequest, securityHeaders)

	return standard.Then(router)
}
