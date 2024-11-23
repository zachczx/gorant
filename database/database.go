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
	_, err = db.Exec(`DROP TABLE IF EXISTS comments cascade;`)
	if err != nil {
		fmt.Println("Error dropping table: comments")
		return err
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS comments_votes cascade;`)
	if err != nil {
		fmt.Println("Error dropping table: comments_votes")
		return err
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS posts cascade;`)
	if err != nil {
		fmt.Println("Error dropping table: posts")
		return err
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS users cascade;`)
	if err != nil {
		fmt.Println("Error dropping table: users")
		return err
	}

	_, err = db.Exec(`CREATE TABLE users (user_id VARCHAR(255) PRIMARY KEY, email VARCHAR(100) NOT NULL, preferred_name VARCHAR(255) DEFAULT '', contact_me INT DEFAULT 1, avatar VARCHAR(255) DEFAULT 'default', sort_comments VARCHAR(15) DEFAULT 'upvote;desc');`)
	if err != nil {
		fmt.Println("Error creating table: users")
		return err
	}
	fmt.Println("Created users table!")

	_, err = db.Exec(`CREATE TABLE posts (post_id VARCHAR(255) PRIMARY KEY, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, description VARCHAR(255) DEFAULT '', protected INT DEFAULT 0, created_at TEXT, mood VARCHAR(15) DEFAULT 'Neutral');`)
	if err != nil {
		fmt.Println("Error creating table: posts")
		return err
	}
	fmt.Println("Created posts table!")

	_, err = db.Exec("CREATE TABLE comments (comment_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, content TEXT, created_at TEXT, post_id VARCHAR(255), FOREIGN KEY(post_id) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE);")
	if err != nil {
		fmt.Println("Error creating table: comments")
		return err
	}
	fmt.Println("Created comments table!")

	_, err = db.Exec(`CREATE INDEX idx_comments_post_id ON comments (post_id);`)
	if err != nil {
		fmt.Println("Error creating index: comments.post_id ")
		return err
	}
	fmt.Println("Created comments index!")

	_, err = db.Exec(`CREATE TABLE comments_votes (vote_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, comment_id INT REFERENCES comments(comment_id) ON DELETE CASCADE ON UPDATE CASCADE, score INT);`)
	if err != nil {
		fmt.Println("Error creating table: comments_votes")
		return err
	}
	fmt.Println("Created comments_votes table!")

	_, err = db.Exec(`CREATE INDEX idx_comments_votes_comment_id ON comments_votes (comment_id);`)
	if err != nil {
		fmt.Println("Error creating index: comments_votes.comment_id")
		return err
	}
	fmt.Println("Created comments_votes index table!")

	return nil
}
