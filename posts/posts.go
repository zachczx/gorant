package posts

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorant/database"
)

type Post struct {
	PostID      string `db:"post_id"`
	UserID      string `db:"user_id"`
	Description string `db:"description"`
	Protected   int    `db:"protected"`
	CreatedAt   string `db:"created_at"`
	Mood        string `db:"mood"`
}

type JoinPost struct {
	PostID        string `db:"post_id"`
	UserID        string `db:"user_id"`
	Description   string `db:"description"`
	Protected     int    `db:"protected"`
	CreatedAt     string `db:"created_at"`
	Mood          string `db:"mood"`
	PreferredName string
}

func ListPosts() ([]JoinPost, error) {
	db, err := database.Connect()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`SELECT posts.post_id, posts.user_id, posts.description, posts.protected, posts.created_at, posts.mood, users.preferred_name FROM posts
							LEFT JOIN users
							ON users.user_id = posts.user_id;`)
	if err != nil {
		fmt.Println("Error executing query: ", err)
		return nil, err
	}

	defer rows.Close()

	var posts []JoinPost

	for rows.Next() {
		var p JoinPost

		if err := rows.Scan(&p.PostID, &p.UserID, &p.Description, &p.Protected, &p.CreatedAt, &p.Mood, &p.PreferredName); err != nil {
			fmt.Println("Error scanning")
			return nil, err
		}

		posts = append(posts, p)
	}

	return posts, nil
}

func NewPost(postID string, username string) error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	exists := VerifyPostID(postID)
	if exists {
		return nil
	}

	t := time.Now().String()

	_, err = db.Exec("INSERT INTO posts (post_id, user_id, created_at) VALUES ($1, $2, $3)", postID, username, t)
	if err != nil {
		return err
	}

	return nil
}

func VerifyPostID(postID string) bool {
	db, err := database.Connect()
	if err != nil {
		return false
	}

	res, err := db.Query("SELECT post_id FROM posts WHERE post_id=$1;", postID)
	if err != nil {
		fmt.Println("Error executing query to verify post exists")
		fmt.Println(err)
	}

	defer res.Close()

	return res.Next()
}

func GetPostInfo(postID string, currentUser string) (Post, error) {
	db, err := database.Connect()
	if err != nil {
		return Post{}, err
	}
	var post Post
	if err := db.QueryRow("SELECT * FROM posts WHERE post_id=$1 AND user_id=$2", postID, currentUser).Scan(&post.PostID, &post.UserID, &post.Description, &post.Protected, &post.CreatedAt, &post.Mood); err != nil {
		return post, err
	}

	return post, nil
}

func EditPostDescription(postID string, description string) error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE posts SET description=$1 WHERE post_id=$2", description, postID)
	if err != nil {
		return err
	}
	return nil
}

func EditMood(postID string, mood string) error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	allowedMoods := [6]string{"Elated", "Happy", "Neutral", "Sad", "Upset", "Angry"}

	res := false
	for _, v := range allowedMoods {
		v = strings.ToUpper(v)

		if strings.Contains(v, strings.ToUpper(mood)) {
			res = true
			break
		}
	}
	if !res {
		err = errors.New("new mood is not in allowed list")
		return err
	}

	_, err = db.Exec("UPDATE posts SET mood=$1 WHERE post_id=$2", mood, postID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
