package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"uuid"

	"forum_backend/client"
	"forum_backend/custom_err"
	"forum_backend/helper"
	"forum_backend/model"
	"forum_backend/service"
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

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var input client.LoginUserRequestDTO

	// get fields
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}
		return custom_err.ErrInvalidInput
	}

	// send req to auth to create session
	loginResp, err := h.auth.LoginUser(ctx, input)
	if err != nil {
		return err
	}

	// set cookie
	helper.SetCookie(w, loginResp.Session.Token, loginResp.Session.ExpiresAt)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(loginResp.Session.ID)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	// get token
	token, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return err
	}
	// send token to auth for session delete
	err = h.auth.LogoutUser(ctx, token)
	if err != nil {
		return err
	}

	// clear Cookie
	helper.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
	return nil
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
		parsedPage, err := strconv.Atoi(page)
		if err != nil || parsedPage < 1 {
			return fmt.Errorf("%w: invalid page", custom_err.ErrInvalidInput)
		}
		pageInt = parsedPage
	}

	if size != "" {
		parsedSize, err := strconv.Atoi(size)
		if err != nil || parsedSize <= 0 {
			return fmt.Errorf("%w: invalid size number", custom_err.ErrInvalidInput)
		}
		sizeInt = parsedSize
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

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}
		return custom_err.ErrInvalidInput

	}

	var input model.UserUpdateDTO
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	if username, ok := r.MultipartForm.Value["username"]; ok {
		input.UserName = &username[0]
	}
	if name, ok := r.MultipartForm.Value["name"]; ok {
		input.Name = &name[0]
	}
	if description, ok := r.MultipartForm.Value["description"]; ok {
		input.Description = &description[0]
	}
	if avatar, err := getAvatar(r.MultipartForm.File); err != nil {
		return err
	} else if avatar != "" {
		input.ProfilePicture = &avatar
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

func getAvatar(files map[string][]*multipart.FileHeader) (string, error) {
	const (
		maxAvatarSize = 8 << 20 // 8 MiB
		avatarDir     = "./data/avatars"
	)

	for field, values := range files {
		if field != "avatar" || len(values) != 1 {
			return "", custom_err.ErrInvalidInput
		}
	}

	avatarFiles := files["avatar"]
	if len(avatarFiles) == 0 {
		return "", nil
	}

	file, err := avatarFiles[0].Open()
	if err != nil {
		return "", fmt.Errorf("open uploaded avatar: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxAvatarSize+1)) // +1 allows determining if we have an illegal size later
	if err != nil {
		return "", fmt.Errorf("read uploaded avatar: %w", err)
	}

	n := len(data)
	if n == 0 || n > maxAvatarSize {
		return "", custom_err.ErrInvalidInput
	}

	var ext string
	switch http.DetectContentType(data) {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	default:
		return "", custom_err.ErrInvalidInput
	}

	if err := os.MkdirAll(avatarDir, 0o700); err != nil {
		return "", fmt.Errorf("create avatars directory: %w", err)
	}

	avatar := uuid.NewV4().String() + ext
	if err := os.WriteFile(filepath.Join(avatarDir, avatar), data, 0o600); err != nil {
		return "", fmt.Errorf("save avatar: %w", err)
	}

	return avatar, nil
}
