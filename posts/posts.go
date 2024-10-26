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
	UserID    int32     `db:"user_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

func Insert() {
	db := Connect()
	post := Post{UserID: 1, Content: "This is some test content.\r\nHere is a new line.", CreatedAt: time.Now()}
	fmt.Println(post)

	r, err := db.NamedExec(`INSERT INTO post VALUES (:user_id, :content, :created_at)`, &post)
	if err != nil {
		fmt.Println("Error inserting values: ", err)
		return
	}
	fmt.Println("Successfully inserted: ", r)
}

func View() []Post {
	db := Connect()
	rows, err := db.Query(`SELECT * FROM post`)
	if err != nil {
		fmt.Println("Error fetching posts: ", err)
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var p Post

		if err := rows.Scan(&p.UserID, &p.Content, &p.CreatedAt); err != nil {
			fmt.Println("Scanning error")
		}

		posts = append(posts, p)
	}

	return posts
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
