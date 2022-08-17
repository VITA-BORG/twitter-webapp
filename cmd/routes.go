package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/schools", dynamic.ThenFunc(app.schoolAddGet))
	router.Handler(http.MethodPost, "/schools", dynamic.ThenFunc(app.schoolAddPost))
	router.Handler(http.MethodGet, "/users", dynamic.ThenFunc(app.users))
	router.Handler(http.MethodGet, "/users/view/:id", dynamic.ThenFunc(app.userView))
	router.Handler(http.MethodPost, "/users/view/:id", dynamic.ThenFunc(app.userViewPost))
	router.Handler(http.MethodGet, "/users/add", dynamic.ThenFunc(app.userAddGet))
	router.Handler(http.MethodPost, "/users/add", dynamic.ThenFunc(app.userAddPost))

	//creates a middleware chain
	standard := alice.New(app.recoverPanic, app.logRequest, securityHeaders)

	return standard.Then(router)
}
