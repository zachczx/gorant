package posts

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorant/database"

	"github.com/rezakhademix/govalidator/v2"
)

type Comment struct {
	CommentID string `db:"comment_id"`
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
	CommentID       string `db:"comment_id"`
	UserID          string `db:"user_id"`
	Content         string `db:"content"`
	CreatedAt       string `db:"created_at"`
	PostID          string `db:"post_id"`
	PostDescription string `db:"description"`
	Initials        string
	PreferredName   string `db:"preferred_name"`

	// Processed
	CreatedAtProcessed string

	// Null handling for counts from DB, since counts are calculated from the query
	Score            sql.NullInt64 `db:"score"`
	ScoreString      string
	Count            sql.NullInt64 `db:"cnt"`
	CountString      string
	IDsVoted         sql.NullString `db:"cnt"`
	IDsVotedString   string         // String separated by "," with the user_ids grouped
	CurrentUserVoted string         // Returns a true or false for use in Templ template
}

func Insert(c Comment) (string, error) {
	var insertedID string
	db, err := database.Connect()
	if err != nil {
		return insertedID, err
	}

	var lastInsertID int
	err = db.QueryRow(`INSERT INTO comments (user_id, content, created_at, post_id) VALUES ($1, $2, $3, $4) RETURNING comment_id`, c.UserID, c.Content, c.CreatedAt, c.PostID).Scan(&lastInsertID)
	if err != nil {
		return insertedID, err
	}

	// r, err := db.NamedExec(`INSERT INTO comments (user_id, content, created_at, post_id) VALUES (:user_id, :content, :created_at, :post_id)`, &c)
	// if err != nil {
	// 	fmt.Println("Error inserting values: ", err)
	// 	return insertedID, err
	// }
	// ID, err := r.LastInsertId()
	// if err != nil {
	// 	return insertedID, err
	// }
	insertedID = strconv.Itoa(lastInsertID)
	fmt.Println("Successfully inserted!")
	return insertedID, nil
}

func GetPostComments(postID string, currentUser string) (Post, []JoinComment, error) {
	db, err := database.Connect()
	if err != nil {
		return Post{}, nil, err
	}

	var post Post
	if err := db.QueryRow("SELECT * FROM posts WHERE post_id=$1", postID).Scan(&post.PostID, &post.UserID, &post.Description, &post.Protected, &post.CreatedAt, &post.Mood); err != nil {
		fmt.Println("Queryrow issue")
		return post, nil, err
	}

	// Useful resource for the join - https://stackoverflow.com/questions/2215754/sql-left-join-count
	// I considered left join for post description, but it was stupid to append description to every comment.
	// Decided to just do a separate query for that instead.
	rows, err := db.Query(`SELECT comments.comment_id, comments.user_id, comments.content, comments.created_at, comments.post_id, cnt, ids_voted, users.preferred_name FROM comments 

							LEFT JOIN (SELECT comments_votes.comment_id, COUNT(1) AS cnt, string_agg(DISTINCT comments_votes.user_id, ',') AS ids_voted 
							FROM comments_votes 
							GROUP BY comments_votes.comment_id) AS comments_votes 
							ON comments.comment_id = comments_votes.comment_id 
							
							LEFT JOIN (SELECT users.user_id, users.preferred_name FROM users) as users
							ON comments.user_id = users.user_id
							WHERE comments.post_id=$1
							ORDER BY cnt DESC NULLS LAST;`, postID)
	if err != nil {
		return post, nil, err
	}
	defer rows.Close()

	var comments []JoinComment

	for rows.Next() {
		var c JoinComment

		if err := rows.Scan(&c.CommentID, &c.UserID, &c.Content, &c.CreatedAt, &c.PostID, &c.Count, &c.IDsVoted, &c.PreferredName); err != nil {
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

		c.CreatedAtProcessed = ConvertDate(c.CreatedAt)

		comments = append(comments, c)

	}

	return post, comments, nil
}

func GetComments(postID string, currentUser string) ([]JoinComment, error) {
	var comments []JoinComment
	db, err := database.Connect()
	if err != nil {
		return comments, err
	}

	// Useful resource for the join - https://stackoverflow.com/questions/2215754/sql-left-join-count
	// I considered left join for post description, but it was stupid to append description to every comment.
	// Decided to just do a separate query for that instead.
	rows, err := db.Query(`SELECT comments.comment_id, comments.user_id, comments.content, comments.created_at, comments.post_id, cnt, ids_voted, users.preferred_name FROM comments 

							LEFT JOIN (SELECT comments_votes.comment_id, COUNT(1) AS cnt, string_agg(DISTINCT comments_votes.user_id, ',') AS ids_voted 
							FROM comments_votes 
							GROUP BY comments_votes.comment_id) AS comments_votes 
							ON comments.comment_id = comments_votes.comment_id 
							
							LEFT JOIN (SELECT users.user_id, users.preferred_name FROM users) as users
							ON comments.user_id = users.user_id
							WHERE comments.post_id=$1
							ORDER BY cnt DESC NULLS LAST;`, postID)
	if err != nil {
		return comments, err
	}
	defer rows.Close()

	for rows.Next() {
		var c JoinComment

		if err := rows.Scan(&c.CommentID, &c.UserID, &c.Content, &c.CreatedAt, &c.PostID, &c.Count, &c.IDsVoted, &c.PreferredName); err != nil {
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

		c.CreatedAtProcessed = ConvertDate(c.CreatedAt)

		comments = append(comments, c)
	}

	return comments, nil
}

func GetComment(commentID string, currentUser string) (Comment, error) {
	var c Comment

	db, err := database.Connect()
	if err != nil {
		return c, err
	}

	err = db.QueryRow("SELECT * FROM comments WHERE comment_id=$1 AND user_id=$2", commentID, currentUser).Scan(&c.CommentID, &c.UserID, &c.Content, &c.CreatedAt, &c.PostID)
	if err != nil {
		return c, err
	}

	return c, nil
}

func EditComment(commentID string, editedContent string, currentUser string) error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	/////////////////////
	// TODO Need to add validation before saving into DB
	/////////////////////
	_, err = db.Exec("UPDATE comments SET content=$1 WHERE comment_id=$2 AND user_id=$3", editedContent, commentID, currentUser)
	if err != nil {
		return err
	}

	return nil
}

func Delete(commentID string, username string) error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	_, err = db.Exec(`DELETE FROM comments WHERE comment_id=$1 AND user_id=$2`, commentID, username)
	if err != nil {
		return err
	}

	return nil
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
	db, err := database.Connect()
	if err != nil {
		return err
	}

	res, err := db.Query("SELECT comment_id FROM comments_votes WHERE comment_id=$1 AND user_id=$2", commentID, username)
	if err != nil {
		fmt.Println("Error querying db", err)
	}
	defer res.Close()

	var q string
	if res.Next() {
		q = "DELETE FROM comments_votes WHERE comment_id=$1 AND user_id=$2"
	} else {
		q = "INSERT INTO comments_votes (comment_id, user_id, score) VALUES ($1, $2, 1)"
	}
	res.Close()

	_, err = db.Exec(q, commentID, username)
	if err != nil {
		fmt.Println("Error inserting upvote value: ", err)
		return err
	}

	return nil
}

func ConvertDate(date string) string {
	var s string
	var suffix string
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return s
	}

	n := time.Now()
	diff := n.Sub(t).Hours()
	switch {
	case diff < 1:
		if n.Sub(t).Minutes() < 2 {
			suffix = " minute ago"
		} else {
			suffix = " minutes ago"
		}
		// Mins
		s = strconv.Itoa(int(n.Sub(t).Minutes())) + suffix
	case diff >= 1 && diff <= 23.99:
		if diff < 2 {
			suffix = " hour ago"
		} else {
			suffix = " hours ago"
		}
		// Hours
		s = strconv.Itoa(int(diff)) + suffix
	case diff > 23.99:
		if diff < 48 {
			suffix = " day ago"
		} else {
			suffix = " days ago"
		}
		// Days
		s = strconv.Itoa(int(n.Sub(t).Hours()/24)) + suffix
	default:
		fmt.Println("Something went wrong")
	}

	return s
}
