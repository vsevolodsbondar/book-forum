package model

type UserInfo struct {
	id             uint
	UserName       string
	CreatedAt      string
	ProfilePicture string
	LastSeen       string
	Name           string
	Description    string
}

// DTO for user
type UserSubmission struct {
	UserName       string
	ProfilePicture string
	Name           string
	Description    string
}
