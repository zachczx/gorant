package posts

import "fmt"

func ResetDB() error {
	db := Connect()

	_, err := db.Exec(`DROP TABLE IF EXISTS comments;`)
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

	_, err = db.Exec(`CREATE TABLE posts (post_id VARCHAR(255) PRIMARY KEY, user_id VARCHAR(255), description VARCHAR(255) DEFAULT '', protected INT DEFAULT 0, created_at TEXT);`)
	if err != nil {
		fmt.Println("Error creating posts table")
		return err
	}

	_, err = db.Exec("CREATE TABLE comments (user_id VARCHAR(255), content TEXT, created_at TEXT, post_id VARCHAR(255), FOREIGN KEY(post_id) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE);")
	if err != nil {
		fmt.Println("Error creating comments table")
		return err
	}

	_, err = db.Exec(`CREATE INDEX idx_comments_post_id ON comments (post_id);`)
	if err != nil {
		fmt.Println("Error creating comments.post_id index")
		return err
	}

	_, err = db.Exec(`CREATE TABLE comments_votes (user_id VARCHAR(255), comment_id INT, score INT, FOREIGN KEY(comment_id) REFERENCES comments(rowid) ON DELETE CASCADE ON UPDATE CASCADE);`)
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
