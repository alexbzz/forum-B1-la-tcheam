package Handler

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

var db *sql.DB

func InitDB(database *sql.DB) {
	db = database
}

func ServeRegisterPage(w http.ResponseWriter, r *http.Request) {
	templatePath := "./templates/register.html"

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func PasswordComplexity(pwd string) error {
	if len(pwd) < 8 {
		return fmt.Errorf("le mot de passe doit contenir au moins 8 caractères")
	}

	maj := regexp.MustCompile(`[A-Z]`)
	min := regexp.MustCompile(`[a-z]`)
	num := regexp.MustCompile(`[0-9]`)
	spec := regexp.MustCompile(`[!@#~$%^&*()+|_.,<>?/\\-]`)

	if !maj.MatchString(pwd) {
		return fmt.Errorf("le mot de passe doit contenir au moins une lettre majuscule")
	}
	if !min.MatchString(pwd) {
		return fmt.Errorf("le mot de passe doit contenir au moins une lettre minuscule")
	}
	if !num.MatchString(pwd) {
		return fmt.Errorf("le mot de passe doit contenir au moins un chiffre")
	}
	if !spec.MatchString(pwd) {
		return fmt.Errorf("le mot de passe doit contenir au moins un caractère spécial")
	}
	return nil
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if username == "" || email == "" || password == "" {
		http.Error(w, "Champs requis manquants", http.StatusBadRequest)
		return
	}

	// Validation complexité mot de passe
	if err := validatePasswordComplexity(password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Erreur de hash", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)", username, email, hash)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(strings.ToLower(errMsg), "duplicate") || strings.Contains(strings.ToLower(errMsg), "unique") {
			http.Error(w, "Cet utilisateur ou cet email existe déjà", http.StatusConflict)
			return
		}
		fmt.Println("Erreur lors de l'insertion en base:", err)
		http.Error(w, "Erreur d'insertion en base", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}
