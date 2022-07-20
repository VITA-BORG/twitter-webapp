package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
)

func (app *application) userView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	username := params.ByName("username")

	user, err := app.getUserByHandle(username)
	if err != nil {
		app.notFound(w)
		return
	}

	data := &templateData{
		CurrentUser: *user,
	}

	app.populateWorkerStatus(data)

	app.renderTemplate(w, http.StatusOK, "userView.html", data)
}

func (app *application) users(w http.ResponseWriter, r *http.Request) {
	Users, err := models.GetAllParticipants(app.connection)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := &templateData{
		Users:           Users,
		ProfileStatus:   app.profileStatus,
		FollowerStatus:  app.followStatus,
		FollowingStatus: app.followingStatus,
		NumberOfUsers:   len(Users),
	}

	app.renderTemplate(w, http.StatusOK, "users.html", data)

	fmt.Fprintf(w, "Users")
}

//home is a handler for the root endpoint.  It shows a simple list of users in the database.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	Users := app.getAllUsernames()
	data := &templateData{
		ProfileStatus:   app.profileStatus,
		FollowerStatus:  app.followStatus,
		FollowingStatus: app.followingStatus,
		NumberOfUsers:   len(Users),
	}

	app.renderTemplate(w, http.StatusOK, "dashboard.html", data)

	fmt.Fprintf(w, "Homepage")
}

func (app *application) userAddGet(w http.ResponseWriter, r *http.Request) {
	data := &templateData{}
	app.renderTemplate(w, http.StatusOK, "userAdd.html", data)
	fmt.Fprintf(w, "User Add Form")
}

func (app *application) userAddPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "User Added")
}

func (app *application) schoolAddGet(w http.ResponseWriter, r *http.Request) {
	data := &templateData{}
	app.renderTemplate(w, http.StatusOK, "schoolAdd.html", data)
}

func (app *application) schoolAddPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "School Added")
}
