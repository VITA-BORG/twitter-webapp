package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//userAPI is a handler for the /api/user endpoint.
//GET /api/user?handle=username returns a JSON representation of the user.
//PUT /api/user?handle=username Adds or updates user in the database with the given handle and returns their JSON representation.
func (app *application) userAPI(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet && r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodGet+","+http.MethodPut)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodGet {
		userHandle := r.URL.Query().Get("handle")
		if userHandle == "" {
			usernames := app.getAllUsernames()
			json.NewEncoder(w).Encode(usernames)
			return
		}

		user, err := app.getUserByHandle(userHandle)
		if err != nil {
			app.notFound(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)

	}

	//Takes user handle, along with a variety of other parameters and adds them to the scrape.
	if r.Method == http.MethodPut {
		fmt.Fprintf(w, "TODO: add or update user in database.")
	}
}

func (app *application) user(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "User Homepage")
}

//home is a handler for the root endpoint.  It shows a simple list of users in the database.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	Users := app.getAllUsernames()
	data := &templateData{
		Users: Users,
	}

	app.renderTemplate(w, http.StatusOK, "dashboard.html", data)

	fmt.Fprintf(w, "Homepage")
}
