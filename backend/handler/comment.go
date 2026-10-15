package handler

import (
	"encoding/json"
	"errors"
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

	sessionCookie, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return err
	}
	response, err := h.auth.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return err
	}

	// 2. what did they send?
	var body model.CreateCommentBody
	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}
		return fmt.Errorf("%w: invalid json format", custom_err.ErrInvalidInput)
	}

	// 3. assemble the DTO: text/parent from body, post from URL, user from session
	dto := model.CreateCommentDTO{
		Comment: model.CreateCommentRequest{
			Text:            body.Text,
			PostID:          int64(idPost),
			ParentCommentID: body.ParentCommentID,
		},
		UserID: int(response.User.ID),
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

	userID := 0
	if sessionCookie, err := helper.ExtractSessionCookie(r); err == nil {
		if response, err := h.auth.ValidateSession(ctx, sessionCookie); err == nil {
			userID = int(response.User.ID)
		}
	}

	commentDTO := model.GetAllCommentDTO{
		PageInt: pageInt,
		SizeInt: sizeInt,
		IDPost:  idPost,
		UserID:  userID,
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

	sessionCookie, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return err
	}
	response, err := h.auth.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return err
	}

	var body model.CommentPatchRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}
		return fmt.Errorf("%w: invalid json format", custom_err.ErrInvalidInput)
	}

	commentDTO := model.UpdateCommentDTO{
		CommentToUpdate: body,
		CommentID:       commentID,
		UserID:          int(response.User.ID),
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
	if sessionCookie, err := helper.ExtractSessionCookie(r); err == nil {
		if response, err := h.auth.ValidateSession(ctx, sessionCookie); err == nil {
			userID = int(response.User.ID)
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
	ctx := r.Context()
	commentID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("%w: invalid id", custom_err.ErrInvalidInput)
	}

	sessionCookie, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return err
	}
	response, err := h.auth.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return err
	}

	commentDTO := model.DeleteCommentDTO{
		CommentID: commentID,
		UserID:    int(response.User.ID),
	}
	if err := h.service.Delete(ctx, commentDTO); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
