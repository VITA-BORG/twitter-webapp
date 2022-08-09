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
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/schools", app.schoolAddGet)
	router.HandlerFunc(http.MethodPost, "/schools", app.schoolAddPost)
	router.HandlerFunc(http.MethodGet, "/users", app.users)
	router.HandlerFunc(http.MethodGet, "/users/view/:id", app.userView)
	router.HandlerFunc(http.MethodPost, "/users/view/:id", app.userViewPost)
	router.HandlerFunc(http.MethodGet, "/users/add", app.userAddGet)
	router.HandlerFunc(http.MethodPost, "/users/add", app.userAddPost)

	//creates a middleware chain
	standard := alice.New(app.recoverPanic, app.logRequest, securityHeaders)

	return standard.Then(router)
}
