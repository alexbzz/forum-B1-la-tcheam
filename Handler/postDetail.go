package Handler

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type PostDetailData struct {
	Username string
	Post     PostView
}

func ServePostDetailPage(w http.ResponseWriter, r *http.Request) {

	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "ID du post manquant", http.StatusBadRequest)
		return
	}

	data := PostDetailData{}
	cookie, err := r.Cookie("username")
	if err == nil {
		data.Username = cookie.Value
	}

	if db == nil {
		fmt.Println("ERREUR CRITIQUE: La connexion à la base de données est nil")
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	query := `
		SELECT p.id, p.title, p.content, u.username, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?
	`
	var postView PostView
	var createdAtBytes []byte

	err = db.QueryRow(query, postID).Scan(
		&postView.ID,
		&postView.Title,
		&postView.Content,
		&postView.Author,
		&createdAtBytes,
	)
	if err != nil {
		fmt.Println("Erreur lors de la récupération du post:", err)
		http.Error(w, "Post non trouvé", http.StatusNotFound)
		return
	}

	createdAtStr := string(createdAtBytes)
	postView.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		fmt.Println("Erreur lors de la conversion du timestamp:", err)
		postView.CreatedAt = time.Now() // Valeur par défaut
	}

	data.Post = postView

	tmpl, err := template.ParseFiles("./forum-B1-la-tcheam/templates/postDetail.gohtml")
	if err != nil {
		fmt.Println("Erreur lors du chargement du template:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println("Erreur lors de l'exécution du template:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}
