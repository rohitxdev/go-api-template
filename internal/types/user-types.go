package types

type UserCore struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

type User struct {
	UserCore
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
