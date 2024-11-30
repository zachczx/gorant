package database

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	HostURL  string
	Port     string
	Name     string
	User     string
	Password string
	FilePath string
}

func Connect() (*sqlx.DB, error) {
	postgres := Config{
		HostURL:  os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Name:     os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}

	var db *sqlx.DB
	var err error

	pg := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", postgres.User, postgres.Password, postgres.HostURL, postgres.Port, postgres.Name)
	db, err = sqlx.Open("pgx", pg)
	if err != nil {
		fmt.Println("Error connecting to db")
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	fmt.Println("DB connected!")

	return db, nil
}

func Reset() error {
	db, err := Connect()
	if err != nil {
		return err
	}
	// Postgres
	_, err = db.Exec(`DROP TABLE IF EXISTS comments CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: comments")
		return err
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS comments_votes CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: comments_votes")
		return err
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS posts CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: posts")
		return err
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS users CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: users")
		return err
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS posts_likes CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: posts_likes")
		return err
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS tags CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: tags")
		return err
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS posts_tags CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: posts_tags")
		return err
	}

	// Users

	_, err = db.Exec(`CREATE TABLE users (user_id VARCHAR(255) PRIMARY KEY, email VARCHAR(100) NOT NULL, preferred_name VARCHAR(255) DEFAULT '', contact_me INT DEFAULT 1, avatar VARCHAR(255) DEFAULT 'default', sort_comments VARCHAR(15) DEFAULT 'upvote;desc');`)
	if err != nil {
		fmt.Println("Error creating table: users")
		return err
	}
	fmt.Println("Created table: users")

	// Posts

	_, err = db.Exec(`CREATE TABLE posts (post_id VARCHAR(255) PRIMARY KEY, post_title VARCHAR(255) NOT NULL, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, description VARCHAR(255) DEFAULT '', protected INT DEFAULT 0, created_at TEXT, mood VARCHAR(15) DEFAULT 'neutral');`)
	if err != nil {
		fmt.Println("Error creating table: posts")
		return err
	}
	fmt.Println("Created table: posts")

	// Comments

	_, err = db.Exec("CREATE TABLE comments (comment_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, content TEXT, created_at TEXT, post_id VARCHAR(255), FOREIGN KEY(post_id) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE);")
	if err != nil {
		fmt.Println("Error creating table: comments")
		return err
	}
	fmt.Println("Created table: comments")

	_, err = db.Exec(`CREATE INDEX idx_comments_post_id ON comments (post_id);`)
	if err != nil {
		fmt.Println("Error creating index: comments.post_id ")
		return err
	}
	fmt.Println("Created index: comments")

	// Posts Likes

	_, err = db.Exec(`CREATE TABLE posts_likes (like_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE, post_id VARCHAR(255) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE, score INT);`)
	if err != nil {
		fmt.Println("Error creating table: posts_likes")
		return err
	}
	fmt.Println("Created table: posts_likes")

	_, err = db.Exec(`CREATE INDEX idx_posts_likes_post_id ON posts_likes (post_id);`)
	if err != nil {
		fmt.Println("Error creating index: idx_posts_likes_post_id")
		return err
	}
	fmt.Println("Created index: idx_posts_likes_post_id")

	// Comments Votes
	_, err = db.Exec(`CREATE TABLE comments_votes (vote_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, comment_id INT REFERENCES comments(comment_id) ON DELETE CASCADE ON UPDATE CASCADE, score INT);`)
	if err != nil {
		fmt.Println("Error creating table: comments_votes")
		return err
	}
	fmt.Println("Created table: comments_votes")

	_, err = db.Exec(`CREATE INDEX idx_comments_votes_comment_id ON comments_votes (comment_id);`)
	if err != nil {
		fmt.Println("Error creating index: comments_votes.comment_id")
		return err
	}
	fmt.Println("Created index: idx_comments_votes_comment_id")

	// Tags
	_, err = db.Exec(`CREATE TABLE tags (tag_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, tag VARCHAR(30) NOT NULL);`)
	if err != nil {
		fmt.Println("Error creating table: tags")
		return err
	}
	fmt.Println("Created table: tags")

	// Intermediate Tags table
	_, err = db.Exec(`CREATE TABLE posts_tags (posts_tags_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, post_id VARCHAR(255) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE, tag_id INT REFERENCES tags(tag_id) ON DELETE CASCADE ON UPDATE CASCADE);`)
	if err != nil {
		fmt.Println("Error creating table: posts_tags")
		return err
	}
	fmt.Println("Created table: posts_tags")

	return nil
}
