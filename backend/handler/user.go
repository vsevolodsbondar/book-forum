package handler

import (
	"encoding/json"
	"fmt"
	"forum_backend/model"
	"net/http"
	"strconv"
)

func (app *Application) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r.PathValue("id"))
	if err != nil {
		///////REMINDER: make a unified error writer
	}

	user, err := app.UserService.GetUser(ctx, id)
	if err != nil {
		///////REMINDER: make a unified error writer
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func parseID(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64) // int64 equivalent of Atoi
	if err != nil || id < 1 {
		return 0, fmt.Errorf("id must be positive integer")
	}

	return id, nil
}

func (app *Application) PostUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var sub model.UserSubmission

	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		///////REMINDER: make a unified error writer
	}

	user, err := app.UserService.CreateUser(ctx, sub)
	if err != nil {
		//////REMINDER: make a unified error writer
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
