package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
<<<<<<< HEAD
	"os"
=======
>>>>>>> 7feb3b0 (feat(google auth))

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	_ "github.com/go-sql-driver/mysql"
)

// https://console.cloud.google.com/auth/overview for connect ID User
<<<<<<< HEAD

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
=======
var googleOAuthConfig = &oauth2.Config{
	ClientID:     "GOOGLE_CLIENT_ID",
	ClientSecret: "GOOGLE_CLIENT_SECRET",
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
>>>>>>> 7feb3b0 (feat(google auth))
	url := googleOAuthConfig.AuthCodeURL("random-state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
<<<<<<< HEAD
	var googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
=======

>>>>>>> 7feb3b0 (feat(google auth))
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := googleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Connexion DB
	db, err := sql.Open("mysql", "root:1234@tcp(127.0.0.1:3306)/forum")
	if err != nil {
		http.Error(w, "Erreur connexion DB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", userInfo.Email).Scan(&userID)
	if err == sql.ErrNoRows {

		_, err = db.Exec("INSERT INTO users (username, email, password_hash) VALUES (?, ?, NULL)", userInfo.Name, userInfo.Email)
		if err != nil {
			http.Error(w, "Erreur insertion user: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Erreur recherche user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
