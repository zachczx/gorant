package posts

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorant/database"
	"gorant/users"

	"github.com/rezakhademix/govalidator/v2"
)

type Comment struct {
	ID            string       `db:"comment_id"`
	UserID        string       `db:"user_id"`
	Content       string       `db:"content"`
	CreatedAt     NewCreatedAt `db:"created_at"`
	PostID        string       `db:"post_id"`
	Initials      string
	PreferredName string `db:"preferred_name"`
	Avatar        string `db:"avatar"`
	CommentStats  CommentStats
	// Processed
	CreatedAtProcessed string
	AvatarPath         string
}

type CommentVote struct {
	VoteID    int `db:"vote_id"`
	UserID    int `db:"user_id"`
	CommentID int `db:"comment_id"`
	Score     int `db:"score"`
}

// Null handling for counts from DB, since counts are calculated from the query
type CommentStats struct {
	Count            sql.NullInt64 `db:"cnt"`
	CountString      string
	IDsVoted         sql.NullString `db:"cnt"`
	IDsVotedString   string         // String separated by "," with the user_ids grouped
	CurrentUserVoted string         // Returns a true or false for use in Templ template
}

func Insert(c Comment) (string, error) {
	var insertedID string

	var lastInsertID int
	err := database.DB.QueryRow(`INSERT INTO comments (user_id, content, created_at, post_id) VALUES ($1, $2, NOW(), $3) RETURNING comment_id`, c.UserID, c.Content, c.PostID).Scan(&lastInsertID)
	if err != nil {
		return insertedID, err
	}

	insertedID = strconv.Itoa(lastInsertID)
	fmt.Println("Successfully inserted!")
	return insertedID, nil
}

func ListComments(postID string, currentUser string) ([]Comment, error) {
	var comments []Comment

	// Useful resource for the join - https://stackoverflow.com/questions/2215754/sql-left-join-count
	// I considered left join for post description, but it was stupid to append description to every comment.
	// Decided to just do a separate query for that instead.
	rows, err := database.DB.Query(`SELECT comments.comment_id, comments.user_id, comments.content, comments.created_at, comments.post_id, cnt, ids_voted, users.preferred_name, users.avatar FROM comments 

							LEFT JOIN (SELECT comments_votes.comment_id, COUNT(1) AS cnt, string_agg(DISTINCT comments_votes.user_id, ',') AS ids_voted 
							FROM comments_votes 
							GROUP BY comments_votes.comment_id) AS comments_votes 
							ON comments.comment_id = comments_votes.comment_id 
							
							LEFT JOIN (SELECT users.user_id, users.preferred_name, users.avatar FROM users) as users
							ON comments.user_id = users.user_id
							WHERE comments.post_id=$1
							ORDER BY cnt DESC NULLS LAST;`, postID)
	if err != nil {
		return comments, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Comment

		if err := rows.Scan(&c.ID, &c.UserID, &c.Content, &c.CreatedAt.Time, &c.PostID, &c.CommentStats.Count, &c.CommentStats.IDsVoted, &c.PreferredName, &c.Avatar); err != nil {
			fmt.Println("Scanning error: ", err)
			return comments, err
		}

		c.Initials = strings.ToUpper(c.UserID[:2])

		c.CommentStats.CountString = NullIntToString(c.CommentStats.Count)

		if c.CommentStats.IDsVoted.Valid && currentUser != "" {
			c.CommentStats.CurrentUserVoted = strconv.FormatBool(strings.Contains(c.CommentStats.IDsVoted.String, currentUser))
		} else {
			c.CommentStats.CurrentUserVoted = "false"
		}

		c.AvatarPath = users.ChooseAvatar(c.Avatar)

		comments = append(comments, c)
	}

	return comments, nil
}

func GetComment(commentID string, currentUser string) (Comment, error) {
	var c Comment

	err := database.DB.QueryRow("SELECT * FROM comments WHERE comment_id=$1 AND user_id=$2", commentID, currentUser).Scan(&c.ID, &c.UserID, &c.Content, &c.CreatedAt.Time, &c.PostID)
	if err != nil {
		return c, err
	}

	return c, nil
}

func EditComment(commentID string, editedContent string, currentUser string) error {
	/////////////////////
	// TODO Need to add validation before saving into DB
	/////////////////////
	_, err := database.DB.Exec("UPDATE comments SET content=$1 WHERE comment_id=$2 AND user_id=$3", editedContent, commentID, currentUser)
	if err != nil {
		return err
	}

	return nil
}

func Delete(commentID string, username string) error {
	_, err := database.DB.Exec(`DELETE FROM comments WHERE comment_id=$1 AND user_id=$2`, commentID, username)
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
	res, err := database.DB.Query("SELECT comment_id FROM comments_votes WHERE comment_id=$1 AND user_id=$2", commentID, username)
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

	_, err = database.DB.Exec(q, commentID, username)
	if err != nil {
		fmt.Println("Error inserting upvote value: ", err)
		return err
	}

	return nil
}

func ConvertDate(date string) (string, error) {
	var s string
	var suffix string
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return s, err
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

	return s, err
}

func ListCommentsFilterSort(postID string, currentUser string, sort string, filter string) ([]Comment, error) {
	var comments []Comment
	var q string = `SELECT comments.comment_id, comments.user_id, comments.content, comments.created_at, comments.post_id, cnt, ids_voted, users.preferred_name, users.avatar FROM comments 

					LEFT JOIN (SELECT comments_votes.comment_id, COUNT(1) AS cnt, string_agg(DISTINCT comments_votes.user_id, ',') AS ids_voted 
					FROM comments_votes 
					GROUP BY comments_votes.comment_id) AS comments_votes 
					ON comments.comment_id = comments_votes.comment_id 
					
					LEFT JOIN (SELECT users.user_id, users.preferred_name, users.avatar FROM users) as users
					ON comments.user_id = users.user_id
					
					WHERE comments.post_id=$1 ` // Still short of ORDER BY clause, deliberate space here

	if filter != "" {
		q += `AND (comments.content ILIKE '%' || $2 || '%') `
	}

	if sort == "upvote;asc" {
		q += `ORDER BY cnt ASC NULLS FIRST;`
	} else if sort == "upvote;desc" {
		q += `ORDER BY cnt DESC NULLS LAST;`
	} else if sort == "date;asc" {
		q += `ORDER BY comments.created_at ASC NULLS LAST;`
	} else if sort == "date;desc" {
		q += `ORDER BY comments.created_at DESC NULLS LAST;`
	} else {
		q += `ORDER BY cnt DESC NULLS LAST;`
	}

	// fmt.Println(q)

	var rows *sql.Rows
	var err error
	// Useful resource for the join - https://stackoverflow.com/questions/2215754/sql-left-join-count
	// I considered left join for post description, but it was stupid to append description to every comment.
	// Decided to just do a separate query for that instead.
	if filter != "" {
		rows, err = database.DB.Query(q, postID, filter)
	} else {
		rows, err = database.DB.Query(q, postID)
	}
	if err != nil {
		return comments, err
	}
	defer rows.Close()

	for rows.Next() {
		var c Comment

		if err := rows.Scan(&c.ID, &c.UserID, &c.Content, &c.CreatedAt.Time, &c.PostID, &c.CommentStats.Count, &c.CommentStats.IDsVoted, &c.PreferredName, &c.Avatar); err != nil {
			fmt.Println("Scanning error: ", err)
			return comments, err
		}

		c.Initials = strings.ToUpper(c.UserID[:2])

		c.CommentStats.CountString = NullIntToString(c.CommentStats.Count)

		if c.CommentStats.IDsVoted.Valid && currentUser != "" {
			c.CommentStats.CurrentUserVoted = strconv.FormatBool(strings.Contains(c.CommentStats.IDsVoted.String, currentUser))
		} else {
			c.CommentStats.CurrentUserVoted = "false"
		}

		c.AvatarPath = users.ChooseAvatar(c.Avatar)

		comments = append(comments, c)
	}

	return comments, nil
}
