package main

import (
	"database/sql"
	"fmt"
	"forum/Handler"
	"forum/auth"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	var err error
	dsn := "root:1234@tcp(127.0.0.1:3306)/forum"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Erreur de connexion à la base :", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Erreur ping DB :", err)
	}

	Handler.InitDB(db)

	staticDir := "C:\\Users\\raphy\\GolandProjects\\forum-B1-la-tcheam\\static"
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Pages Register Login
	http.HandleFunc("/register", Handler.RegisterHandler)
	http.HandleFunc("/registe", Handler.ServeRegisterPage)

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			Handler.LoginHandler(w, r)
		} else {
			Handler.ServeLoginPage(w, r)
		}
	})

	// Page d'accueil
	http.HandleFunc("/", Handler.ServeIndexPage)

	//  Routes Google Auth
	http.HandleFunc("/auth/google", auth.GoogleLogin)
	http.HandleFunc("/auth/google/callback", auth.GoogleCallback)

	http.HandleFunc("/index", Handler.ServeIndexPage)

	fmt.Println("Serveur lancé sur : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
