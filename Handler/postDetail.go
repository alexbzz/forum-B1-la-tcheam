package Handler

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"
)

var db *sql.DB

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

type Comment struct {
	ID        int
	PostID    int
	Author    string
	Content   string
	CreatedAt string
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
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	if err := db.Ping(); err != nil {
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	var postView PostView
	var createdAtBytes []byte

	postQuery := `
        SELECT p.id, p.title, p.content, u.username, p.created_at,
               (SELECT COUNT(*) FROM likes WHERE post_id = p.id AND reaction_type = 'like') as likes,
               (SELECT COUNT(*) FROM likes WHERE post_id = p.id AND reaction_type = 'dislike') as dislikes
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.id = ?
    `

	err = db.QueryRow(postQuery, postID).Scan(
		&postView.ID,
		&postView.Title,
		&postView.Content,
		&postView.Author,
		&createdAtBytes,
		&postView.Likes,
		&postView.Dislikes,
	)

	if err != nil {
		http.Error(w, "Post non trouvé", http.StatusNotFound)
		return
	}

	createdAtStr := string(createdAtBytes)
	postView.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		postView.CreatedAt = time.Now()
	}

	data.Post = postView
	data.Found = true

	comments, err := getCommentsForPost(postID)
	if err != nil {
		data.Comments = []Comment{}
	} else {
		data.Comments = comments
		data.CommentCount = len(comments)
	}

	likeCountQuery := "SELECT COUNT(*) FROM likes WHERE post_id = ?"
	err = db.QueryRow(likeCountQuery, postID).Scan(&data.LikeCount)
	if err != nil {
		data.LikeCount = 0
	}

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

	tmpl, err := template.ParseFiles("templates/postDetail.gohtml")
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func getCommentsForPost(postID string) ([]Comment, error) {
	commentsQuery := `
        SELECT c.id, c.post_id, u.username, c.content, c.created_at
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at ASC
    `

	rows, err := db.Query(commentsQuery, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var createdAt time.Time

		err := rows.Scan(&comment.ID, &comment.PostID, &comment.Author, &comment.Content, &createdAt)
		if err != nil {
			continue
		}

		comment.CreatedAt = createdAt.Format("02/01/2006 15:04")
		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

func AddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("username")
	if err != nil {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	postID := r.FormValue("post_id")
	content := r.FormValue("content")

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username = ?", cookie.Value).Scan(&userID)
	if err != nil {
		http.Error(w, "Utilisateur non trouvé", http.StatusUnauthorized)
		return
	}

	_, err = db.Exec(
		"INSERT INTO comments (post_id, user_id, content, created_at) VALUES (?, ?, ?, NOW())",
		postID, userID, content,
	)
	if err != nil {
		http.Error(w, "Erreur lors de l'insertion", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/post-detail?id="+postID, http.StatusSeeOther)
}
