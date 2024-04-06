package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ralphvw/go-template/constants"
	"github.com/ralphvw/go-template/helpers"
	"github.com/ralphvw/go-template/models"
	"github.com/ralphvw/go-template/queries"
	"golang.org/x/crypto/bcrypt"
)

func Login(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			helpers.HandleOptions(w, r)
			return
		}
		helpers.EnableCors(w)
		var user models.User

		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			helpers.LogAction("Error: Login Handler: " + err.Error())
			http.Error(w, "Invalid Input", http.StatusBadRequest)
			return
		}

		requiredFields := []string{"Email", "Password"}
		fieldsExist, missingFields := helpers.CheckFields(user, requiredFields)

		if !fieldsExist {
			helpers.LogAction(fmt.Sprintf("Missing fields: %v\n", missingFields))
			http.Error(w, fmt.Sprintf("Missing fields: %v\n", missingFields), http.StatusBadRequest)
			return
		}

		authenticatedUser, err := helpers.AuthenticateUser(db, strings.ToLower(user.Email), user.Password)
		if err != nil {
			helpers.LogAction("Invalid Credentials " + user.Email)
			http.Error(w, "Invalid credentials", http.StatusBadRequest)
			return
		}

		token, err := helpers.CreateToken(authenticatedUser.ID, authenticatedUser.FirstName, authenticatedUser.LastName, authenticatedUser.Email,
			authenticatedUser.IsAdmin)
		if err != nil {
			helpers.LogAction("Token creation failed")
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}

		userResponse := models.UserResponse{
			ID:        authenticatedUser.ID,
			Email:     authenticatedUser.Email,
			FirstName: authenticatedUser.FirstName,
			LastName:  authenticatedUser.LastName,
		}

		result := make(map[string]interface{})
		result["token"] = token
		result["user"] = userResponse
		message := "Login Successful"

		helpers.LogAction("Login Successful: " + authenticatedUser.Email)
		helpers.SendResponse(w, r, message, result)
	}
}

func SignUp(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			helpers.HandleOptions(w, r)
			return
		}
		helpers.EnableCors(w)
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			helpers.LogAction("Error: Signup Handler: " + err.Error())
			http.Error(w, "Invalid Input", http.StatusBadRequest)
			return
		}

		requiredFields := []string{"FirstName", "LastName", "Email", "Password"}
		fieldsExist, missingFields := helpers.CheckFields(user, requiredFields)

		if !fieldsExist {
			helpers.LogAction(fmt.Sprintf("Missing fields: %v\n", missingFields))
			http.Error(w, fmt.Sprintf("Missing fields: %v\n", missingFields), http.StatusBadRequest)
			return
		}

		plainTextPassword := user.Password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), bcrypt.DefaultCost)

		if err != nil {
			helpers.LogAction("Error: Hashing Password " + err.Error())
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}

		user.Hash = hashedPassword
		query := queries.CreateUser

		var newUser models.User

		err = db.QueryRow(query, user.FirstName, user.LastName, strings.ToLower(user.Email), user.Hash).Scan(&newUser.ID, &newUser.FirstName, &newUser.LastName, &newUser.Email)

		if err != nil {
			helpers.LogAction("Error: Failed to create user " + err.Error())
			http.Error(w, "User with this email already exists", http.StatusBadRequest)
			return
		}

		userResponse := models.UserResponse{
			ID:        newUser.ID,
			FirstName: newUser.FirstName,
			LastName:  newUser.LastName,
			Email:     newUser.Email,
		}

		message := "User created successfully"

		helpers.LogAction("User created: " + newUser.Email)
		helpers.SendResponse(w, r, message, userResponse)

	}
}

func SendResetMail(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			helpers.HandleOptions(w, r)
			return
		}
		helpers.EnableCors(w)
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			helpers.LogAction("Error: Invalid Input: Send Reset Email Handler")
			http.Error(w, "Invalid Input", http.StatusBadRequest)
			return
		}

		requiredFields := []string{"Email"}
		fieldsExist, missingFields := helpers.CheckFields(user, requiredFields)

		if !fieldsExist {
			helpers.LogAction(fmt.Sprintf("Missing fields: %v\n", missingFields))
			http.Error(w, fmt.Sprintf("Missing fields: %v\n", missingFields), http.StatusBadRequest)
			return
		}

		var res models.User

		query := queries.GetUserByEmail
		row := db.QueryRow(query, strings.ToLower(user.Email))

		e := row.Scan(&res.ID, &res.Email, &res.Hash, &res.FirstName, &res.LastName)
		if e != nil {
			if e == sql.ErrNoRows {
				helpers.LogAction("User not found: " + user.Email)
				http.Error(w, "User does not exist", http.StatusNotFound)
				return
			}
		}
		token, err := helpers.CreateToken(res.ID, res.FirstName, res.LastName, res.Email, res.IsAdmin)
		if err != nil {
			helpers.LogAction("Token creation failed")
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}

		link := "https://thesprint.vercel.app/auth/forgot-password/reset?token=" + token

		emailerr := helpers.SendMail(constants.ResetPasswordEmail(res.FirstName, link), "Reset Password", res.Email, res.FirstName)

		if emailerr != nil {
			http.Error(w, "Email could not be sent", http.StatusInternalServerError)
			return
		}

		message := "Reset Password Email Sent"

		helpers.LogAction("Reset password email sent: " + res.Email)

		helpers.SendResponse(w, r, message, res.Email)

	}
}

func ResetPassword(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			helpers.HandleOptions(w, r)
			return
		}
		helpers.EnableCors(w)
		var body models.TokenBody
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			helpers.LogAction("Error: Reset Password: " + err.Error())
			http.Error(w, "Invalid Input", http.StatusBadRequest)
			return
		}
		helpers.LogAction(fmt.Sprintf("Body: %v", body))

		requiredFields := []string{"Token", "Password"}
		fieldsExist, missingFields := helpers.CheckFields(body, requiredFields)

		if !fieldsExist {
			helpers.LogAction(fmt.Sprintf("Missing fields: %v\n", missingFields))
			http.Error(w, fmt.Sprintf("Missing fields: %v\n", missingFields), http.StatusBadRequest)
			return
		}
		token := body.Token
		password := body.Password
		claims, err := helpers.DecodeToken(token)
		if err != nil {
			helpers.LogAction("Token decoding failed: " + err.Error())
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}

		helpers.LogAction(fmt.Sprintf("Claims: %v", claims))
		email := claims["email"]
		var formattedEmail string
		if str, ok := email.(string); ok {
			formattedEmail = strings.ToLower(str)
		} else {
			helpers.LogAction("Email is not a string")
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}
		var returnedEmail string
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		helpers.LogAction(fmt.Sprintf("Formatted Email: %v", formattedEmail))
		e := db.QueryRow(queries.ResetPassword, hashedPassword, formattedEmail).Scan(&returnedEmail)
		helpers.LogAction(fmt.Sprintf("Email: %v", returnedEmail))
		if e != nil {
			if e == sql.ErrNoRows {
				helpers.LogAction("User not found " + formattedEmail)
				http.Error(w, "User does not exist", http.StatusNotFound)
				return
			}
		}

		result := make(map[string]interface{})
		result["email"] = email
		message := "Password reset successfully"

		helpers.LogAction("Password reset: " + returnedEmail)

		helpers.SendResponse(w, r, message, result)
	}

}
