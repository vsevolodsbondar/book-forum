package handler

import (
	"encoding/json"
	"fmt"
	"forum_backend/client"
	"forum_backend/custom_err"
	"forum_backend/helper"
	"forum_backend/model"
	"forum_backend/service"
	"net/http"
	"net/url"
	"strconv"
)

type PostHandler struct {
	Service *service.PostService
	Auth    client.AuthInterface
}

func NewPostHandler(service *service.PostService, authService client.AuthInterface) *PostHandler {
	return &PostHandler{Service: service, Auth: authService}
}

func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) error {
	page, limit, err := helper.GetPaginationParams(r)
	if err != nil {
		return fmt.Errorf("%w: invalid pagination", custom_err.ErrInvalidInput)
	}

	dto, err := parseSearchPostsDTO(r.URL.Query(), page, limit)
	if err != nil {
		return err
	}

	posts, err := h.Service.GetAllPosts(r.Context(), dto)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) PostPost(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	dto := model.CreatePostRequestDTO{}

	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		return custom_err.ErrInvalidInput
	}

	sessionCookie, err := helper.ExtractSessionCookie(r)
	response, err := h.Auth.ValidateSession(ctx, sessionCookie)

	res, err := h.Service.PostMaker(ctx, dto, response.User.ID)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(res)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	dto := model.DeletePostDTO{}

	postID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return custom_err.ErrInvalidID
	}

	sessionCookie, err := helper.ExtractSessionCookie(r)
	response, err := h.Auth.ValidateSession(ctx, sessionCookie)

	dto.PostID = int64(postID)
	dto.UserID = int64(response.User.ID)

	res, err := h.Service.DeletePost(ctx, dto)
	if err != nil {
		return err
	}

	if res {
		w.WriteHeader(http.StatusNoContent)
	}

	w.Header().Set("Content-Type", "application/json")
	return nil
}

func (h *PostHandler) GetPostByID(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	postID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return custom_err.ErrInvalidID
	}

	res, err := h.Service.GetByID(ctx, int64(postID))
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(res)
}

func (h *PostHandler) PatchPost(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	postID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return custom_err.ErrInvalidID
	}

	dto := model.PatchPostDTO{}

	err = json.NewDecoder(r.Body).Decode(&dto)

	sessionCookie, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return err
	}
	response, err := h.Auth.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return err
	}

	dto.PostID = int64(postID)
	dto.UserID = int64(response.User.ID)

	res, err := h.Service.PatchPost(ctx, dto)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(res)
}

func parseSearchPostsDTO(query url.Values, page int, limit int) (model.SearchPostsDTO, error) {
	dto := model.SearchPostsDTO{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	// Search
	search := query.Get("search")
	if search == "true" {
		// Search field
		field := query.Get("field")
		// Search value
		value := query.Get("value")

		if field == "" || value == "" {
			return model.SearchPostsDTO{}, fmt.Errorf("%w: if search is true - should fill search fields", custom_err.ErrInvalidInput)
		}

		dto.IsSearch = true
		dto.SearchField = &field
		dto.SearchValue = &value
	}

	// Ordering
	byLatest := query.Get("byLatest")

	switch byLatest {
	case "true":
		dto.IsLatestPostsFirst = true

	case "false":
		dto.IsLatestPostsFirst = false

	default:
		dto.IsLatestPostsFirst = true
	}

	return dto, nil
}
