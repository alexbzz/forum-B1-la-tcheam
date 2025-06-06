package Handler

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func SetDB(database *sql.DB) {
	db = database
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
		// Sauvegarde temporaire de l'image dans static/uploads pour affichage
		imagePath = "static/uploads/" + cookie.Value + "_profile.jpg"
		err := os.WriteFile(imagePath, profilePicture, 0644)
		if err != nil {
			fmt.Println("Erreur d'écriture de l'image :", err)
			imagePath = ""
		}
		imagePath = filepath.Base(imagePath) // pour n'envoyer que le nom du fichier à la template
	}

	tmpl, err := template.ParseFiles("templates/account.gohtml")
	if err != nil {
		http.Error(w, "Erreur template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, map[string]string{
		"Username":       cookie.Value,
		"ProfilePicture": imagePath,
	})
}

func AccountHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	fmt.Println("Cookie username =", cookie.Value)

	if r.Method == http.MethodPost {
		newUsername := r.FormValue("new_username")
		fmt.Println("Username actuel cookie:", cookie.Value, ", nouveau username formulaire:", newUsername)

		// Si un fichier est uploadé
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

			// Mise à jour de l'image dans la BDD
			res, err := db.Exec("UPDATE users SET photo_profil = ? WHERE username = ?", imageData, cookie.Value)
			if err != nil {
				http.Error(w, "Erreur mise à jour image", http.StatusInternalServerError)
				fmt.Println("Erreur DB UPDATE:", err)
				return
			}
			affected, _ := res.RowsAffected()
			if affected == 0 {
				http.Error(w, "Utilisateur introuvable pour la photo", http.StatusBadRequest)
				fmt.Println("Aucune ligne affectée pour la photo. Username utilisé:", cookie.Value)
				return
			}
			fmt.Println("Photo de profil mise à jour pour", cookie.Value)
		}

		// Mise à jour du nom d'utilisateur si rempli
		if newUsername != "" && newUsername != cookie.Value {
			res, err := db.Exec("UPDATE users SET username=? WHERE username=?", newUsername, cookie.Value)
			if err != nil {
				http.Error(w, "Erreur lors de la mise à jour du nom", http.StatusInternalServerError)
				fmt.Println("Erreur DB UPDATE username:", err)
				return
			}
			affected, _ := res.RowsAffected()
			if affected > 0 {
				http.SetCookie(w, &http.Cookie{
					Name:  "username",
					Value: newUsername,
					Path:  "/",
				})
				fmt.Println("Nom d'utilisateur changé de", cookie.Value, "à", newUsername)
			} else {
				fmt.Println("Aucune ligne affectée pour changement nom.")
			}
		} else {
			fmt.Println("Pas de changement de nom d'utilisateur")
		}

		http.Redirect(w, r, "/account", http.StatusSeeOther)
		return
	}

	ServeAccountPage(w, r)
}
