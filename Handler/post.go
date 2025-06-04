package Handler

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type PostPageData struct {
	Username string
	Error    string
}

type Post struct {
	ID        int
	UserID    int
	Title     string
	Content   string
	CreatedAt time.Time
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

	templatePath := "./forum-B1-la-tcheam/templates/post.gohtml"
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Erreur de template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func ServePostPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	username := cookie.Value

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		fmt.Println("Erreur lors de la récupération de l'ID utilisateur:", err)
		http.Error(w, "Utilisateur non trouvé", http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(
		"INSERT INTO posts (user_id, title, content, created_at) VALUES (?, ?, ?, ?)",
		userID, title, content, time.Now())

	if err != nil {
		fmt.Println("Erreur lors de la création du post:", err)
		http.Error(w, "Erreur lors de la création du post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
