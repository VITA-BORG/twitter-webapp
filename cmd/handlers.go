package main

import (
	"encoding/json"
	"net/http"
)

//userAPI is a handler for the /api/user endpoint.
//GET /api/user?handle=username returns a JSON representation of the user.
//POST /api/user?handle=username Adds or updates user in the database with the given handle and returns their JSON representation.
func (app *application) userAPI(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodGet+","+http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodGet {
		userHandle := r.URL.Query().Get("handle")
		if userHandle == "" {
			app.clientError(w, http.StatusBadRequest)
		}

		user, err := app.getUserByHandle(userHandle)
		if err != nil {
			app.notFound(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)

	}
}
