package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"forum_backend/model"
	"forum_backend/service"
	"net/http"
	"strconv"
)

// limit for request body. To prevent reading 1tb JSON into memory
const maxReqBodySize = 1024 * 1024

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sub model.UserSubmission

	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&sub)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			///////REMINDER: make a unified error writer
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		///////REMINDER: make a unified error writer
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.CreateUser(ctx, sub)
	if err != nil {
		///////REMINDER: make a unified error writer
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUser(ctx, id)
	if err != nil {
		///////REMINDER: make a unified error writer
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r.PathValue("id"))
	if err != nil {
		///////REMINDER: make a unified error writer
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)

	var input model.UserUpdateInfo
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		///////REMINDER: make a unified error writer
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	err = h.service.UpdateUser(ctx, id, input)
	if err != nil {
		///////REMINDER: make a unified error writer
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseID(r.PathValue("id"))
	if err != nil {
		///////REMINDER: make a unified error writer
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.service.DeleteUser(ctx, id)
	if err != nil {
		///////REMINDER: make a unified error writer
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseID(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("id must be positive integer")
	}

	return id, nil
}
