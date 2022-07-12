package main

import (
	"fmt"
	"net/http"
)

func (app *application) users(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "User Homepage")
}

//home is a handler for the root endpoint.  It shows a simple list of users in the database.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	Users := app.getAllUsernames()
	data := &templateData{
		Users:           Users,
		ProfileStatus:   app.profileStatus,
		FollowerStatus:  app.followStatus,
		FollowingStatus: app.followingStatus,
	}

	app.renderTemplate(w, http.StatusOK, "dashboard.html", data)

	fmt.Fprintf(w, "Homepage")
}

func (app *application) userAddGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "User Add Form")
}

func (app *application) userAddPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "User Added")
}
