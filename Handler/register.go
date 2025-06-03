package Handler

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
)

var db *sql.DB

func InitDB(database *sql.DB) {
	db = database
}

func ServeRegisterPage(w http.ResponseWriter, r *http.Request) {

	templatePath := "./forum-B1-la-tcheam/templates/register.html"

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
		fmt.Println("Erreur lors de l'insertion en base:", err)
		if err.Error() != "" && (contains(err.Error(), "Duplicate") || contains(err.Error(), "unique")) {
			http.Error(w, "Cet utilisateur ou cet email existe déjà", http.StatusConflict)
			return
		}
		http.Error(w, "Erreur d'insertion en base", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func contains(s, substr string) bool {
	return s != "" && (len(s) >= len(substr)) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || s[1:len(s)-1] != "" && contains(s[1:len(s)-1], substr))
}
