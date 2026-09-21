package client

import "context"

type AuthInterface interface {
	RegisterUser(context.Context, RegisterUserRequestDTO) (RegisterUserResponseDTO, error)
	LoginUser(context.Context, LoginUserRequestDTO) (LoginUserResponseDTO, error)
	ValidateSession(context.Context, string) (ValidateSessionResponseDTO, error)
	LogoutUser(context.Context, string) error
}
