package Handler

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// PostView est une structure spécifique pour l'affichage des posts
type PostView struct {
	ID        int
	Title     string
	Content   string
	Author    string
	CreatedAt time.Time
}

type AllPostData struct {
	Username string
	Posts    []PostView // Changé de []Post à []PostView
}

func ServeAllPostPage(w http.ResponseWriter, r *http.Request) {
	data := AllPostData{
		Posts: []PostView{},
	}

	cookie, err := r.Cookie("username")
	if err == nil {
		data.Username = cookie.Value
	}

	if db == nil {
		fmt.Println("ERREUR CRITIQUE: La connexion à la base de données est nil")
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(`
  SELECT p.id, p.title, p.content, u.username, p.created_at
  FROM posts p
  JOIN users u ON p.user_id = u.id
  ORDER BY p.created_at DESC
 `)
	if err != nil {
		fmt.Println("Erreur lors de la récupération des posts:", err)
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		var username string
		var createdAtBytes []byte

		err := rows.Scan(&post.ID, &post.Title, &post.Content, &username, &createdAtBytes)
		if err != nil {
			fmt.Println("Erreur lors du scan d'un post:", err)
			continue
		}

		createdAtStr := string(createdAtBytes)
		createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			fmt.Println("Erreur lors de la conversion du timestamp:", err)
			createdAt = time.Now() // Valeur par défaut
		}

		// Création d'un PostView à partir des données récupérées
		postView := PostView{
			ID:        post.ID,
			Title:     post.Title,
			Content:   post.Content,
			Author:    username,
			CreatedAt: createdAt,
		}

		data.Posts = append(data.Posts, postView)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("Erreur lors du parcours des résultats:", err)
	}

	tmpl, err := template.ParseFiles("./forum-B1-la-tcheam/templates/AllPost.gohtml")
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
