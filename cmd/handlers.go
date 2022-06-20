package main

import (
	"encoding/json"
	"net/http"

	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
)

func (app *application) user(w http.ResponseWriter, r *http.Request) {

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

		if !models.UserExists(app.connection, userHandle) {
			app.notFound(w)
			return
		}

		user, err := models.GetUserByHandle(app.connection, userHandle)
		if err != nil {
			app.serverError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)

	}
}
