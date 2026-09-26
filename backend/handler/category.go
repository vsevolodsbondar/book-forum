package handler

import (
	"encoding/json"
	"forum_backend/service"
	"net/http"
)

type CategoryHandler struct {
	service *service.CategoryService
}

func NewCategoryHandler(service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) error {
	categories, err := h.service.GetAll(r.Context())
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(categories)
}
