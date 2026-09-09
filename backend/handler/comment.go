package handler

import (
	"encoding/json"
	"forum_backend/model"
	"forum_backend/service"
	"net/http"
)

type CommentHandler struct {
	service *service.CommentService
}

func NewCommentHandler(service *service.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var commentRaw model.CreateCommentRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&commentRaw)
	if err != nil {
		return http.Error(w, err.Error(), http.StatusBadRequest)
	}
	ctx := r.Context()
}
