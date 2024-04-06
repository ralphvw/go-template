package helpers

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ralphvw/go-template/models"
	"github.com/ralphvw/go-template/queries"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("SECRET_KEY"))

func AuthenticateUser(db *sql.DB, email string, password string) (*models.User, error) {
	query := queries.GetUserByEmail
	row := db.QueryRow(query, email)

	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.Hash, &user.FirstName, &user.LastName, &user.IsAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			LogAction("User not found " + email)
			return nil, fmt.Errorf("User not found")
		}
	}

	if !comparePasswords(password, string(user.Hash)) {
		LogAction("Authentication failed " + email)
		return nil, fmt.Errorf("Authentication Failed")
	}

	return &user, nil
}

func comparePasswords(plain, hash string) bool {
	hashBytes := []byte(hash)
	err := bcrypt.CompareHashAndPassword(hashBytes, []byte(plain))
	return err == nil
}

func CreateToken(userId int, firstName string, lastName string, email string, isAdmin bool) (string, error) {
	claims := models.Claims{
		UserID:    userId,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		IsAdmin:   isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func DecodeToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)
	return claims, nil
}
