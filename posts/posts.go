package posts

import "fmt"

type Post struct {
	ID        string `db:"post_id"`
	UserID    string `db:"user_id"`
	CreatedAt string `db:"created_at"`
}

func ListPosts() ([]Post, error) {
	db := Connect()

	rows, err := db.Query(`SELECT post_id, user_id, created_at FROM comments GROUP BY post_id;`)
	if err != nil {
		fmt.Println("Error executing query")
		return nil, err
	}

	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post

		if err := rows.Scan(&p.ID, &p.UserID, &p.CreatedAt); err != nil {
			fmt.Println("Error scanning")
			return nil, err
		}

		posts = append(posts, p)
		fmt.Println(p)
	}

	return posts, nil
}
