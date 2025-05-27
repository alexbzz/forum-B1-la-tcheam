package Handler

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"path/filepath"
)

var db *sql.DB

func InitDB(database *sql.DB) {
	db = database
}

func ServeRegisterPage(w http.ResponseWriter, r *http.Request) {
	// Chemin direct vers votre dossier templates
	// Remplacez par le chemin réel où se trouve votre dossier templates
	templatePath := filepath.Join("C:\\Users\\raphy\\GolandProjects\\forum-B1-la-tcheam\\templates", "register.html")

	// Ajoutez cette ligne pour déboguer
	fmt.Println("Tentative de charger le template:", templatePath)

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if username == "" || email == "" || password == "" {
		http.Error(w, "Champs requis manquants", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Erreur de hash", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)", username, email, hash)
	if err != nil {
		http.Error(w, "Erreur d'insertion en base", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Inscription réussie 🎉")
}
