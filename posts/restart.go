package posts

import (
	"fmt"

	"gorant/database"
)

func ResetDB() error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS comments;`)
	if err != nil {
		fmt.Println("Error dropping comments table")
		return err
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS comments_votes;`)
	if err != nil {
		fmt.Println("Error dropping comments_votes table")
		return err
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS posts;`)
	if err != nil {
		fmt.Println("Error dropping posts table")
		return err
	}
	_, err = db.Exec(`DROP TABLE IF EXISTS users;`)
	if err != nil {
		fmt.Println("Error dropping table table")
		return err
	}

	_, err = db.Exec(`CREATE TABLE users (user_id VARCHAR(255) PRIMARY KEY, email VARCHAR(100) NOT NULL, preferred_name VARCHAR(255) DEFAULT '', contact_me INT DEFAULT 1);`)
	if err != nil {
		fmt.Println("Error creating users table")
		return err
	}

	_, err = db.Exec(`CREATE TABLE posts (post_id VARCHAR(255) PRIMARY KEY, FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE, description VARCHAR(255) DEFAULT '', protected INT DEFAULT 0, created_at TEXT, mood VARCHAR(15) DEFAULT 'Neutral');`)
	if err != nil {
		fmt.Println("Error creating posts table")
		return err
	}

	_, err = db.Exec("CREATE TABLE comments (FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE, content TEXT, created_at TEXT, post_id VARCHAR(255), FOREIGN KEY(post_id) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE);")
	if err != nil {
		fmt.Println("Error creating comments table")
		return err
	}

	_, err = db.Exec(`CREATE INDEX idx_comments_post_id ON comments (post_id);`)
	if err != nil {
		fmt.Println("Error creating comments.post_id index")
		return err
	}

	_, err = db.Exec(`CREATE TABLE comments_votes (FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE, comment_id INT, score INT, FOREIGN KEY(comment_id) REFERENCES comments(rowid) ON DELETE CASCADE ON UPDATE CASCADE);`)
	if err != nil {
		fmt.Println("Error creating comments_votes table")
		return err
	}

	_, err = db.Exec(`CREATE INDEX idx_comments_votes_comment_id ON comments_votes (comment_id);`)
	if err != nil {
		fmt.Println("Error creating comments_votes.comment_id index")
		return err
	}

	return nil
}
