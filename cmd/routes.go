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

	//unprotected routes
	dynamic := alice.New(app.sessionManager.LoadAndSave)
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))

	//protected routes
	protected := dynamic.Append(app.requireAuthentication)
	router.Handler(http.MethodGet, "/schools", protected.ThenFunc(app.schoolAddGet))
	router.Handler(http.MethodPost, "/schools", protected.ThenFunc(app.schoolAddPost))
	router.Handler(http.MethodGet, "/users", protected.ThenFunc(app.users))
	router.Handler(http.MethodGet, "/users/view/:id", protected.ThenFunc(app.userView))
	router.Handler(http.MethodPost, "/users/view/:id", protected.ThenFunc(app.userViewPost))
	router.Handler(http.MethodGet, "/users/add", protected.ThenFunc(app.userAddGet))
	router.Handler(http.MethodPost, "/users/add", protected.ThenFunc(app.userAddPost))
	router.Handler(http.MethodGet, "/user/signup", protected.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", protected.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))

	//creates a middleware chain
	standard := alice.New(app.recoverPanic, app.logRequest, securityHeaders)

	return standard.Then(router)
}
