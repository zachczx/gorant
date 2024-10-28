package posts

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rezakhademix/govalidator/v2"
)

const (
	file string = "./starter.db"
)

type Comment struct {
	RowID     string `db:"rowid"`
	UserID    string `db:"user_id"`
	Name      string `db:"name"`
	Content   string `db:"content"`
	CreatedAt string `db:"created_at"`
	PostID    string `db:"post_id"`
}

type CommentVote struct {
	UserID    int `db:"user_id"`
	CommentID int `db:"comment_id"`
	Score     int `db:"score"`
}

func Insert(c Comment) error {
	db := Connect()

	r, err := db.NamedExec(`INSERT INTO comments (user_id, name, content, created_at, post_id) VALUES (:user_id, :name, :content, :created_at, :post_id)`, &c)
	if err != nil {
		fmt.Println("Error inserting values: ", err)
		return err
	}
	fmt.Println("Successfully inserted: ", r)
	return nil
}

func View(postID string) ([]Comment, error) {
	db := Connect()
	if postID == "" {
		postID = "demo"
	}
	rows, err := db.Query(`SELECT rowid, user_id, IFNULL(name, "[]"),content, created_at, post_id FROM comments WHERE post_id=?`, postID)
	if err != nil {
		fmt.Println("Error fetching comments: ", err)
		return nil, err
	}
	defer rows.Close()

	var comments []Comment

	for rows.Next() {
		var c Comment

		if err := rows.Scan(&c.RowID, &c.UserID, &c.Name, &c.Content, &c.CreatedAt, &c.PostID); err != nil {
			fmt.Println("Scanning error: ", err)
			return nil, err
		}

		c.CreatedAt = c.CreatedAt[:16]
		comments = append(comments, c)

	}

	return comments, nil
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

func Validate(p Comment) map[string](string) {
	v := govalidator.New()

	v.RequiredString(p.Name, "name", "Please enter a name")
	v.RequiredString(p.Content, "content", "Please enter a message").MaxString(p.Content, 2000, "content", "Max length exceeded")

	if v.IsFailed() {
		return v.Errors()
	}

	return nil
}

func UpVote(commentID string) error {
	db := Connect()

	res, err := db.Query("SELECT * FROM comments_votes WHERE comment_id=?", commentID)
	if err != nil {
		fmt.Println("Error querying db", err)
	}
	defer res.Close()

	var q string
	if res.Next() {
		q = "UPDATE comments_votes SET score=1 WHERE comment_id=? AND user_id=1"
	} else {
		q = "INSERT INTO comments_votes VALUES (1, ?, 1)"
	}
	res.Close()

	_, err = db.Exec(q, commentID)
	if err != nil {
		fmt.Println("Error inserting upvote value: ", err)
		return err
	}

	// rows, sErr := db.Query("SELECT * FROM comments_votes WHERE comment_id=?", commentID)
	// if sErr != nil {
	// 	fmt.Println("Error querying db: ", sErr)
	// 	return sErr
	// }

	// defer rows.Close()

	// var votes []CommentVote
	// for rows.Next() {
	// 	var v CommentVote

	// 	if err := res.Scan(&v.UserID, &v.CommentID, &v.Score); err != nil {
	// 		fmt.Println("Error scanning results", err)
	// 		return err
	// 	}

	// 	votes = append(votes, v)
	// }
	// fmt.Println(votes)
	return nil
}

func DownVote(commentID string) error {
	db := Connect()

	res, err := db.Query("SELECT * FROM comments_votes WHERE comment_id=?", commentID)
	if err != nil {
		fmt.Println("Error querying db", err)
	}
	defer res.Close()

	var q string
	if res.Next() {
		q = "UPDATE comments_votes SET score=-1 WHERE comment_id=? AND user_id=1"
	} else {
		q = "INSERT INTO comments_votes VALUES (1, ?, -1)"
	}
	res.Close()

	_, err = db.Exec(q, commentID)
	if err != nil {
		fmt.Println("Error inserting upvote value: ", err)
		return err
	}
	return nil
}
