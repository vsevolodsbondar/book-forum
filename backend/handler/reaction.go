package handler

import (
	"encoding/json"
	"errors"
	"forum_backend/client"
	"forum_backend/custom_err"
	"forum_backend/helper"
	"forum_backend/model"
	"forum_backend/service"
	"net/http"
)

type ReactionHandler struct {
	service *service.ReactionService
	auth    client.AuthInterface
}

func NewReactionHandler(service *service.ReactionService, auth client.AuthInterface) *ReactionHandler {
	return &ReactionHandler{
		service: service,
		auth:    auth,
	}
}

func (h *ReactionHandler) CreateReaction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	//check auth session
	sessionCookie, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return custom_err.ErrExpiredSession
	}

	session, err := h.auth.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return custom_err.ErrExpiredSession
	}

	//get input
	var input model.ReactionDTO
	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}
		return custom_err.ErrInvalidInput
	}
	//do not trust the body, trust the session
	input.UserID = session.User.ID
	//create reaction
	err = h.service.CreateReaction(ctx, input)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

func (h *ReactionHandler) UpdateReaction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	//check auth session
	sessionCookie, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return custom_err.ErrExpiredSession
	}

	session, err := h.auth.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return custom_err.ErrExpiredSession
	}

	//get input
	var input model.ReactionDTO
	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}
		return custom_err.ErrInvalidInput
	}

	//do not trust the body, trust the session
	input.UserID = session.User.ID

	//update reaction
	err = h.service.UpdateReaction(ctx, input)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *ReactionHandler) DeleteReaction(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	//check auth session
	sessionCookie, err := helper.ExtractSessionCookie(r)
	if err != nil {
		return custom_err.ErrExpiredSession
	}

	session, err := h.auth.ValidateSession(ctx, sessionCookie)
	if err != nil {
		return custom_err.ErrExpiredSession
	}

	//get input
	var input model.ReactionDTO
	r.Body = http.MaxBytesReader(w, r.Body, maxReqBodySize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return custom_err.ErrBadRequest
		}
		return custom_err.ErrInvalidInput
	}

	//do not trust the body, trust the session
	input.UserID = session.User.ID

	//delete reaction
	err = h.service.DeleteReaction(ctx, input)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
