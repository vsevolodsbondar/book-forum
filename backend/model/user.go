package model

type UserInfo struct {
	ID             int64  `json:"id"`
	UserName       string `json:"username"`
	CreatedAt      string `json:"createdat"`
	ProfilePicture string `json:"profilepic"`
	LastSeen       string `json:"lastseen"`
	Name           string `json:"name"`
	Description    string `json:"description"`
}

// DTO for user
type UserSubmission struct {
	UserName       string `json:"username"`
	ProfilePicture string `json:"profilepic"`
	Name           string `json:"name"`
	Description    string `json:"description"`
}
