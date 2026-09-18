package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type AuthInterface interface {
	RegisterUser(context.Context, RegisterUserRequestDTO) (RegisterUserResponseDTO, error)
	LoginUser(context.Context, LoginUserRequestDTO) (LoginUserResponsetDTO, error)
	ValidateSession(context.Context, string) (ValidateSessionResponseDTO, error)
	LogoutUser(context.Context) error
}

type MockAuthInterface interface {
	RegisterUser(context.Context, RegisterUserRequestDTO) (RegisterUserResponseDTO, error)
	LoginUser(context.Context, LoginUserRequestDTO) (LoginUserResponsetDTO, error)
	ValidateSession(context.Context, string) (ValidateSessionResponseDTO, error)
	LogoutUser(context.Context) error
}

type AuthHTTPClient struct {
	BaseUrl string
	Client  *http.Client
}

func (c *AuthHTTPClient) RegisterUser(ctx context.Context, dto RegisterUserRequestDTO) (RegisterUserResponseDTO, error) {
	body, err := json.Marshal(dto)
	if err != nil {
		return RegisterUserResponseDTO{}, err
	}

	//where
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseUrl+"/v1/register",
		bytes.NewReader(body), //here dto
	)
	if err != nil {
		return RegisterUserResponseDTO{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	//execute and reaction on response
	resp, err := c.Client.Do(req)
	if err != nil {
		return RegisterUserResponseDTO{}, err
	}
	defer resp.Body.Close()

	var responseError ResponseError
	var userData RegisterUserResponseDTO

	if resp.StatusCode != http.StatusCreated {
		if err := json.NewDecoder(resp.Body).Decode(&responseError); err != nil {
			return RegisterUserResponseDTO{}, fmt.Errorf("error in auth %s", responseError.Message)
		}
	} else {
		if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
			return RegisterUserResponseDTO{}, err
		}
	}

	return userData, nil
}

func (c *AuthHTTPClient) ValidateSession(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseUrl+"/v1/session/validate",
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token is invalid")
	}

	return nil
}
