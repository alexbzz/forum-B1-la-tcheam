package Handler

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func validatePasswordComplexity(pwd string) error {
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

func ServeAccountPage(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var profilePicture []byte
	err = db.QueryRow("SELECT photo_profil FROM users WHERE username=?", cookie.Value).Scan(&profilePicture)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Erreur BDD", http.StatusInternalServerError)
		fmt.Println("Erreur lors de la récupération de la photo :", err)
		return
	}

	var imagePath string
	if len(profilePicture) > 0 {
		imagePath = "static/uploads/" + cookie.Value + "_profile.jpg"
		err := os.WriteFile(imagePath, profilePicture, 0644)
		if err != nil {
			fmt.Println("Erreur d'écriture de l'image :", err)
			imagePath = ""
		} else {
			imagePath = filepath.Base(imagePath)
		}
		imagePath = filepath.Base(imagePath)
	}

	// Correction du chemin du template
	tmpl, err := template.ParseFiles("./templates/account.gohtml")
	if err != nil {
		fmt.Println("Erreur template:", err)
		http.Error(w, "Erreur template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, map[string]string{
		"Username":       cookie.Value,
		"ProfilePicture": imagePath,
	})
	if err != nil {
		fmt.Println("Erreur d'exécution du template:", err)
		http.Error(w, "Erreur d'affichage: "+err.Error(), http.StatusInternalServerError)
	}
}

func AccountHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	username := cookie.Value

	if r.Method == http.MethodPost {
		newUsername := r.FormValue("new_username")
		newEmail := r.FormValue("new_email")
		oldPassword := r.FormValue("old_password")
		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_password")

		file, handler, err := r.FormFile("profile_picture")
		if err == nil {
			defer file.Close()

			ext := strings.ToLower(filepath.Ext(handler.Filename))
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
				http.Error(w, "Format d'image non supporté", http.StatusBadRequest)
				return
			}

			fmt.Println("Fichier uploadé :", handler.Filename, "avec extension", ext)

			imageData, err := io.ReadAll(file)
			if err != nil {
				http.Error(w, "Erreur lecture image", http.StatusInternalServerError)
				fmt.Println("Erreur lecture image :", err)
				return
			}
			_, err = db.Exec("UPDATE users SET photo_profil=? WHERE username=?", imageData, username)
			if err != nil {
				http.Error(w, "Erreur mise à jour image", http.StatusInternalServerError)
				return
			}
			fmt.Println("Photo mise à jour")
		}

		if newUsername != "" && newUsername != username {
			_, err := db.Exec("UPDATE users SET username=? WHERE username=?", newUsername, username)
			if err != nil {
				http.Error(w, "Erreur mise à jour nom d'utilisateur", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "username", Value: newUsername, Path: "/"})
			fmt.Println("Nom d'utilisateur mis à jour :", newUsername)
			username = newUsername
		}

		if newEmail != "" {
			_, err := db.Exec("UPDATE users SET email=? WHERE username=?", newEmail, username)
			if err != nil {
				http.Error(w, "Erreur mise à jour email", http.StatusInternalServerError)
				return
			}
			fmt.Println("Email mis à jour :", newEmail)
		}

		// Mise à jour du mot de passe avec vérification de l'ancien
		if newPassword != "" {
			if newPassword != confirmPassword {
				http.Error(w, "Les mots de passe ne correspondent pas", http.StatusBadRequest)
				return
			}
			if oldPassword == "" {
				http.Error(w, "L'ancien mot de passe est requis pour changer le mot de passe", http.StatusBadRequest)
				return
			}

			// Validation de la complexité du nouveau mot de passe
			if err := validatePasswordComplexity(newPassword); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var currentHash []byte
			err := db.QueryRow("SELECT password_hash FROM users WHERE username=?", username).Scan(&currentHash)
			if err != nil {
				http.Error(w, "Erreur récupération mot de passe", http.StatusInternalServerError)
				return
			}

			// Vérifier que l'ancien mot de passe est correct
			err = bcrypt.CompareHashAndPassword(currentHash, []byte(oldPassword))
			if err != nil {
				http.Error(w, "Ancien mot de passe incorrect", http.StatusUnauthorized)
				return
			}

			// Hacher le nouveau mot de passe
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Erreur hachage mot de passe", http.StatusInternalServerError)
				return
			}

			_, err = db.Exec("UPDATE users SET password_hash=? WHERE username=?", hashedPassword, username)
			if err != nil {
				http.Error(w, "Erreur mise à jour mot de passe", http.StatusInternalServerError)
				return
			}
			fmt.Println("Mot de passe mis à jour")
		}

		http.Redirect(w, r, "/account", http.StatusSeeOther)
		return
	}

	ServeAccountPage(w, r)
}
