package model

type UserInfo struct {
	ID             int64  `json:"id"`
	UserName       string `json:"username"`
	CreatedAt      string `json:"created_at"`
	ProfilePicture string `json:"profilepic"`
	LastSeen       string `json:"last_seen"`
	Name           string `json:"name"`
	Description    string `json:"description"`
}

// DTO for creating a user (for handler and service layers)
type UserSubmission struct {
	UserName       string `json:"username"`
	ProfilePicture string `json:"profilepic"`
	Name           string `json:"name"`
	Description    string `json:"description"`
}

// for PATCH requests
type UserUpdateInfo struct {
	UserName       *string `json:"username"`
	ProfilePicture *string `json:"profilepic"`
	Name           *string `json:"name"`
	Description    *string `json:"description"`
}
