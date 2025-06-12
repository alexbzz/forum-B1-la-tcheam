package Handler

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

type Category struct {
	ID   int
	Name string
}

type AllPostData struct {
	Username           string
	Posts              []PostView
	Categories         []Category
	SelectedCategoryID string
	SelectedFilter     string
}

// Handler pour afficher tous les posts avec filtres
func ServeAllPostPage(w http.ResponseWriter, r *http.Request) {
	data := AllPostData{
		Posts:      []PostView{},
		Categories: []Category{},
	}

	// Récupérer le nom d'utilisateur depuis les cookies
	if cookie, err := r.Cookie("username"); err == nil {
		data.Username = cookie.Value
	}

	// Vérifier la connexion à la base de données
	if db == nil {
		fmt.Println("ERREUR CRITIQUE: La connexion à la base de données est nil")
		http.Error(w, "Erreur de connexion à la base de données", http.StatusInternalServerError)
		return
	}

	// Gérer les actions like/dislike si l'utilisateur est connecté
	if r.Method == "POST" && data.Username != "" {
		err := handleReaction(w, r, data.Username)
		if err != nil {
			fmt.Println("Erreur lors du traitement de la réaction:", err)
			http.Error(w, "Erreur lors du traitement de la réaction", http.StatusInternalServerError)
			return
		} else {
			// Rediriger après traitement pour éviter la resoumission
			http.Redirect(w, r, r.URL.Path+"?"+r.URL.RawQuery, http.StatusSeeOther)
			return
		}
	}

	// Récupérer les posts avec filtres
	posts, err := getPosts(data.Username, r.URL.Query().Get("filter"))
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des posts: %v\n", err)
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return
	}
	data.Posts = posts
	data.SelectedFilter = r.URL.Query().Get("filter")

	// Récupérer les catégories
	categories, err := getCategories()
	if err != nil {
		fmt.Printf("Erreur lors de la récupération des catégories: %v\n", err)
		// Ne pas faire d'erreur fatale pour les catégories
	}
	data.Categories = categories

	// Charger et exécuter le template
	tmplPath := filepath.Join("templates", "AllPost.gohtml")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		fmt.Printf("Erreur lors du chargement du template (%s): %v\n", tmplPath, err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println("Erreur lors de l'exécution du template:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

// Handler pour afficher un post individuel
func ServePostDetail(w http.ResponseWriter, r *http.Request) {
	data := PostDetailData{
		Found: false,
	}

	// Récupérer le nom d'utilisateur depuis les cookies
	if cookie, err := r.Cookie("username"); err == nil {
		data.Username = cookie.Value
	}

	// Récupérer l'ID du post depuis l'URL
	postIDStr := r.URL.Query().Get("id")
	if postIDStr == "" {
		fmt.Println("ID du post manquant dans l'URL")
		http.Error(w, "ID du post manquant", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		fmt.Printf("ID du post invalide: %s\n", postIDStr)
		http.Error(w, "ID du post invalide", http.StatusBadRequest)
		return
	}

	// Gérer les actions like/dislike si l'utilisateur est connecté
	if r.Method == "POST" && data.Username != "" {
		err := handleReaction(w, r, data.Username)
		if err != nil {
			fmt.Println("Erreur lors du traitement de la réaction:", err)
			http.Error(w, "Erreur lors du traitement de la réaction", http.StatusInternalServerError)
			return
		} else {
			// Rediriger après traitement
			http.Redirect(w, r, fmt.Sprintf("/post?id=%d", postID), http.StatusSeeOther)
			return
		}
	}

	// Récupérer le post spécifique
	post, found, err := getPostByID(postID, data.Username)
	if err != nil {
		fmt.Printf("Erreur lors de la récupération du post ID %d: %v\n", postID, err)
		http.Error(w, "Erreur lors de la récupération du post", http.StatusInternalServerError)
		return
	}

	if !found {
		fmt.Printf("Post avec ID %d non trouvé\n", postID)
		http.Error(w, "Post non trouvé", http.StatusNotFound)
		return
	}

	data.Post = post
	data.Found = true

	// Charger et exécuter le template
	tmplPath := filepath.Join("templates", "PostDetail.gohtml")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		fmt.Printf("Erreur lors du chargement du template (%s): %v\n", tmplPath, err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println("Erreur lors de l'exécution du template:", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

// Fonction pour récupérer tous les posts avec filtres
func getPosts(username string, filter string) ([]PostView, error) {
	if db == nil {
		return nil, fmt.Errorf("connexion à la base de données non initialisée")
	}

	query := `
		SELECT p.id, p.title, p.content, u.username, p.created_at,
			COALESCE((
				SELECT COUNT(*) FROM reactions r 
				WHERE r.post_id = p.id AND r.type = 'like'
			), 0) as likes,
			COALESCE((
				SELECT COUNT(*) FROM reactions r 
				WHERE r.post_id = p.id AND r.type = 'dislike'
			), 0) as dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
	`

	args := []interface{}{}
	where := ""

	// Appliquer les filtres
	if filter == "mine" && username != "" {
		where += "u.username = ?"
		args = append(args, username)
	}
	if filter == "liked" && username != "" {
		if where != "" {
			where += " AND "
		}
		where += `EXISTS (
			SELECT 1 FROM reactions r
			WHERE r.post_id = p.id AND r.type = 'like' AND r.user_id = (SELECT id FROM users WHERE username = ?)
		)`
		args = append(args, username)
	}

	if where != "" {
		query += " WHERE " + where
	}

	// Ordre de tri
	if filter == "liked" && username != "" {
		query += " ORDER BY likes DESC, p.created_at DESC"
	} else {
		query += " ORDER BY p.created_at DESC"
	}

	fmt.Printf("Exécution de la requête: %s avec args: %v\n", query, args)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'exécution de la requête: %v", err)
	}
	defer rows.Close()

	var posts []PostView
	for rows.Next() {
		post, err := scanPost(rows, username)
		if err != nil {
			fmt.Printf("Erreur lors du scan d'un post: %v\n", err)
			continue
		}
		posts = append(posts, post)
		fmt.Printf("Post récupéré: ID=%d, Title=%s\n", post.ID, post.Title)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors du parcours des résultats: %v", err)
	}

	fmt.Printf("Total des posts récupérés: %d\n", len(posts))
	return posts, nil
}

// Fonction pour récupérer un post spécifique par ID
func getPostByID(postID int, username string) (PostView, bool, error) {
	if db == nil {
		return PostView{}, false, fmt.Errorf("connexion à la base de données non initialisée")
	}

	query := `
		SELECT p.id, p.title, p.content, u.username, p.created_at,
			COALESCE((
				SELECT COUNT(*) FROM reactions r 
				WHERE r.post_id = p.id AND r.type = 'like'
			), 0) as likes,
			COALESCE((
				SELECT COUNT(*) FROM reactions r 
				WHERE r.post_id = p.id AND r.type = 'dislike'
			), 0) as dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?
	`

	fmt.Printf("Recherche du post ID: %d\n", postID)

	row := db.QueryRow(query, postID)
	post, err := scanPostRow(row, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return PostView{}, false, nil
		}
		return PostView{}, false, fmt.Errorf("erreur lors de la récupération du post: %v", err)
	}

	fmt.Printf("Post trouvé: ID=%d, Title=%s\n", post.ID, post.Title)
	return post, true, nil
}

// Fonction pour scanner un post depuis une Row
func scanPostRow(row *sql.Row, username string) (PostView, error) {
	var post PostView
	var createdAt []byte

	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &createdAt, &post.Likes, &post.Dislikes)
	if err != nil {
		return PostView{}, err
	}

	return processPost(post, createdAt, username)
}

// Fonction pour scanner un post depuis des Rows
func scanPost(rows *sql.Rows, username string) (PostView, error) {
	var post PostView
	var createdAt []byte

	err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &createdAt, &post.Likes, &post.Dislikes)
	if err != nil {
		return PostView{}, err
	}

	return processPost(post, createdAt, username)
}

// Fonction pour traiter un post après scan
func processPost(post PostView, createdAt []byte, username string) (PostView, error) {
	// Traiter la date de création
	createdAtStr := string(createdAt)
	createdAtTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		fmt.Printf("Erreur lors de la conversion du timestamp: %v\n", err)
		post.CreatedAt = time.Now()
	} else {
		post.CreatedAt = createdAtTime
	}
	post.CreatedAtStr = post.CreatedAt.Format("02/01/2006 15:04")

	// Récupérer la réaction de l'utilisateur si connecté
	post.UserReaction = ""
	if username != "" {
		var reactionType sql.NullString
		err := db.QueryRow(`
			SELECT type FROM reactions 
			WHERE post_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)
		`, post.ID, username).Scan(&reactionType)

		if err != nil && err != sql.ErrNoRows {
			fmt.Printf("Erreur lors de la récupération de la réaction: %v\n", err)
		} else if reactionType.Valid {
			post.UserReaction = reactionType.String
		}
	}

	return post, nil
}

// Fonction pour récupérer les catégories
func getCategories() ([]Category, error) {
	if db == nil {
		return nil, fmt.Errorf("connexion à la base de données non initialisée")
	}

	query := "SELECT id, name FROM categories ORDER BY name"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des catégories: %v", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			fmt.Printf("Erreur lors du scan d'une catégorie: %v\n", err)
			continue
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors du parcours des catégories: %v", err)
	}

	return categories, nil
}

// Fonction pour gérer les réactions (like/dislike)
func handleReaction(w http.ResponseWriter, r *http.Request, username string) error {
	if db == nil {
		return fmt.Errorf("connexion à la base de données non initialisée")
	}

	postID := r.FormValue("post_id")
	action := r.FormValue("action")

	fmt.Printf("Traitement réaction: postID=%s, action=%s, user=%s\n", postID, action, username)

	if postID == "" || (action != "like" && action != "dislike" && action != "remove") {
		return fmt.Errorf("paramètres invalides: postID=%s, action=%s", postID, action)
	}

	// Vérifier que le post existe
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)", postID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification du post: %v", err)
	}
	if !exists {
		return fmt.Errorf("le post avec l'ID %s n'existe pas", postID)
	}

	// Commencer une transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("erreur début transaction: %v", err)
	}
	defer tx.Rollback()

	// Vérifier s'il y a déjà une réaction
	var existingReaction sql.NullString
	err = tx.QueryRow(`
		SELECT type FROM reactions 
		WHERE post_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)
	`, postID, username).Scan(&existingReaction)

	if action == "remove" {
		// Supprimer la réaction
		_, err = tx.Exec(`
			DELETE FROM reactions 
			WHERE post_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)
		`, postID, username)
		if err != nil {
			return fmt.Errorf("erreur lors de la suppression de la réaction: %v", err)
		}
	} else if err == nil && existingReaction.Valid {
		// Une réaction existe déjà
		if existingReaction.String == action {
			// Même réaction, ne rien faire
			return tx.Commit()
		}
		// Changer la réaction
		_, err = tx.Exec(`
			UPDATE reactions SET type = ?
			WHERE post_id = ? AND user_id = (SELECT id FROM users WHERE username = ?)
		`, action, postID, username)
		if err != nil {
			return fmt.Errorf("erreur lors de la mise à jour de la réaction: %v", err)
		}
	} else if err == sql.ErrNoRows {
		// Aucune réaction existante, en créer une nouvelle
		_, err = tx.Exec(`
			INSERT INTO reactions (post_id, user_id, type)
			VALUES (?, (SELECT id FROM users WHERE username = ?), ?)
		`, postID, username, action)
		if err != nil {
			return fmt.Errorf("erreur lors de l'insertion de la réaction: %v", err)
		}
	} else {
		return fmt.Errorf("erreur lors de la vérification de la réaction existante: %v", err)
	}

	return tx.Commit()
}
