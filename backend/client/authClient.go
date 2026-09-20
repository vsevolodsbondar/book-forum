package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	}

	err := c.doRequest(ctx, params)
	if err != nil {
		return ValidateSessionResponseDTO{}, err
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
	}

	err := c.doRequest(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (c *AuthHTTPClient) doRequest(ctx context.Context, params AuthClientRequestParams) error {
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

		return fmt.Errorf(
			"%w: %s",
			e.ErrAuthService,
			responseError.Message,
		)
	}

	if params.RequestResult != nil {
		if err := json.NewDecoder(resp.Body).Decode(params.RequestResult); err != nil {
			return fmt.Errorf(
				"%w: %w",
				e.ErrJSONDecodeFailed,
				err,
			)
		}
	}

	return nil
}
