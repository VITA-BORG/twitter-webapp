package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
)

// Contains the data that will be passed to the templates.
type statusData struct {
	FollowerStatus    string
	FollowingStatus   string
	ProfileStatus     string
	ConnectionsStatus string
	NumberOfUsers     int
}

type usersPage struct {
	Participants    []models.User
	NumParticipants int
}

type userAddPage struct {
	Schools []models.School
	Form    any
}

type schoolAddPage struct {
	Schools []models.School
	Form    any
}

type userViewPage struct {
	CurrentUser models.User
	Schools     []models.School
	Form        any
}

type adminSignupPage struct {
	Form any
}

type adminLoginPage struct {
	Form any
}

type templateData struct {
	StatusData      statusData
	UsersPage       usersPage
	UserAddPage     userAddPage
	SchoolAddPage   schoolAddPage
	UserViewPage    userViewPage
	AdminSignupPage adminSignupPage
	AdminLoginPage  adminLoginPage
	Flash           string
	IsAdmin         bool
	CSRFToken       string
}

var functions = template.FuncMap{
	"currentDate": currDateFormatter,
}

func currDateFormatter() string {
	return time.Now().Format("January 1 2006 at 15:04")
}

// newTemplateCache is a helper function that loads all HTML templates into a template cache, and returns a map of template names to template.
// This will make it easy to render templates in the future, since the templates will be in the cache already and you will not have to parse them for every request.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {

		name := filepath.Base(page)

		//Parse base template file and add to set
		//Registers functions in the template, allowing them to be called from the template
		tmpl, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		//Parse partials and add to set
		tmpl, err = tmpl.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		//Parse page and add to set
		tmpl, err = tmpl.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = tmpl

	}
	return cache, nil
}

// renderTemplate is a helper function that renders a template with the given name and data.
func (app *application) renderTemplate(w http.ResponseWriter, status int, page string, data *templateData) {
	//Retrieves template from cache
	tmpl, ok := app.templateCache[page]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", page))
		return
	}

	//Writes template to buffer to check errors
	buf := new(bytes.Buffer)
	err := tmpl.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	//Writes buffer to response
	w.WriteHeader(status)
	buf.WriteTo(w)

}

// populateWorkerStatus is a helper function that populates the FollowerStatus and FollowingStatus fields of the templateData struct.
func (app *application) populateStatusData(data *templateData) {
	data.StatusData = statusData{
		FollowerStatus:    app.followStatus,
		FollowingStatus:   app.followingStatus,
		ProfileStatus:     app.profileStatus,
		ConnectionsStatus: app.connectionsStatus,
	}

	data.StatusData.NumberOfUsers, _ = models.GetUserCount(app.connection)
}

// populateTemplateData is a helper function that populates the templateData struct with the data needed to render the templates.
func (app *application) populateTemplateData(r *http.Request, data *templateData) {
	app.populateStatusData(data)
	data.Flash = app.sessionManager.PopString(r.Context(), "flash")
	data.IsAdmin = app.isAdmin(r)
}
