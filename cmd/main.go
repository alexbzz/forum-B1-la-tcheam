package main

import (
	"database/sql"
	"fmt"
	"forum/Handler"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	var err error
	dsn := "root:Alexandre08@tcp(db:3306)/forum"

	// Retry de connexion à la DB
	for i := 0; i < 10; i++ {
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Tentative %d: DB non prête, nouvel essai dans 3s...", i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatal("Impossible de se connecter à la base après plusieurs tentatives :", err)
	}

	Handler.InitDB(db)

	staticDir := "C:\\Users\\alexb\\Documents\\B1forum\\static"
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

	http.HandleFunc("/", Handler.ServeIndexPage)

	fmt.Println("Serveur lancé sur : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
