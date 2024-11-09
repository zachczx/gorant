package posts

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rezakhademix/govalidator/v2"
)

const (
	file string = "./starter.db"
)

type Comment struct {
	RowID     string `db:"rowid"`
	UserID    string `db:"user_id"`
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
	RowID           string `db:"rowid"`
	UserID          string `db:"user_id"`
	Content         string `db:"content"`
	CreatedAt       string `db:"created_at"`
	PostID          string `db:"post_id"`
	PostDescription string `db:"description"`
	Initials        string

	// Null handling for counts from DB, since counts are calculated from the query
	Score            sql.NullInt64 `db:"score"`
	ScoreString      string
	Count            sql.NullInt64 `db:"cnt"`
	CountString      string
	IDsVoted         sql.NullString `db:"cnt"`
	IDsVotedString   string         // String separated by "," with the user_ids grouped
	CurrentUserVoted string         // Returns a true or false for use in Templ template
}

func Insert(c Comment) error {
	db, err := Connect()
	if err != nil {
		return err
	}

	r, err := db.NamedExec(`INSERT INTO comments (user_id, content, created_at, post_id) VALUES (:user_id, :content, :created_at, :post_id)`, &c)
	if err != nil {
		fmt.Println("Error inserting values: ", err)
		return err
	}
	fmt.Println("Successfully inserted: ", r)
	return nil
}

func GetPostComments(postID string, currentUser string) (Post, []JoinComment, error) {
	db, err := Connect()
	if err != nil {
		return Post{}, nil, err
	}

	var post Post
	if err := db.QueryRow("SELECT * FROM posts WHERE post_id=?", postID).Scan(&post.PostID, &post.UserID, &post.Description, &post.Protected, &post.CreatedAt, &post.Mood); err != nil {
		fmt.Println("Queryrow issue")
		return post, nil, err
	}

	// Useful resource for the join - https://stackoverflow.com/questions/2215754/sql-left-join-count
	// I considered left join for post description, but it was stupid to append description to every comment.
	// Decided to just do a separate query for that instead.
	rows, err := db.Query(`SELECT comments.rowid, comments.user_id, comments.content, comments.created_at, comments.post_id, comments_votes.score, cnt, ids_voted FROM comments 
					LEFT JOIN (SELECT comments_votes.user_id, comments_votes.comment_id, comments_votes.score, COUNT(1) cnt, GROUP_CONCAT(comments_votes.user_id) ids_voted 
					FROM comments_votes 
					GROUP BY comments_votes.comment_id) as comments_votes 
					ON comments.rowid = comments_votes.comment_id 
					AND comments.post_id=?
					ORDER BY cnt DESC;`, postID)
	if err != nil {
		return post, nil, err
	}
	defer rows.Close()

	var comments []JoinComment

	for rows.Next() {
		var c JoinComment

		if err := rows.Scan(&c.RowID, &c.UserID, &c.Content, &c.CreatedAt, &c.PostID, &c.Score, &c.Count, &c.IDsVoted); err != nil {
			fmt.Println("Scanning error: ", err)
			return post, nil, err
		}

		c.Initials = strings.ToUpper(c.UserID[:2])

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

		if c.IDsVoted.Valid && currentUser != "" {
			c.CurrentUserVoted = strconv.FormatBool(strings.Contains(c.IDsVoted.String, currentUser))
		} else {
			c.CurrentUserVoted = "false"
		}

		c.CreatedAt = c.CreatedAt[:16]
		comments = append(comments, c)
	}

	return post, comments, nil
}

func GetComments(postID string, currentUser string) ([]JoinComment, error) {
	var comments []JoinComment
	db, err := Connect()
	if err != nil {
		return comments, err
	}

	// Useful resource for the join - https://stackoverflow.com/questions/2215754/sql-left-join-count
	// I considered left join for post description, but it was stupid to append description to every comment.
	// Decided to just do a separate query for that instead.
	rows, err := db.Query(`SELECT comments.rowid, comments.user_id, comments.content, comments.created_at, comments.post_id, comments_votes.score, cnt, ids_voted FROM comments 
					LEFT JOIN (SELECT comments_votes.user_id, comments_votes.comment_id, comments_votes.score, COUNT(1) cnt, GROUP_CONCAT(comments_votes.user_id) ids_voted 
					FROM comments_votes 
					GROUP BY comments_votes.comment_id) as comments_votes 
					ON comments.rowid = comments_votes.comment_id 
					AND comments.post_id=?
					ORDER BY cnt DESC;`, postID)
	if err != nil {
		return comments, err
	}
	defer rows.Close()

	for rows.Next() {
		var c JoinComment

		if err := rows.Scan(&c.RowID, &c.UserID, &c.Content, &c.CreatedAt, &c.PostID, &c.Score, &c.Count, &c.IDsVoted); err != nil {
			fmt.Println("Scanning error: ", err)
			return comments, err
		}

		c.Initials = strings.ToUpper(c.UserID[:2])

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

		if c.IDsVoted.Valid && currentUser != "" {
			c.CurrentUserVoted = strconv.FormatBool(strings.Contains(c.IDsVoted.String, currentUser))
		} else {
			c.CurrentUserVoted = "false"
		}

		c.CreatedAt = c.CreatedAt[:16]
		comments = append(comments, c)
	}

	return comments, nil
}

func Delete(commentID string, username string) error {
	db, err := Connect()
	if err != nil {
		return err
	}

	_, err = db.Exec(`DELETE FROM comments WHERE rowid=? AND user_id=?`, commentID, username)
	if err != nil {
		return err
	}

	return nil
}

func Connect() (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", file)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	fmt.Println("DB connected!")

	return db, nil
}

func Validate(c Comment) map[string](string) {
	v := govalidator.New()

	// v.RequiredString(c.Name, "name", "Please enter a name")
	v.RequiredString(c.Content, "content", "Please enter a message").MinString(c.Content, 10, "content", "Message needs to be at least 10 characters long.").MaxString(c.Content, 2000, "content", "Message is more than 2000 characters.")

	if v.IsFailed() {
		return v.Errors()
	}

	return nil
}

func UpVote(commentID string, username string) error {
	db, err := Connect()
	if err != nil {
		return err
	}

	res, err := db.Query("SELECT rowid FROM comments_votes WHERE comment_id=? AND user_id=?", commentID, username)
	if err != nil {
		fmt.Println("Error querying db", err)
	}
	defer res.Close()

	var q string
	if res.Next() {
		q = "DELETE FROM comments_votes WHERE comment_id=? AND user_id=?"
	} else {
		q = "INSERT INTO comments_votes (comment_id, user_id, score) VALUES (?, ?, 1)"
	}
	res.Close()

	_, err = db.Exec(q, commentID, username)
	if err != nil {
		fmt.Println("Error inserting upvote value: ", err)
		return err
	}

	return nil
}
