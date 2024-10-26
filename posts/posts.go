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
}

func Insert(name string, msg string) error {
	db := Connect()
	t := time.Now().String()
	post := Post{UserID: 1, Content: msg, CreatedAt: t}
	fmt.Println(post)

	r, err := db.NamedExec(`INSERT INTO post VALUES (:user_id, :content, :created_at)`, &post)
	if err != nil {
		fmt.Println("Error inserting values: ", err)
		return err
	}
	fmt.Println("Successfully inserted: ", r)
	return nil
}

func View() ([]Post, error) {
	db := Connect()
	rows, err := db.Query(`SELECT * FROM post`)
	if err != nil {
		fmt.Println("Error fetching posts: ", err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post

		if err := rows.Scan(&p.UserID, &p.Content, &p.CreatedAt); err != nil {
			fmt.Println("Scanning error: ", err)
			return nil, err
		}

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
