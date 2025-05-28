package Handler

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

func ServeLoginPage(w http.ResponseWriter, r *http.Request) {
	templatePath := filepath.Join("./templates/login.html")
	fmt.Println("Tentative de charger le template:", templatePath)

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// LoginHandler gère la soumission du formulaire de connexion
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier si la connexion à la base de données est établie
	if db == nil {
		fmt.Println("La connexion à la base de données n'est pas initialisée")
		http.Error(w, "Erreur de configuration du serveur", http.StatusInternalServerError)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Nom d'utilisateur et mot de passe requis", http.StatusBadRequest)
		return
	}

	// Récupérer l'utilisateur depuis la base de données
	var storedHash string
	var userID int
	err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username).Scan(&userID, &storedHash)
	if err != nil {
		fmt.Println("Erreur de recherche utilisateur:", err)
		http.Error(w, "Nom d'utilisateur ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	// Vérifier le mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		fmt.Println("Mot de passe incorrect pour l'utilisateur", username)
		http.Error(w, "Nom d'utilisateur ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	// Authentification réussie
	fmt.Println("Connexion réussie pour", username)
	http.Redirect(w, r, "/index", http.StatusSeeOther)
}
