package posts

import (
	"fmt"
	"time"
)

type Post struct {
	PostID      string `db:"post_id"`
	UserID      string `db:"user_id"`
	Description string `db:"description"`
	Protected   int    `db:"protected"`
	CreatedAt   string `db:"created_at"`
}

func ListPosts() ([]Post, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`SELECT post_id, user_id, description, protected, created_at FROM posts;`)
	if err != nil {
		fmt.Println("Error executing query: ", err)
		return nil, err
	}

	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post

		if err := rows.Scan(&p.PostID, &p.UserID, &p.Description, &p.Protected, &p.CreatedAt); err != nil {
			fmt.Println("Error scanning")
			return nil, err
		}

		posts = append(posts, p)
		fmt.Println(p)
	}

	return posts, nil
}

func NewPost(postID string, username string) error {
	db, err := Connect()
	if err != nil {
		return err
	}

	exists := VerifyPostID(postID)
	if exists {
		return nil
	}

	t := time.Now().String()

	_, err = db.Exec("INSERT INTO posts (post_id, user_id, created_at) VALUES (?, ?, ?)", postID, username, t)
	if err != nil {
		return err
	}

	return nil
}

func VerifyPostID(postID string) bool {
	db, err := Connect()
	if err != nil {
		return false
	}

	res, err := db.Query("SELECT rowid FROM posts WHERE post_id=?", postID)
	if err != nil {
		fmt.Println("Error executing query to verify post exists")
	}

	defer res.Close()

	return res.Next()
}

func GetPostInfo(postID string, currentUser string) (Post, error) {
	db, err := Connect()
	if err != nil {
		return Post{}, err
	}
	var post Post
	if err := db.QueryRow("SELECT * FROM posts WHERE post_id=? AND user_id=?", postID, currentUser).Scan(&post.PostID, &post.UserID, &post.Description, &post.Protected, &post.CreatedAt); err != nil {
		return post, err
	}

	return post, nil
}

func EditPostDescription(postID string, description string) error {
	db, err := Connect()
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE posts SET description=? WHERE post_id=?", description, postID)
	if err != nil {
		return err
	}
	return nil
}
