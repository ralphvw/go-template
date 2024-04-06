package models

type TokenBody struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
