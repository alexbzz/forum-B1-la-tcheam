package Handler

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type PostDetailData struct {
	Username     string
	Post         PostView
	Comments     []Comment
	CommentCount int
	LikeCount    int
	UserHasLiked bool
	Found        bool
}

type PostView struct {
	ID           int
	Title        string
	Content      string
	Author       string
	CreatedAt    time.Time
	Likes        any
	Dislikes     any
	CreatedAtStr string
	UserReaction string
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

	// Test de la connexion DB
	if err := db.Ping(); err != nil {
		fmt.Printf("ERREUR: Connexion DB fermée: %v\n", err)
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	// 1. Récupération des détails du post
	postQuery := `
        SELECT p.id, p.title, p.content, u.username, p.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.id = ?
    `
	var postView PostView
	var createdAtBytes []byte

	err = db.QueryRow(postQuery, postID).Scan(
		&postView.ID,
		&postView.Title,
		&postView.Content,
		&postView.Author,
		&createdAtBytes,
	)
	if err != nil {
		fmt.Printf("Erreur lors de la récupération du post: %v\n", err)
		http.Error(w, "Post non trouvé", http.StatusNotFound)
		return
	}

	// Conversion du timestamp
	createdAtStr := string(createdAtBytes)
	postView.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		fmt.Printf("Erreur lors de la conversion du timestamp: %v\n", err)
		postView.CreatedAt = time.Now() // Valeur par défaut
	}

	data.Post = postView
	data.Found = true

	// 2. Récupération des commentaires
	commentsQuery := `
        SELECT c.id, c.post_id, u.username, c.content, c.created_at
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at ASC
    `

	rows, err := db.Query(commentsQuery, postID)
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des commentaires: %v\n", err)
		// On continue même si les commentaires ne se chargent pas
	} else {
		defer rows.Close()

		var comments []Comment
		for rows.Next() {
			var comment Comment
			var createdAt time.Time

			err := rows.Scan(&comment.ID, &comment.PostID, &comment.Author, &comment.Content, &createdAt)
			if err != nil {
				fmt.Printf("Erreur lors du scan d'un commentaire: %v\n", err)
				continue
			}

			// Formatage de la date
			comment.CreatedAt = createdAt.Format("02/01/2006 15:04")
			comments = append(comments, comment)
		}

		if err = rows.Err(); err != nil {
			fmt.Printf("Erreur lors du parcours des commentaires: %v\n", err)
		}

		data.Comments = comments
		data.CommentCount = len(comments)
	}

	// 3. Récupération du nombre de likes
	likeCountQuery := "SELECT COUNT(*) FROM likes WHERE post_id = ?"
	err = db.QueryRow(likeCountQuery, postID).Scan(&data.LikeCount)
	if err != nil {
		fmt.Printf("Erreur lors du comptage des likes: %v\n", err)
		data.LikeCount = 0
	}

	// 4. Vérifier si l'utilisateur connecté a liké ce post
	if data.Username != "" {
		var userID int
		userQuery := "SELECT id FROM users WHERE username = ?"
		err = db.QueryRow(userQuery, data.Username).Scan(&userID)
		if err == nil {
			var likeExists int
			likeCheckQuery := "SELECT COUNT(*) FROM likes WHERE post_id = ? AND user_id = ?"
			err = db.QueryRow(likeCheckQuery, postID, userID).Scan(&likeExists)
			if err == nil {
				data.UserHasLiked = likeExists > 0
			}
		}
	}

	fmt.Printf("DEBUG: Post %s chargé - %d commentaires, %d likes, utilisateur a liké: %v\n",
		postID, data.CommentCount, data.LikeCount, data.UserHasLiked)

	// 5. Chargement et exécution du template
	tmpl, err := template.ParseFiles("./forum-B1-la-tcheam/templates/postDetail.gohtml")
	if err != nil {
		fmt.Printf("Erreur lors du chargement du template: %v\n", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Printf("Erreur lors de l'exécution du template: %v\n", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}
