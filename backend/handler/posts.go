package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"forum_backend/client"
	"forum_backend/custom_err"
	"forum_backend/helper"
	"forum_backend/model"
	"forum_backend/service"
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

	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&dto); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}

		return fmt.Errorf("%w: invalid json format", custom_err.ErrInvalidInput)
	}

	sessionCookie, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return err
	}
	response, err := h.Auth.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return err
	}

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
	if err != nil {
		return err
	}
	response, err := h.Auth.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return err
	}

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

	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&dto); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}

		return fmt.Errorf("%w: invalid json format", custom_err.ErrInvalidInput)
	}

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

	// Category
	category := query.Get("categoryId")
	if category != "" {
		id, err := strconv.Atoi(category)
		if err != nil || id <= 0 {
			return model.SearchPostsDTO{},
				fmt.Errorf("%w: invalid category ID", custom_err.ErrInvalidInput)
		}

		dto.CategoryID = &id
	}

	// Reaction filter
	reaction := query.Get("reaction")
	userID := query.Get("userID")

	switch reaction {
	case "":
		if userID != "" {
			return model.SearchPostsDTO{},
				fmt.Errorf("%w: userID requires a reaction filter",
					custom_err.ErrInvalidInput)
		}
	case "liked", "disliked":
		if userID == "" {
			return model.SearchPostsDTO{},
				fmt.Errorf("%w: userID required for reaction filter",
					custom_err.ErrInvalidInput)
		}

		dto.Reaction = &reaction
		dto.UserID = &userID
	default:
		return model.SearchPostsDTO{},
			fmt.Errorf("%w: invalid reaction filter",
				custom_err.ErrInvalidInput)
	}

	// Ordering
	switch query.Get("byLatest") {
	case "", "true":
		dto.IsLatestPostsFirst = true
	case "false":
		dto.IsLatestPostsFirst = false
	default:
		return model.SearchPostsDTO{}, fmt.Errorf("%w: invalid sort order", custom_err.ErrInvalidInput)
	}

	// Search
	title := strings.TrimSpace(query.Get("title"))
	if title != "" {
		if utf8.RuneCountInString(title) < 3 ||
			utf8.RuneCountInString(title) > 100 {
			return model.SearchPostsDTO{},
				fmt.Errorf("%w: title must be 3–100 characters",
					custom_err.ErrInvalidInput)
		}

		dto.Title = &title
	}

	dto.IsSearch = dto.Title != nil || dto.CategoryID != nil || dto.Reaction != nil

	return dto, nil
}
