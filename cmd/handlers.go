package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
)

type userViewForm struct {
	Handle      string `schema:"handle"`
	School      string `schema:"school"`
	StartDate   string `schema:"start-date"`
	Follows     bool   `schema:"follows"`
	Content     bool   `schema:"content"`
	Cohort      int    `schema:"cohort"`
	FieldErrors map[string]string
}
type userAddForm struct {
	Handle      string `schema:"handle"`
	School      string `schema:"school"`
	StartDate   string `schema:"start-date"`
	Follows     bool   `schema:"follows"`
	Content     bool   `schema:"content"`
	Cohort      string `schema:"cohort"`
	FieldErrors map[string]string
}

type schoolAddForm struct {
	Name        string `schema:"name"`
	City        string `schema:"city"`
	State       string `schema:"state"`
	Country     string `schema:"country"`
	Handle      string `schema:"handle"`
	TopRated    bool   `schema:"top-rated"`
	Public      bool   `schema:"public"`
	FieldErrors map[string]string
}

func (app *application) userView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	uid, err := strconv.ParseInt(params.ByName("id"), 10, 64)

	if err != nil {
		app.notFound(w)
		return
	}

	user, err := models.GetUserByID(app.connection, uid)
	if err != nil {
		app.notFound(w)
		return
	}

	student, err := models.GetStudentByID(app.connection, user.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	school, err := models.GetSchoolByID(app.connection, student.SchoolID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	schools, err := models.GetAllSchools(app.connection)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form := userViewForm{
		Handle: user.Handle,
		School: school.Name,
		Cohort: student.Cohort,
	}

	//removes user's school from the slice of schools available
	for i, s := range schools {
		if s.ID == school.ID {
			schools[i] = schools[len(schools)-1]
			schools = schools[:len(schools)-1]
		}
	}

	data := &templateData{
		UserViewPage: userViewPage{
			CurrentUser: *user,
			Schools:     schools,
			Form:        form,
		},
	}

	app.populateStatusData(data)

	app.renderTemplate(w, http.StatusOK, "userView.html", data)
}

//userAddPost is a handler for the POST request to the /user/view/:id endpoint.  It validates the form data and, if valid, updates user to the database.
func (app *application) userViewPost(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	uid, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		app.notFound(w)
		return
	}

	form := userViewForm{}

	form.Handle = strings.TrimSpace(r.PostForm.Get("handle"))
	form.School = strings.TrimSpace(r.PostForm.Get("school"))
	form.StartDate = r.PostForm.Get("start-date")
	form.Follows = r.PostForm.Get("follows") == "true"
	form.Content = r.PostForm.Get("content") == "true"

	form.Cohort, err = strconv.Atoi(strings.TrimSpace(r.PostForm.Get("cohort")))

	if err != nil {
		form.Cohort = 0
		form.FieldErrors["cohort"] = "Cohort must be a number"
	}

	if form.Handle == "" {
		form.FieldErrors["handle"] = "Required"
	}
	if form.School == "" {
		form.FieldErrors["school"] = "Required"
	}
	if form.StartDate == "" {
		form.FieldErrors["startDate"] = "Required"
	}

	startDate, err := time.Parse("2006-01-02", form.StartDate)

	if err != nil {
		form.FieldErrors["startDate"] = "Invalid date format"
	}

	if len(form.FieldErrors) > 0 {
		app.infoLog.Println("Errors found in form")
		schools, err := models.GetAllSchools(app.connection)
		if err != nil {
			app.serverError(w, err)
			return
		}
		user, err := models.GetUserByID(app.connection, uid)
		if err != nil {
			app.serverError(w, err)
			return
		}
		data := &templateData{
			UserViewPage: userViewPage{
				CurrentUser: *user,
				Schools:     schools,
				Form:        form,
			},
		}
		app.populateStatusData(data)
		app.renderTemplate(w, http.StatusUnprocessableEntity, "userAdd.html", data)
		return
	}

	user := &models.User{
		ID:     uid,
		Handle: form.Handle,
	}

	//updates the user handle if there has been a change
	err = models.UpdateUserHandle(app.connection, user)
	if err != nil {
		app.errorLog.Println("Error updating user handle:", user)
		app.serverError(w, err)
		return
	}

	schoolID, err := models.GetSchoolIDByName(app.connection, form.School)
	if err != nil {
		app.serverError(w, err)
		return
	}

	toScrape := &simplifiedUser{
		ID:                  uid,
		Username:            form.Handle,
		IsSchool:            false,
		IsParticipant:       true,
		ParticipantCohort:   form.Cohort,
		ParticipantSchoolID: schoolID,
		StartDate:           startDate,
		ScrapeConnections:   form.Follows,
		ScrapeContent:       form.Content,
	}

	app.profileChan <- toScrape

	http.Redirect(w, r, "/users/add", http.StatusSeeOther)

}

func (app *application) users(w http.ResponseWriter, r *http.Request) {
	Users, err := models.GetAllParticipants(app.connection)
	if err != nil {
		app.serverError(w, err)
		return
	}

	usersData := usersPage{
		Participants:    Users,
		NumParticipants: len(Users),
	}
	data := &templateData{
		UsersPage: usersData,
	}
	app.populateStatusData(data)

	app.renderTemplate(w, http.StatusOK, "users.html", data)

	fmt.Fprintf(w, "Users")
}

//home is a handler for the root endpoint.  It shows a simple list of users in the database.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	data := &templateData{}
	app.populateStatusData(data)

	app.renderTemplate(w, http.StatusOK, "dashboard.html", data)

	fmt.Fprintf(w, "Homepage")
}

func (app *application) userAddGet(w http.ResponseWriter, r *http.Request) {
	schools, err := models.GetAllSchools(app.connection)
	if err != nil {
		app.serverError(w, err)
		return
	}

	userAddData := userAddPage{
		Schools: schools,
		Form:    userAddForm{},
	}
	data := &templateData{
		UserAddPage: userAddData,
	}

	app.populateStatusData(data)

	app.renderTemplate(w, http.StatusOK, "userAdd.html", data)
	fmt.Fprintf(w, "User Add Form")
}

func (app *application) userAddPost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.serverError(w, err)
		return
	}

	form := userAddForm{
		Handle:      r.PostForm.Get("handle"),
		School:      r.PostForm.Get("school"),
		StartDate:   r.PostForm.Get("start-date"),
		Cohort:      r.PostForm.Get("cohort"),
		Follows:     r.PostForm.Get("follows") == "true",
		Content:     r.PostForm.Get("content") == "true",
		FieldErrors: map[string]string{},
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

	cohort, err := strconv.Atoi(r.PostForm.Get("cohort"))
	if err != nil {
		form.FieldErrors["cohort"] = "Cohort must be a number"
		return
	}

	if strings.TrimSpace(toScrape.Username) == "" {
		form.FieldErrors["handle"] = "Handle is required"
	}

	if strings.TrimSpace(schoolName) == "" {
		form.FieldErrors["school"] = "School is required"
	}

	if strings.TrimSpace(dateForm) == "" {
		form.FieldErrors["startDate"] = "Start date is required"
	} else if dateForm == "yyyy-mm-dd" {
		form.FieldErrors["startDate"] = "Start date is required"
	}

	//if there are any errors, render the form again with the field errors and repopulated fields
	if len(form.FieldErrors) > 0 {
		app.infoLog.Println("Errors found in form")
		schools, err := models.GetAllSchools(app.connection)
		if err != nil {
			app.serverError(w, err)
			return
		}
		data := &templateData{
			UserAddPage: userAddPage{
				Schools: schools,
				Form:    form,
			},
		}
		app.populateStatusData(data)
		app.renderTemplate(w, http.StatusUnprocessableEntity, "userAdd.html", data)
		return
	}

	schoolID, err := models.GetSchoolIDByName(app.connection, schoolName)
	if err != nil {
		app.serverError(w, err)
		return
	}

	toScrape.ParticipantCohort = cohort
	toScrape.ParticipantSchoolID = schoolID

	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(dateForm))

	if err != nil {
		startDate = time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
		return
	}

	toScrape.StartDate = startDate

	//sends the user to the scraper to scrape the user's profile
	app.profileChan <- toScrape

	http.Redirect(w, r, "/users/add", http.StatusSeeOther)
}

func (app *application) schoolAddGet(w http.ResponseWriter, r *http.Request) {
	data := &templateData{}
	app.populateStatusData(data)
	schools, err := models.GetAllSchools(app.connection)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.SchoolAddPage = schoolAddPage{
		Schools: schools,
		Form:    schoolAddForm{},
	}
	app.renderTemplate(w, http.StatusOK, "schoolAdd.html", data)
	fmt.Fprintf(w, "School Add Form")
}

func (app *application) schoolAddPost(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	form := schoolAddForm{
		Name:        r.PostForm.Get("name"),
		City:        r.PostForm.Get("city"),
		State:       r.PostForm.Get("state"),
		Country:     r.PostForm.Get("country"),
		Handle:      r.PostForm.Get("handle"),
		TopRated:    r.PostForm.Get("top-rated") == "true",
		Public:      r.PostForm.Get("public") == "true",
		FieldErrors: map[string]string{},
	}

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

	if toInsert.Name == "" {
		form.FieldErrors["name"] = "Name is required"
	}

	if toInsert.City == "" {
		form.FieldErrors["city"] = "City is required"
	}

	if toInsert.State == "" {
		form.FieldErrors["state"] = "State is required"
	}

	if toInsert.Country == "" {
		form.FieldErrors["country"] = "Country is required"
	}

	if toInsert.TwitterHandle == "" {
		form.FieldErrors["handle"] = "Handle is required"
	}

	if len(form.FieldErrors) > 0 {
		app.infoLog.Println("Errors found in form")
		data := &templateData{
			SchoolAddPage: schoolAddPage{
				Form: form,
			},
		}
		schools, err := models.GetAllSchools(app.connection)
		if err != nil {
			app.serverError(w, err)
			return
		}
		data.SchoolAddPage.Schools = schools
		app.populateStatusData(data)
		app.renderTemplate(w, http.StatusUnprocessableEntity, "schoolAdd.html", data)
		return
	}

	toScrape.IsSchool = true
	toScrape.SchoolInfo = toInsert
	toScrape.Username = toInsert.TwitterHandle

	app.profileChan <- toScrape

	http.Redirect(w, r, "/schools", http.StatusSeeOther)
}
