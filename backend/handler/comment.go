package handler

import (
	"encoding/json"
	"forum_backend/client"
	"forum_backend/helper"
	"forum_backend/model"
	"forum_backend/service"
	"net/http"
	"strconv"
)

type CommentHandler struct {
	service *service.CommentService
	auth    client.AuthInterface
}

func NewCommentHandler(service *service.CommentService, authService client.AuthInterface) *CommentHandler {
	return &CommentHandler{service: service, auth: authService}
}

// on all hanlders has be a check of userID
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

	sessionCookie, err := helper.ExtractSessionCookie(r)

	response, err := h.auth.ValidateSession(ctx, sessionCookie)

	commentDTO := model.CreateCommentDTO{
		Comment: commentRaw,
		UserID:  int(response.User.ID),
	}

	comment, err := h.service.Create(ctx, &commentDTO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) GetAllByPostID(w http.ResponseWriter, r *http.Request) {
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
	//dummy user_id
	userID := 1
	commentDTO := model.GetAllCommentDTO{
		PageInt: pageInt,
		SizeInt: sizeInt,
		IDPost:  idPost,
		UserID:  userID,
	}
	comments, err := h.service.GetAllByPostID(ctx, commentDTO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commentID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	comentToUpdate := model.CommentPatchRequest{}
	err = json.NewDecoder(r.Body).Decode(&comentToUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//dummy user_id
	userID := 1
	commentDTO := model.UpdateCommentDTO{
		CommentToUpdate: comentToUpdate,
		CommentID:       commentID,
		UserID:          userID,
	}
	updatedComment, err := h.service.Update(ctx, commentDTO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedComment)
}
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	commentID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	//dummy user_id
	userID := 1
	commentDTO := model.DeleteCommentDTO{
		CommentID: commentID,
		UserID:    userID,
	}
	err = h.service.Delete(ctx, commentDTO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
