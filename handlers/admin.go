package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ralphvw/go-template/helpers"
	"github.com/ralphvw/go-template/models"
	"github.com/ralphvw/go-template/services"
	"golang.org/x/crypto/bcrypt"
)

func CreateAdmin(db *sql.DB) http.HandlerFunc {
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

		queryerr := services.CreateAdmin(db, &user)

		if queryerr != nil {
			helpers.LogAction("Error running create admin query " + queryerr.Error())
			http.Error(w, "User already exists ", http.StatusInternalServerError)
			return
		}

		message := "Admin created successfully"

		helpers.LogAction("New admin created: " + user.Email)

		helpers.SendResponse(w, r, message, nil)

	}
}

func AdminLogin(db *sql.DB) http.HandlerFunc {

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

		if !authenticatedUser.IsAdmin {
			helpers.LogAction("Unauthorised admin access: " + user.Email)
			http.Error(w, "Unauthorised", http.StatusUnauthorized)
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

func GetUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			helpers.HandleOptions(w, r)
			return
		}
		helpers.EnableCors(w)

		claims, ok := r.Context().Value("userClaims").(map[string]interface{})
		if !ok {
			helpers.LogAction("Error extracting user claims")
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}

		isAdmin, ok := claims["is_admin"].(bool)

		if !ok {
			helpers.LogAction("Wrong role type assertion")
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}

		userEmail, ok := claims["email"].(string)

		if !ok {
			helpers.LogAction("Wrong user email type assertion")
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}

		if !isAdmin {
			helpers.LogAction("Attempt to get all users with unauthorised access: " + userEmail)
			http.Error(w, "Unauthorised", http.StatusUnauthorized)
			return
		}

		search := r.URL.Query().Get("search")
		searchTerm := "%" + search + "%"

		users, queryErr := services.GetAllUsers(db, searchTerm)
		if queryErr != nil {
			helpers.LogAction("Error fetching users: " + queryErr.Error())
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		message := "Users fetched"
		helpers.LogAction("Users fetched by: " + userEmail)
		helpers.SendResponse(w, r, message, users)
	}
}
