package models

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID    int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"is_admin"`
	jwt.RegisteredClaims
}
