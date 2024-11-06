package posts

import "fmt"

func ResetDB() {
	db := Connect()

	_, err := db.Exec(`DROP TABLE IF EXISTS comments, comments_votes;`)
	if err != nil {
		fmt.Println("Error dropping tables")
	}

	_, err = db.Exec("CREATE TABLE comments (user_id VARCHAR(255), name VARCHAR(255), content TEXT, created_at TEXT, post_id INT);")
	if err != nil {
		fmt.Println("Error creating tables")
	}

	_, err = db.Exec(`CREATE TABLE comments_votes (user_id VARCHAR(255), comment_id INT, score INT, FOREIGN KEY(comment_id) REFERENCES comments(rowid) ON DELETE CASCADE ON UPDATE CASCADE);`)
	if err != nil {
		fmt.Println("Error creating tables")
	}
}
