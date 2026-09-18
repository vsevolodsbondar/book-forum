package client

type RegisterUserRequestDTO struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterUserResponseDTO struct {
	User UserCreatedDTO `json:"user"`
}

type UserCreatedDTO struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	CreatedAt int64  `json:"created_at"`
}

type UserDTO struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type LoginUserRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserResponsetDTO struct {
	User    UserDTO    `json:"user"`
	Session SessionDTO `json:"session"`
}

type SessionDTO struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type AuthServiceError struct {
	Error ResponseError `json:"error"`
}

type ValidateSessionResponseDTO struct {
	Session SessionDTO `json:"session"`
	User    UserDTO    `json:"user"`
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
