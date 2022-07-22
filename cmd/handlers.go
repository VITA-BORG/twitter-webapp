package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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

	err := r.ParseForm()
	if err != nil {
		app.serverError(w, err)
		return
	}

	toScrape := &simplifiedUser{}

	toScrape.Username = r.PostForm.Get("handle")
	toScrape.IsParticipant = true
	toScrape.IsSchool = false

	if r.PostForm.Get("follows") == "true" {
		toScrape.ScrapeConnections = true
	}

	if r.PostForm.Get("content") == "true" {
		toScrape.ScrapeContent = true
	}

	schoolName := r.PostForm.Get("school")

	dateForm := r.PostForm.Get("start-date")

	//validate form input
	fieldErrors := make(map[string]string)

	if strings.TrimSpace(toScrape.Username) == "" {
		fieldErrors["handle"] = "Handle is required"
	}

	if strings.TrimSpace(schoolName) == "" {
		fieldErrors["school"] = "School is required"
	}

	if strings.TrimSpace(dateForm) == "" {
		fieldErrors["start-date"] = "Start date is required"
	} else if dateForm == "yyyy-mm-dd" {
		fieldErrors["start-date"] = "Start date is required"
	}

	if len(fieldErrors) > 0 {
		fmt.Fprint(w, fieldErrors)
		return
	}

	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(dateForm))

	if err != nil {
		app.serverError(w, err)
		return
	}

	toScrape.StartDate = startDate

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

	r.ParseForm()

	toScrape := &simplifiedUser{}
	toInsert := &simplifiedSchool{}

	toInsert.Name = strings.TrimSpace(r.PostForm.Get("name"))
	toInsert.City = strings.ToLower(strings.TrimSpace(r.PostForm.Get("city")))
	toInsert.State = strings.ToLower(strings.TrimSpace(r.PostForm.Get("state")))
	toInsert.Country = strings.ToLower(strings.TrimSpace(r.PostForm.Get("country")))
	toInsert.TwitterHandle = strings.ToLower(strings.TrimSpace(r.PostForm.Get("handle")))
	if r.PostForm.Get("top-rated") == "true" {
		toInsert.TopRated = true
	}
	if r.PostForm.Get("public") == "true" {
		toInsert.Public = true
	}

	//validate form input
	fieldErrors := make(map[string]string)

	if toInsert.Name == "" {
		fieldErrors["name"] = "Name is required"
	}

	if toInsert.City == "" {
		fieldErrors["city"] = "City is required"
	}

	if toInsert.State == "" {
		fieldErrors["state"] = "State is required"
	}

	if toInsert.Country == "" {
		fieldErrors["country"] = "Country is required"
	}

	if toInsert.TwitterHandle == "" {
		fieldErrors["handle"] = "Handle is required"
	}

	if len(fieldErrors) > 0 {
		fmt.Fprint(w, fieldErrors)
		return
	}

	toScrape.IsSchool = true
	toScrape.SchoolInfo = toInsert
	toScrape.Username = toInsert.TwitterHandle

	app.profileChan <- toScrape

	fmt.Fprintf(w, r.FormValue("top-rated"))
}
