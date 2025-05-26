package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {
	// Connexion à la base MySQL
	dsn := "root:Youyou3000.@tcp(127.0.0.1:3306)/forum"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Erreur ouverture base : ", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Erreur connexion base : ", err)
	}
	fmt.Println("Connexion réussie à la base ✅")

	// Création du router
	router := mux.NewRouter()

	// Route principale
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "Bienvenue sur le forum !")
		if err != nil {
			http.Error(w, "Erreur serveur", http.StatusInternalServerError)
			return
		}
	})

	// Démarrage du serveur HTTP sur le port 8080
	fmt.Println("Serveur démarré sur http://localhost:8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Erreur serveur HTTP : ", err)
	}
}
