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
	ctx := r.Context()

	idPost, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("%w: invalid id", custom_err.ErrInvalidInput)
	}

	// 1. who is it? Check this before touching the body
	cookie, err := r.Cookie("session")
	if err != nil {
		return fmt.Errorf("%w: no session cookie", custom_err.ErrExpiredSession)
	}
	session, err := h.auth.ValidateSession(ctx, cookie.Value)

	// 2. what did they send?
	var body model.CreateCommentBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		return fmt.Errorf("%w: invalid json format", custom_err.ErrInvalidInput)
	}

	// 3. assemble the DTO: text/parent from body, post from URL, user from session
	dto := model.CreateCommentDTO{
		Comment: model.CreateCommentRequest{
			Text:            body.Text,
			PostID:          int64(idPost),
			ParentCommentID: body.ParentCommentID,
		},
		UserID: int(session.User.ID),
	}

	comment, err := h.service.Create(ctx, &dto)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(comment)
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
			return fmt.Errorf("%w: invalid id", custom_err.ErrInvalidInput)
		}
	}
	if size != "" {
		sizeInt, err = strconv.Atoi(size)
		if err != nil || sizeInt <= 0 {
			return fmt.Errorf("%w: invalid size number", custom_err.ErrInvalidInput)
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
		return fmt.Errorf("%w: invalid id", custom_err.ErrInvalidInput)
	}

	cookie, err := r.Cookie("session")
	if err != nil {
		return fmt.Errorf("%w: no session cookie", custom_err.ErrExpiredSession)
	}
	session, err := h.auth.ValidateSession(ctx, cookie.Value)
	if err != nil {
		return fmt.Errorf("%w: session validation failed", custom_err.ErrSessionValidationFail)
	}

	var body model.CommentPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return fmt.Errorf("%w: invalid json format", custom_err.ErrInvalidInput)
	}

	commentDTO := model.UpdateCommentDTO{
		CommentToUpdate: body,
		CommentID:       commentID,
		UserID:          int(session.User.ID),
	}
	updatedComment, err := h.service.Update(ctx, commentDTO)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(updatedComment)
}
func (h *CommentHandler) GetByID(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	commentID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("%w: invalid id", custom_err.ErrInvalidInput)
	}

	// session is optional: guests can read, they just have no likes of their own
	userID := 0
	if cookie, err := r.Cookie("session"); err == nil {
		if session, err := h.auth.ValidateSession(ctx, cookie.Value); err == nil {
			userID = int(session.User.ID)
		}
	}

	comment, err := h.service.GetByID(ctx, commentID, userID)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(comment)
}
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	commentID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("%w: invalid id", custom_err.ErrInvalidInput)
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
