package Handler

import (
	"fmt"
	"html/template"
	"net/http"
)

type PostPageData struct {
	Username string
	Error    string
}

type Post struct {
	ID      int
	UserID  int
	Title   string
	Content string
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	username := cookie.Value

	data := PostPageData{
		Username: username,
	}

	templatePath := "./templates/post.gohtml"
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func ServePostPage(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		fmt.Println("ERREUR CRITIQUE: La connexion à la base de données est nil")
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("username")
	if err != nil {
		fmt.Println("Utilisateur non connecté:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	username := cookie.Value
	fmt.Println("Utilisateur connecté:", username)

	title := r.FormValue("title")
	content := r.FormValue("content")
	fmt.Println("Données du formulaire - Titre:", title)
	fmt.Println("Données du formulaire - Contenu:", content)

	if title == "" || content == "" {
		fmt.Println("Validation échouée: champs vides")
		http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
		return
	}

	var testResult int
	err = db.QueryRow("SELECT 1").Scan(&testResult)
	if err != nil {
		fmt.Println("Test de connexion échoué:", err)
		http.Error(w, "La base de données n'est pas accessible", http.StatusInternalServerError)
		return
	}
	fmt.Println("Test de connexion réussi:", testResult)

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		fmt.Println("Erreur lors de la récupération de l'ID utilisateur:", err)
		http.Error(w, "Utilisateur non trouvé", http.StatusInternalServerError)
		return
	}
	fmt.Println("ID utilisateur récupéré:", userID)

	result, err := db.Exec(
		"INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)",
		userID, title, content)

	if err != nil {
		fmt.Println("Erreur SQL lors de la création du post:", err)
		http.Error(w, "Erreur lors de la création du post", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("Erreur lors de la vérification des lignes affectées:", err)
	} else if rowsAffected == 0 {
		fmt.Println("Attention: Aucune ligne n'a été insérée")
	} else {
		fmt.Println("Post créé avec succès! Lignes affectées:", rowsAffected)
	}

	http.Redirect(w, r, "/AllPost", http.StatusSeeOther)
}
