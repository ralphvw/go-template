package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ralphvw/go-template/db"
	"github.com/ralphvw/go-template/handlers"
	"github.com/ralphvw/go-template/helpers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	mux := http.NewServeMux()

	db := db.InitDb()

	fmt.Print("Server started at " + port + "\n")
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		helpers.EnableCors(w)
		helpers.LogAction("Welcome")
	})

	mux.HandleFunc("/auth/login", handlers.Login(db))
	mux.HandleFunc("/auth/signup", handlers.SignUp(db))
	mux.HandleFunc("/auth/send-reset-password-email", handlers.SendResetMail(db))
	mux.HandleFunc("/auth/reset-password", handlers.ResetPassword(db))
	mux.HandleFunc("POST /admin", handlers.CreateAdmin(db))
	mux.HandleFunc("POST /admin/login", handlers.AdminLogin(db))
	//mux.Handle("POST /reports", middleware.CheckToken(handlers.CreateReview(db))) *middleware wrapper example
	err := http.ListenAndServe(":"+port, mux)

	if err != nil {
		log.Fatal("Server error:", err)
	}
}
