package client

import "context"

type MockAuthClient struct{}

func (m *MockAuthClient) RegisterUser(ctx context.Context, dto RegisterUserRequestDTO) (RegisterUserResponseDTO, error) {
	return RegisterUserResponseDTO{
		User: UserCreatedDTO{
			ID:       2,
			Email:    dto.Email,
			Username: dto.Username,
		},
	}, nil
}

func (m *MockAuthClient) LoginUser(ctx context.Context, dto LoginUserRequestDTO) (LoginUserResponseDTO, error) {
	return LoginUserResponseDTO{
		User: UserDTO{
			ID:       2,
			Email:    dto.Email,
			Username: "Username",
		},
		Session: SessionDTO{
			ID:        "1",
			Token:     "mock-token",
			ExpiresAt: 123456,
		},
	}, nil
}

func (m *MockAuthClient) ValidateSession(ctx context.Context, token string) (ValidateSessionResponseDTO, error) {
	return ValidateSessionResponseDTO{
		Session: SessionDTO{
			ID:        "1",
			Token:     "mock-token",
			ExpiresAt: 123456,
		},
		User: UserDTO{
			ID:       2,
			Email:    "Email",
			Username: "Username",
		},
	}, nil
}

func (m *MockAuthClient) LogoutUser(ctx context.Context, token string) error {
	return nil
}
