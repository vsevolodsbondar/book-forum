package handler

import (
	"fmt"
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
}

func parseID(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64) // int64 equivalent of Atoi
	if err != nil || id < 1 {
		return 0, fmt.Errorf("id must be positive integer")
	}

	return id, nil
}
