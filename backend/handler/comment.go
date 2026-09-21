package handler

import (
	"encoding/json"
<<<<<<< HEAD
=======
	"forum_backend/client"
	"forum_backend/helper"
>>>>>>> back-auth-connection
	"forum_backend/model"
	"forum_backend/service"
	"net/http"
	"strconv"
)

type CommentHandler struct {
	service *service.CommentService
<<<<<<< HEAD
}

func NewCommentHandler(service *service.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
=======
	auth    client.AuthInterface
}

func NewCommentHandler(service *service.CommentService, authService client.AuthInterface) *CommentHandler {
	return &CommentHandler{service: service, auth: authService}
>>>>>>> back-auth-connection
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
<<<<<<< HEAD
	comment, err := h.service.Create(ctx, &commentRaw)
=======

	sessionCookie, err := helper.ExtractSessionCookie(r)

	response, err := h.auth.ValidateSession(ctx, sessionCookie)

	commentDTO := model.CreateCommentDTO{
		Comment: commentRaw,
		UserID:  int(response.User.ID),
	}

	comment, err := h.service.Create(ctx, &commentDTO)
>>>>>>> back-auth-connection
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}
<<<<<<< HEAD
func (h *CommentHandler) GetAll(w http.ResponseWriter, r *http.Request) {
=======

func (h *CommentHandler) GetAllByPostID(w http.ResponseWriter, r *http.Request) {
>>>>>>> back-auth-connection
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
<<<<<<< HEAD
	comments, err := h.service.GetAll(ctx, pageInt, sizeInt, idPost)
=======
	//dummy user_id
	userID := 1
	commentDTO := model.GetAllCommentDTO{
		PageInt: pageInt,
		SizeInt: sizeInt,
		IDPost:  idPost,
		UserID:  userID,
	}
	comments, err := h.service.GetAllByPostID(ctx, commentDTO)
>>>>>>> back-auth-connection
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
<<<<<<< HEAD
	updatedComment, err := h.service.Update(ctx, comentToUpdate, commentID)
=======
	//dummy user_id
	userID := 1
	commentDTO := model.UpdateCommentDTO{
		CommentToUpdate: comentToUpdate,
		CommentID:       commentID,
		UserID:          userID,
	}
	updatedComment, err := h.service.Update(ctx, commentDTO)
>>>>>>> back-auth-connection
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
<<<<<<< HEAD
	err = h.service.Delete(ctx, commentID)
=======
	//dummy user_id
	userID := 1
	commentDTO := model.DeleteCommentDTO{
		CommentID: commentID,
		UserID:    userID,
	}
	err = h.service.Delete(ctx, commentDTO)
>>>>>>> back-auth-connection
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
