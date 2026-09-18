package handler

import (
	"encoding/json"
	"fmt"
	"forum_backend/model"
	"forum_backend/service"
	"net/http"
	"strconv"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sub model.UserSubmission

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&sub)
	if err != nil {
		///////REMINDER: make a unified error writer
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.CreateUser(ctx, sub)
	if err != nil {
		//////REMINDER: make a unified error writer
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r.PathValue("id"))
	if err != nil {
		///////REMINDER: make a unified error writer
	}

	user, err := h.service.GetUser(ctx, id)
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
