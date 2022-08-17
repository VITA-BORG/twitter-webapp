package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
	"github.com/rainbowriverrr/F3Ytwitter/internal/validation"
)

type userViewForm struct {
	Handle    string `form:"handle"`
	School    string `form:"school"`
	StartDate string `form:"start-date"`
	Follows   bool   `form:"follows"`
	Content   bool   `form:"content"`
	Cohort    string `form:"cohort"`
	validation.Validator
}
type userAddForm struct {
	Handle    string `form:"handle"`
	School    string `form:"school"`
	StartDate string `form:"start-date"`
	Follows   bool   `form:"follows"`
	Content   bool   `form:"content"`
	Cohort    string `form:"cohort"`
	validation.Validator
}

type schoolAddForm struct {
	Name     string `form:"name"`
	City     string `form:"city"`
	State    string `form:"state"`
	Country  string `form:"country"`
	Handle   string `form:"handle"`
	TopRated bool   `form:"top-rated"`
	Public   bool   `form:"public"`
	validation.Validator
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
		Cohort: strconv.Itoa(student.Cohort),
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

	flash := app.sessionManager.PopString(r.Context(), "flash")
	data.Flash = flash

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

	form := userViewForm{
		Handle:    strings.TrimSpace(r.PostForm.Get("handle")),
		School:    strings.TrimSpace(r.PostForm.Get("school")),
		StartDate: strings.TrimSpace(r.PostForm.Get("start-date")),
		Cohort:    strings.TrimSpace(r.PostForm.Get("cohort")),
		Follows:   r.PostForm.Get("follows") == "true",
		Content:   r.PostForm.Get("content") == "true",
	}

	form.CheckField(validation.ValidInt(form.Cohort), "cohort", "Cohort must be a number")
	form.CheckField(validation.NotEmpty(form.Handle), "handle", "Handle is required")
	form.CheckField(validation.NotEmpty(form.School), "school", "School is required")
	form.CheckField(validation.NotEmpty(form.StartDate), "start-date", "Start Date is required")
	form.CheckField(validation.PermittedDate(form.StartDate), "start-date", "Start Date must be a valid date")

	if !form.Valid() {
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
		app.renderTemplate(w, http.StatusUnprocessableEntity, "userView.html", data)
		return
	}

	startDate, err := time.Parse("2006-01-02", form.StartDate)
	if err != nil {
		app.serverError(w, err)
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

	cohortInt, err := strconv.Atoi(form.Cohort)
	if err != nil {
		app.serverError(w, err)
		return
	}

	toScrape := &simplifiedUser{
		ID:                  uid,
		Username:            form.Handle,
		IsSchool:            false,
		IsParticipant:       true,
		ParticipantCohort:   cohortInt,
		ParticipantSchoolID: schoolID,
		StartDate:           startDate,
		ScrapeConnections:   form.Follows,
		ScrapeContent:       form.Content,
	}

	app.profileChan <- toScrape

	app.sessionManager.Put(r.Context(), "flash", "User updated successfully")

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

	flash := app.sessionManager.PopString(r.Context(), "flash")
	data.Flash = flash

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
		Handle:    strings.TrimSpace(r.PostForm.Get("handle")),
		School:    strings.TrimSpace(r.PostForm.Get("school")),
		StartDate: strings.TrimSpace(r.PostForm.Get("start-date")),
		Cohort:    strings.TrimSpace(r.PostForm.Get("cohort")),
		Follows:   strings.TrimSpace(r.PostForm.Get("follows")) == "true",
		Content:   strings.TrimSpace(r.PostForm.Get("content")) == "true",
	}

	form.CheckField(validation.NotEmpty(form.Handle), "handle", "Handle is required")
	form.CheckField(validation.NotEmpty(form.School), "school", "School is required")
	form.CheckField(validation.NotEmpty(form.StartDate), "start-date", "Start Date is required")
	form.CheckField(validation.PermittedDate(form.StartDate), "start-date", "Start Date must be a valid date")
	form.CheckField(validation.NotEmpty(form.Cohort), "cohort", "Cohort is required")
	form.CheckField(validation.ValidInt(form.Cohort), "cohort", "Cohort must be a valid integer")

	//if there are any errors, render the form again with the field errors and repopulated fields
	if !form.Valid() {
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

	cohort, err := strconv.Atoi(form.Cohort)
	if err != nil {
		app.serverError(w, err)
		return
	}

	schoolID, err := models.GetSchoolIDByName(app.connection, form.School)
	if err != nil {
		app.serverError(w, err)
		return
	}

	startDate, err := time.Parse("2006-01-02", form.StartDate)

	if err != nil {
		startDate = time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
		return
	}

	toScrape := &simplifiedUser{
		Username:            form.Handle,
		IsSchool:            false,
		IsParticipant:       true,
		ScrapeConnections:   form.Follows,
		ScrapeContent:       form.Content,
		ParticipantCohort:   cohort,
		ParticipantSchoolID: schoolID,
		StartDate:           startDate,
	}

	//sends the user to the scraper to scrape the user's profile
	app.profileChan <- toScrape

	app.sessionManager.Put(r.Context(), "flash", "User added successfully")

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
	flash := app.sessionManager.PopString(r.Context(), "flash")
	data.Flash = flash
	app.renderTemplate(w, http.StatusOK, "schoolAdd.html", data)
	fmt.Fprintf(w, "School Add Form")
}

func (app *application) schoolAddPost(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	form := schoolAddForm{
		Name:     strings.ToLower(strings.TrimSpace(r.PostForm.Get("name"))),
		City:     strings.ToLower(strings.TrimSpace(r.PostForm.Get("city"))),
		State:    strings.ToLower(strings.TrimSpace(r.PostForm.Get("state"))),
		Country:  strings.ToLower(strings.TrimSpace(r.PostForm.Get("country"))),
		Handle:   strings.ToLower(strings.TrimSpace(r.PostForm.Get("handle"))),
		TopRated: r.PostForm.Get("top-rated") == "true",
		Public:   r.PostForm.Get("public") == "true",
	}

	form.CheckField(validation.NotEmpty(form.Name), "name", "Name is required")
	form.CheckField(validation.NotEmpty(form.City), "city", "City is required")
	form.CheckField(validation.NotEmpty(form.State), "state", "State is required")
	form.CheckField(validation.NotEmpty(form.Country), "country", "Country is required")
	form.CheckField(validation.NotEmpty(form.Handle), "handle", "Handle is required")

	if !form.Valid() {
		app.infoLog.Println("Errors found in School Add Form")
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

	toInsert := &simplifiedSchool{
		Name:          form.Name,
		City:          form.City,
		State:         form.State,
		Country:       form.Country,
		TwitterHandle: form.Handle,
		TopRated:      form.TopRated,
		Public:        form.Public,
	}

	toScrape := &simplifiedUser{
		IsSchool:   true,
		Username:   form.Handle,
		SchoolInfo: toInsert,
	}

	app.profileChan <- toScrape

	app.sessionManager.Put(r.Context(), "flash", "School added successfully")

	http.Redirect(w, r, "/schools", http.StatusSeeOther)
}
