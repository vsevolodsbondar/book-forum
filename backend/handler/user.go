package handler

import (
	"encoding/json"
	"errors"
	"fmt"
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
	var input model.UserDTO

	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}
		return custom_err.ErrInvalidInput
	}

	authResp, err := h.auth.RegisterUser(ctx, client.RegisterUserRequestDTO{
		Email:    input.Email,
		Username: input.UserName,
		Password: input.Password,
	})

	if err != nil {
		return err
	}

	user, err := h.service.CreateUser(ctx, input, authResp)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) error {
	page := r.URL.Query().Get("page")
	size := r.URL.Query().Get("size")
	ctx := r.Context()
	pageInt, sizeInt := 1, 20
	if page != "" {
		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt < 1 {
			return fmt.Errorf("%w: invalid id", custom_err.ErrInvalidInput)
		}
	}

	if size != "" {
		sizeInt, err := strconv.Atoi(size)
		if err != nil || sizeInt <= 0 {
			return fmt.Errorf("%w: invalid size number", custom_err.ErrInvalidInput)
		}
	}

	GetAllUsersParams := model.GetAllUsersDTO{Page: pageInt, Size: sizeInt}

	usersInfo, err := h.service.GetAllUsers(ctx, GetAllUsersParams)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usersInfo)
	return nil

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

	var input model.UserUpdateDTO
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
