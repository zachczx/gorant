package posts

import (
	"database/sql"
	"fmt"
	"strconv"

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

type JoinComment struct {
	RowID     string `db:"rowid"`
	UserID    string `db:"user_id"`
	Name      string `db:"name"`
	Content   string `db:"content"`
	CreatedAt string `db:"created_at"`
	PostID    string `db:"post_id"`

	// Null handling for counts from DB, since counts are calculated from the query
	Score       sql.NullInt64 `db:"score"`
	ScoreString string
	Count       sql.NullInt64 `db:"cnt"`
	CountString string
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

func View(postID string) ([]JoinComment, error) {
	db := Connect()
	if postID == "" {
		postID = "demo"
	}

	// Useful resource for the join - https://stackoverflow.com/questions/2215754/sql-left-join-count
	rows, err := db.Query(`SELECT comments.rowid, comments.user_id, comments.name, comments.content, comments.created_at, comments.post_id, comments_votes.score, cnt FROM comments 
					LEFT JOIN (SELECT comments_votes.user_id, comments_votes.comment_id, comments_votes.score, COUNT(1) cnt 
					FROM comments_votes 
					GROUP BY comments_votes.comment_id) as comments_votes 
					ON comments.rowid = comments_votes.comment_id 
					AND comments.post_id=?;`, postID)
	if err != nil {
		fmt.Println("Error fetching comments: ", err)
		return nil, err
	}
	defer rows.Close()

	var comments []JoinComment

	for rows.Next() {
		var c JoinComment

		if err := rows.Scan(&c.RowID, &c.UserID, &c.Name, &c.Content, &c.CreatedAt, &c.PostID, &c.Score, &c.Count); err != nil {
			fmt.Println("Scanning error: ", err)
			return nil, err
		}

		if c.Score.Valid {
			c.ScoreString = strconv.FormatInt(c.Count.Int64, 10)
		} else {
			c.ScoreString = ""
		}

		if c.Count.Valid {
			c.CountString = strconv.FormatInt(c.Count.Int64, 10)
		} else {
			c.CountString = ""
		}

		c.CreatedAt = c.CreatedAt[:16]
		comments = append(comments, c)
		fmt.Println(c.Count)
	}

	return comments, nil
}

func Delete(commentID string) error {
	db := Connect()

	_, err := db.Exec(`DELETE FROM comments WHERE rowid=? AND user_id=1`, commentID)
	if err != nil {
		return err
	}

	return nil
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

func Validate(c Comment) map[string](string) {
	v := govalidator.New()

	v.RequiredString(c.Name, "name", "Please enter a name")
	v.RequiredString(c.Content, "content", "Please enter a message").MaxString(c.Content, 2000, "content", "Max length exceeded")

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
		q = "DELETE FROM comments_votes WHERE comment_id=? AND user_id=1"
	} else {
		q = "INSERT INTO comments_votes VALUES (1, ?, 1)"
	}
	res.Close()

	_, err = db.Exec(q, commentID)
	if err != nil {
		fmt.Println("Error inserting upvote value: ", err)
		return err
	}

	return nil
}
