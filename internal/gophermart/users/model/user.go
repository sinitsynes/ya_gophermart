package model

type UserInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	ID           int32
	Login        string
	PasswordHash string
}
