package services

import (
	"database/sql"

	"github.com/ralphvw/go-template/models"
	"github.com/ralphvw/go-template/queries"
)

func CreateAdmin(db *sql.DB, user *models.User) error {
	_, err := db.Exec(queries.CreateAdmin, &user.FirstName, &user.LastName, &user.Email, &user.Hash)

	return err
}

func GetAllUsers(db *sql.DB, searchTerm string) (*[]models.UserResponse, error) {
	var users []models.UserResponse

	rows, err := db.Query(queries.GetAllUsers, searchTerm)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var user models.UserResponse
		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return &users, nil
}
