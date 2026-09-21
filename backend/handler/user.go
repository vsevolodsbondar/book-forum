package handler

import (
	"encoding/json"
	"errors"
	"forum_backend/client"
	"forum_backend/custom_err"
	"forum_backend/helper"
	"forum_backend/model"
	"forum_backend/service"
	"net/http"
	"strconv"
)

// limit for request body. To prevent reading 1tb JSON into memory
const maxReqBodySize = 1024 * 1024

type UserHandler struct {
	service *service.UserService
	auth    client.AuthInterface
}

func NewUserHandler(service *service.UserService, auth client.AuthInterface) *UserHandler {
	return &UserHandler{
		service: service,
		auth:    auth,
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	var sub model.UserDTO

	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&sub); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}
		return custom_err.ErrInvalidInput
	}

	user, err := h.service.CreateUser(ctx, sub)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	id, err := parseID(r.PathValue("id"))
	if err != nil {
		return custom_err.ErrInvalidInput
	}

	user, err := h.service.GetUser(ctx, id)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	id, err := parseID(r.PathValue("id"))
	if err != nil {
		return custom_err.ErrInvalidInput
	}

	sessionCookie, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return custom_err.ErrExpiredSession
	}

	session, err := h.auth.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return custom_err.ErrExpiredSession
	}

	if session.User.ID != id {
		return custom_err.ErrForbidden
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)

	var input model.UserUpdateInfo
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}
		return custom_err.ErrInvalidInput
	}

	if err := h.service.UpdateUser(ctx, id, input); err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	id, err := parseID(r.PathValue("id"))
	if err != nil {
		return custom_err.ErrInvalidInput
	}

	sessionCookie, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return custom_err.ErrExpiredSession
	}

	session, err := h.auth.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return custom_err.ErrExpiredSession
	}

	if session.User.ID != id {
		return custom_err.ErrForbidden
	}

	if err := h.service.DeleteUser(ctx, id); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func parseID(idStr string) (int64, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		return 0, custom_err.ErrInvalidInput
	}
	return id, nil
}
