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
	numUsers, err := models.GetUserCount(app.connection)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := &templateData{
		Users:           Users,
		ProfileStatus:   app.profileStatus,
		FollowerStatus:  app.followStatus,
		FollowingStatus: app.followingStatus,
		NumberOfUsers:   numUsers,
		NumParticipants: len(Users),
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
	schools, err := models.GetAllSchools(app.connection)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := &templateData{
		Schools: schools,
	}

	app.populateWorkerStatus(data)

	app.renderTemplate(w, http.StatusOK, "userAdd.html", data)
	fmt.Fprintf(w, "User Add Form")
}

func (app *application) userAddPost(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	toScrape := &simplifiedUser{}

	toScrape.Username = r.FormValue("handle")
	toScrape.IsParticipant = true
	toScrape.IsSchool = false

	if r.FormValue("follows") == "true" {
		toScrape.ScrapeConnections = true
	}

	if r.FormValue("content") == "true" {
		toScrape.ScrapeContent = true
	}

	schoolName := r.FormValue("school")
	school, err := models.GetSchoolByName(app.connection, schoolName)
	if err != nil {
		app.serverError(w, err)
		return
	}

	toScrape.ParticipantSchoolID = school.ID

	app.profileChan <- toScrape

	fmt.Fprintf(w, "User Added")
}

func (app *application) schoolAddGet(w http.ResponseWriter, r *http.Request) {
	data := &templateData{}
	app.renderTemplate(w, http.StatusOK, "schoolAdd.html", data)
	fmt.Fprintf(w, "School Add Form")
}

func (app *application) schoolAddPost(w http.ResponseWriter, r *http.Request) {

	//TODO check form values.  Make sure they are within limits

	r.ParseForm()

	toScrape := &simplifiedUser{}
	toInsert := &simplifiedSchool{}

	toInsert.Name = r.FormValue("name")
	toInsert.City = r.FormValue("city")
	toInsert.State = r.FormValue("state")
	toInsert.Country = r.FormValue("country")
	toInsert.TwitterHandle = r.FormValue("handle")
	if r.FormValue("top-rated") == "true" {
		toInsert.TopRated = true
	}
	if r.FormValue("public") == "true" {
		toInsert.Public = true
	}

	toScrape.IsSchool = true
	toScrape.SchoolInfo = toInsert
	toScrape.Username = toInsert.TwitterHandle

	app.profileChan <- toScrape

	fmt.Fprintf(w, r.FormValue("top-rated"))
}
