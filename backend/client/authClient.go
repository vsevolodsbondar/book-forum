package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"forum_backend/custom_err"
	e "forum_backend/custom_err"
	"io"
	"net/http"
)

type AuthHTTPClient struct {
	BaseURL string
	Client  *http.Client
}

func (c *AuthHTTPClient) RegisterUser(ctx context.Context, dto RegisterUserRequestDTO) (RegisterUserResponseDTO, error) {
	var result RegisterUserResponseDTO

	params := AuthClientRequestParams{
		Method:         http.MethodPost,
		Path:           "/v1/register",
		RequestBody:    dto,
		ExpectedStatus: http.StatusCreated,
		RequestResult:  &result,
		SafeToRetry:    false,
	}

	err := c.doRequest(ctx, params)

	if err != nil {
		return RegisterUserResponseDTO{}, err
	}

	return result, nil
}

func (c *AuthHTTPClient) LoginUser(ctx context.Context, dto LoginUserRequestDTO) (LoginUserResponseDTO, error) {
	var result LoginUserResponseDTO

	params := AuthClientRequestParams{
		Method:         http.MethodPost,
		Path:           "/v1/login",
		RequestBody:    dto,
		ExpectedStatus: http.StatusOK,
		RequestResult:  &result,
		SafeToRetry:    true,
	}

	err := c.doRequest(ctx, params)
	if err != nil {
		return LoginUserResponseDTO{}, err
	}

	return result, nil
}

func (c *AuthHTTPClient) ValidateSession(ctx context.Context, token string) (ValidateSessionResponseDTO, error) {
	var result ValidateSessionResponseDTO

	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := AuthClientRequestParams{
		Method:         http.MethodPost,
		Path:           "/v1/session/validate",
		ExpectedStatus: http.StatusOK,
		RequestResult:  &result,
		RequestHeaders: headers,
		SafeToRetry:    true,
	}

	err := c.doRequest(ctx, params)
	if err != nil {
		return ValidateSessionResponseDTO{}, fmt.Errorf("%w: no session cookie", custom_err.ErrExpiredSession)
	}

	return result, nil
}

func (c *AuthHTTPClient) LogoutUser(ctx context.Context, token string) error {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	params := AuthClientRequestParams{
		Method:         http.MethodPost,
		Path:           "/v1/logout",
		ExpectedStatus: http.StatusNoContent,
		RequestHeaders: headers,
		SafeToRetry:    true,
	}

	err := c.doRequest(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

type ErrDoReq struct {
	Status  int
	Code    int
	Message string
}

func (c *AuthHTTPClient) doRequest(ctx context.Context, params AuthClientRequestParams) error {
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = c.doRequestOnce(ctx, params)
		if err == nil {
			return nil
		}

		//cases when we give up
		triesExceeded := attempt >= maxAttempts-1
		if !ShouldRetry(err, params.SafeToRetry) || triesExceeded {
			return err
		}

		wait := buildBackoff(attempt)
		sleepErr := sleepCtx(ctx, wait)
		if sleepErr != nil {
			return err
		}
	}
	return err
}

func (c *AuthHTTPClient) doRequestOnce(ctx context.Context, params AuthClientRequestParams) error {
	ctx, cancel := context.WithTimeout(ctx, attemptTimeout)
	defer cancel()

	var reader io.Reader

	if params.RequestBody != nil {
		jsonBody, err := json.Marshal(params.RequestBody)
		if err != nil {
			return err
		}

		reader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, params.Method, c.BaseURL+params.Path, reader)
	if err != nil {
		return err
	}

	if params.RequestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range params.RequestHeaders {
		req.Header.Set(key, value)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != params.ExpectedStatus {
		var responseError ResponseError

		if err := json.NewDecoder(resp.Body).Decode(&responseError); err != nil {
			return fmt.Errorf(
				"auth service returned status %d. %w: %w",
				resp.StatusCode,
				e.ErrJSONDecodeFailed,
				err,
			)
		}

		//that way we do not lose the statuscode
		return &AuthStatusError{
			Status: resp.StatusCode,
			Err:    mapAuthError(resp.StatusCode, responseError),
		}
	}

	if params.RequestResult != nil {
		if err := json.NewDecoder(resp.Body).Decode(params.RequestResult); err != nil {
			return &AuthStatusError{
				Status: resp.StatusCode,
				Err:    fmt.Errorf("%w: %w", e.ErrJSONDecodeFailed, err),
			}
		}
	}

	return nil
}

func (c *AuthHTTPClient) Check(ctx context.Context) error {
	params := AuthClientRequestParams{
		Method:         http.MethodGet,
		Path:           "/health/ready",
		ExpectedStatus: http.StatusOK,
	}

	err := c.doRequest(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

// function to diverse  errors for frontend
func mapAuthError(status int, resp ResponseError) error {
	msg := resp.Error.Message

	switch status {
	case http.StatusBadRequest:
		return fmt.Errorf("%w: %s", e.ErrInvalidInput, msg) // 400
	case http.StatusUnauthorized:
		if resp.Error.Code == "invalid_session" {
			return fmt.Errorf("%w: %s", e.ErrExpiredSession, msg) // 401
		}
		return fmt.Errorf("%w: %s", e.ErrWrongCredentials, msg) // 401
	case http.StatusConflict:
		return fmt.Errorf("%w: %s", e.ErrConflict, msg) // 409
	default:
		return fmt.Errorf("%w: %s", e.ErrAuthService, msg) // 500
	}
}
