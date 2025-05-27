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

	http.HandleFunc("/register", Handler.RegisterHandler)
	http.HandleFunc("/", Handler.ServeRegisterPage)

	fmt.Println("Serveur lancé sur : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
