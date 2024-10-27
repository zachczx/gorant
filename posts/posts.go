package posts

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	file string = "./starter.db"
)

type Post struct {
	UserID    int32  `db:"user_id"`
	Content   string `db:"content"`
	CreatedAt string `db:"created_at"`
	Name      string `db:"name"`
	PostID    string `db:"post_id"`
}

func Insert(name string, msg string, postID string) error {
	db := Connect()
	t := time.Now().String()
	post := Post{UserID: 1, Content: msg, CreatedAt: t, Name: name, PostID: postID}
	fmt.Println(post)

	if vErr := Validate(post); vErr != nil {
		fmt.Println("Error: ", vErr)
	}

	r, err := db.NamedExec(`INSERT INTO posts VALUES (:user_id, :content, :created_at, :name, :post_id)`, &post)
	if err != nil {
		fmt.Println("Error inserting values: ", err)
		return err
	}
	fmt.Println("Successfully inserted: ", r)
	return nil
}

func View(postID string) ([]Post, error) {
	db := Connect()
	if postID == "" {
		postID = "demo"
	}
	rows, err := db.Query(`SELECT user_id, content, created_at, IFNULL(name, "[]"), post_id FROM posts WHERE post_id=?`, postID)
	if err != nil {
		fmt.Println("Error fetching posts: ", err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post

		if err := rows.Scan(&p.UserID, &p.Content, &p.CreatedAt, &p.Name, &p.PostID); err != nil {
			fmt.Println("Scanning error: ", err)
			return nil, err
		}

		p.CreatedAt = p.CreatedAt[:16]
		posts = append(posts, p)
	}

	return posts, nil
}

func Connect() *sqlx.DB {
	db, err := sqlx.Open("sqlite", file)
	if err != nil {
		fmt.Println("Error connecting to DB: ", err)
	}
	if err = db.Ping(); err != nil {
		fmt.Println("Error pinging DB: ", err)
	}
	fmt.Println("DB connected!")

	return db
}
