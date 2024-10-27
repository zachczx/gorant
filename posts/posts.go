package posts

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rezakhademix/govalidator/v2"
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

func Insert(post Post) error {
	db := Connect()
	fmt.Println(post)

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

func Validate(p Post) map[string](string) {
	v := govalidator.New()

	v.RequiredString(p.Name, "name", "Please enter a name")
	v.RequiredString(p.Content, "content", "Please enter a message").MaxString(p.Content, 2000, "content", "Max length exceeded")

	if v.IsFailed() {
		return v.Errors()
	}

	return nil
}
