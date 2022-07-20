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

//Containst the data that will be passed to the templates.
type templateData struct {
	CurrentUser     models.User
	FollowerStatus  string
	FollowingStatus string
	ProfileStatus   string
	Users           []models.User
	Schools         []models.School
	NumberOfUsers   int
	NumParticipants int
}

var functions = template.FuncMap{
	"currentDate": currDateFormatter,
}

func currDateFormatter() string {
	return time.Now().Format("January 1 2006 at 15:04")
}

//newTemplateCache is a helper function that loads all HTML templates into a template cache, and returns a map of template names to template.
//This will make it easy to render templates in the future, since the templates will be in the cache already and you will not have to parse them for every request.
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

//renderTemplate is a helper function that renders a template with the given name and data.
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

//populateWorkerStatus is a helper function that populates the FollowerStatus and FollowingStatus fields of the templateData struct.
func (app *application) populateWorkerStatus(data *templateData) {
	data.FollowerStatus = app.followStatus
	data.FollowingStatus = app.followingStatus
	data.ProfileStatus = app.profileStatus
	data.NumberOfUsers = len(app.getAllUsernames())
}
