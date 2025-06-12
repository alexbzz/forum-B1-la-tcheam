package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	_ "github.com/go-sql-driver/mysql"
)

// https://console.cloud.google.com/auth/overview for connect ID User

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8081/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	url := googleOAuthConfig.AuthCodeURL("random-state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	var googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8081/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
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
	db, err := sql.Open("mysql", "root:123456789@tcp(127.0.0.1:3306)/forum")
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

func GithubLogin(w http.ResponseWriter, r *http.Request) {
	var githubOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8081/auth/github/callback",
		Scopes:       []string{"user:email", "user:info"},
		Endpoint:     github.Endpoint,
	}
	url := githubOAuthConfig.AuthCodeURL("random-state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GithubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	var githubOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8081/auth/github/callback",
		Scopes:       []string{"user:email", "user:info"},
		Endpoint:     github.Endpoint,
	}

	token, err := githubOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := githubOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Login string `json:"login"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("GitHub user:", userInfo.Login, userInfo.Email)

	if userInfo.Email == "" {
		respEmail, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			http.Error(w, "Failed to get emails: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer respEmail.Body.Close()

		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}

		if err := json.NewDecoder(respEmail.Body).Decode(&emails); err == nil {
			for _, e := range emails {
				if e.Primary && e.Verified {
					userInfo.Email = e.Email
					break
				}
			}
		}
	}

	fmt.Println("Email utilisé pour DB :", userInfo.Email)

	if userInfo.Email == "" {
		http.Error(w, "Impossible d'obtenir l'adresse email GitHub", http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("mysql", "root:123456789@tcp(127.0.0.1:3306)/forum")
	if err != nil {
		http.Error(w, "Erreur connexion DB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", userInfo.Email).Scan(&userID)
	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO users (username, email, password_hash) VALUES (?, ?, NULL)", userInfo.Login, userInfo.Email)
		if err != nil {
			http.Error(w, "Erreur insertion user: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println("Nouveau user GitHub enregistré :", userInfo.Email)
	} else if err != nil {
		http.Error(w, "Erreur recherche user: "+err.Error(), http.StatusInternalServerError)
		return
	} else {
		fmt.Println("User déjà existant :", userInfo.Email)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
