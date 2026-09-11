package handler

import (
	"encoding/json"
	"forum_backend/model"
	"forum_backend/service"
	"net/http"
	"strconv"
)

type CommentHandler struct {
	service *service.CommentService
}

func NewCommentHandler(service *service.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var commentRaw model.CreateCommentRequest
	var commentBody model.CreateCommentBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&commentBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	idPost, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	commentRaw.Text = commentBody.Text
	commentRaw.PostID = int64(idPost)
	ctx := r.Context()
	_, err = h.service.Create(ctx, &commentRaw)

}
