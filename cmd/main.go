package main

import (
	"database/sql"
	"fmt"
	"forum/Handler"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	var err error
	dsn := "root:Youyou3000.@tcp(127.0.0.1:3306)/forum"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Erreur de connexion à la base :", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Erreur ping DB :", err)
	}

	Handler.InitDB(db)

	staticDir := "./static"

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.HandleFunc("/register", Handler.RegisterHandler)
	http.HandleFunc("/registe", Handler.ServeRegisterPage)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			Handler.LoginHandler(w, r)
		} else {
			Handler.ServeLoginPage(w, r)
		}
	})

	http.HandleFunc("/post/", Handler.ServePostDetailPage)
	http.HandleFunc("/AllPost", Handler.ServeAllPostPage)
	http.HandleFunc("/posts", Handler.ServePostPage)
	http.HandleFunc("/create-post", Handler.CreatePostHandler)
	http.HandleFunc("/", Handler.ServeIndexPage)
	///http.HandleFunc("/logout", Handler.LogoutHandler)
	fmt.Println("Serveur lancé sur : http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
