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

// not done fully
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
func (h *CommentHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	size := r.URL.Query().Get("size")
	ctx := r.Context()
	idPost, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pageInt, sizeInt := 1, 20
	if page != "" {
		pageInt, err = strconv.Atoi(page)
		if err != nil || pageInt < 1 {
			http.Error(w, "invalid page number", http.StatusBadRequest)
			return
		}
	}
	if size != "" {
		sizeInt, err = strconv.Atoi(size)
		if err != nil || sizeInt <= 0 {
			http.Error(w, "invalid size number", http.StatusBadRequest)
			return
		}
	}
	comments, err := h.service.GetAll(ctx, pageInt, sizeInt, idPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
