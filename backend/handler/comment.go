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
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var commentRaw model.CreateCommentRequest
	var commentBody model.CreateCommentBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&commentBody)
	if err != nil {
		return err
	}
	idPost, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid id: %w", custom_err.ErrInvalidInput)
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
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
	return nil
}

func (h *CommentHandler) GetAllByPostID(w http.ResponseWriter, r *http.Request) error {
	page := r.URL.Query().Get("page")
	size := r.URL.Query().Get("size")
	ctx := r.Context()
	idPost, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return err
	}
	pageInt, sizeInt := 1, 20
	if page != "" {
		pageInt, err = strconv.Atoi(page)
		if err != nil || pageInt < 1 {
			return fmt.Errorf("invalid id: %w", custom_err.ErrInvalidInput)
		}
	}
	if size != "" {
		sizeInt, err = strconv.Atoi(size)
		if err != nil || sizeInt <= 0 {
			return fmt.Errorf("invalid size number: %w", custom_err.ErrInvalidInput)
		}
	}
	sessionCookie, err := helper.ExtractSessionCookie(r)

	response, err := h.auth.ValidateSession(ctx, sessionCookie)
	commentDTO := model.GetAllCommentDTO{
		PageInt: pageInt,
		SizeInt: sizeInt,
		IDPost:  idPost,
		UserID:  int(response.User.ID),
	}
	comments, err := h.service.GetAllByPostID(ctx, commentDTO)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
	return nil
}
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	commentID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid id: %w", custom_err.ErrInvalidInput)
	}
	comentToUpdate := model.CommentPatchRequest{}
	err = json.NewDecoder(r.Body).Decode(&comentToUpdate)
	if err != nil {
		return fmt.Errorf("invalid json format: %w", custom_err.ErrInvalidInput)
	}
	sessionCookie, err := helper.ExtractSessionCookie(r)

	response, err := h.auth.ValidateSession(ctx, sessionCookie)
	commentDTO := model.UpdateCommentDTO{
		CommentToUpdate: comentToUpdate,
		CommentID:       commentID,
		UserID:          int(response.User.ID),
	}
	updatedComment, err := h.service.Update(ctx, commentDTO)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedComment)
	return nil
}
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	commentID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("invalid id: %w", custom_err.ErrInvalidInput)
	}
	ctx := r.Context()
	sessionCookie, err := helper.ExtractSessionCookie(r)

	response, err := h.auth.ValidateSession(ctx, sessionCookie)
	commentDTO := model.DeleteCommentDTO{
		CommentID: commentID,
		UserID:    int(response.User.ID),
	}
	err = h.service.Delete(ctx, commentDTO)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
