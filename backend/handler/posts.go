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
)

type PostHandler struct {
	Service *service.PostService
	Auth    client.AuthInterface
}

func NewPostHandler(service *service.PostService, authService client.AuthInterface) *PostHandler {
	return &PostHandler{Service: service, Auth: authService}
}

func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	page, limit, err := helper.GetPaginationParams(r)
	if err != nil {
		return fmt.Errorf("%w: invalid pagination", custom_err.ErrInvalidInput)
	}

	dto := parseSearchPostsDTO(r.URL.Query(), page, limit)

	posts, err := h.Service.GetAllPosts(ctx, dto)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(posts)
}

func parseSearchPostsDTO(query url.Values, page int, limit int) model.SearchPostsDTO {
	dto := model.SearchPostsDTO{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	// Search
	search := query.Get("search")
	if search != "" {
		dto.IsSearch = true
	}

	// Search field
	if field := query.Get("field"); field != "" {
		dto.SearchField = &field
	}

	// Search value
	if value := query.Get("value"); value != "" {
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

	return dto
}
