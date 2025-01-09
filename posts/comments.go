package posts

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorant/database"
	"gorant/upload"
	"gorant/users"

	"github.com/google/uuid"
	"github.com/rezakhademix/govalidator/v2"
	"github.com/sym01/htmlsanitizer"
)

type Comment struct {
	ID                 uuid.UUID    `db:"comment_id"`
	UserID             string       `db:"user_id"`
	Content            string       `db:"content"`
	CreatedAt          NewCreatedAt `db:"created_at"`
	PostID             string       `db:"post_id"`
	Initials           string
	PreferredName      string `db:"preferred_name"`
	Avatar             string `db:"avatar"`
	CommentStats       CommentStats
	CreatedAtProcessed string
	AvatarPath         string
	File               upload.LookupFile
	NullFile           upload.NullFile
}

func (c *Comment) IDString() string {
	return c.ID.String()
}

func (c *Comment) FileURL() string {
	k := c.File.BaseURL + "/" + c.File.ID.String() + "-" + c.File.Key
	return k
}

type CommentVote struct {
	VoteID    uuid.UUID `db:"vote_id"`
	UserID    string    `db:"user_id"`
	CommentID uuid.UUID `db:"comment_id"`
	Score     int       `db:"score"`
}

// Null handling for counts from DB, since counts are calculated from the query
type CommentStats struct {
	Count            sql.NullInt64 `db:"cnt"`
	CountString      string
	IDsVoted         sql.NullString `db:"cnt"`
	IDsVotedString   string         // String separated by "," with the user_ids grouped
	CurrentUserVoted string         // Returns a true or false for use in Templ template
}

func sanitizeHTML(HTML string) (string, error) {
	sanitizer := htmlsanitizer.NewHTMLSanitizer()
	sanitizer.AllowList.Tags = allowedHTMLTags
	sHTML, err := sanitizer.SanitizeString(HTML)
	if err != nil {
		return "", err
	}
	return sHTML, err
}

var allowedHTMLTags = []*htmlsanitizer.Tag{
	{"a", nil, []string{"href"}},
	{"h1", []string{"style"}, []string{}},
	{"h2", []string{"style"}, []string{}},
	{"li", []string{"style"}, []string{}},
	{"strong", []string{}, []string{}},
	{"li", []string{"style"}, []string{}},
	{"ol", []string{"style"}, []string{}},
	{"p", []string{"style"}, []string{}},
	{"ul", []string{"style"}, []string{}},
	{"b", []string{}, []string{}},
	{"span", []string{"style"}, []string{}},
	{"i", []string{}, []string{}},
	{"u", []string{}, []string{}},
}

func Insert(c Comment) (string, error) {
	var returnCommentID uuid.UUID
	var insertedCommentID string
	var err error
	c.Content, err = sanitizeHTML(c.Content)
	if err != nil {
		fmt.Println(err)
	}
	// Insert into DB record if there's no uploaded file. By this time, the upload would have completed successfully.
	if c.File.File == nil && c.File.Key == "" {
		err := database.DB.QueryRow(`INSERT INTO comments (user_id, content, created_at, post_id) VALUES ($1, $2, NOW(), $3) RETURNING comment_id`, c.UserID, c.Content, c.PostID).Scan(&returnCommentID)
		if err != nil {
			return insertedCommentID, err
		}
		insertedCommentID = returnCommentID.String()
		return insertedCommentID, err
	}

	_, err = database.DB.Exec(`INSERT INTO files (file_id, user_id, file_key, file_thumbnail_key, file_store, file_bucket, file_base_url, uploaded_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`, c.File.ID, c.UserID, c.File.Key, c.File.ThumbnailKey, c.File.Store, c.File.Bucket, c.File.BaseURL, time.Now())
	if err != nil {
		return insertedCommentID, err
	}
	err = database.DB.QueryRow(`INSERT INTO comments (user_id, content, created_at, post_id, file_id) VALUES ($1, $2, NOW(), $3, $4) RETURNING comment_id`, c.UserID, c.Content, c.PostID, c.File.ID).Scan(&returnCommentID)
	if err != nil {
		return insertedCommentID, err
	}

	insertedCommentID = returnCommentID.String()
	fmt.Println("Successfully inserted!")

	return insertedCommentID, nil
}

func ListComments(postID string, currentUser string) ([]Comment, error) {
	var comments []Comment

	// Useful resource for the join - https://stackoverflow.com/questions/2215754/sql-left-join-count
	// I considered left join for post description, but it was stupid to append description to every comment.
	// Decided to just do a separate query for that instead.
	//
	// This is not being used, everything is using FilterSort variant.
	rows, err := database.DB.Query(`SELECT comments.comment_id, comments.user_id, comments.content, comments.created_at, comments.post_id, comments.file_id, files.file_key, files.file_thumbnail_key, files.file_store, files.file_bucket, cnt, ids_voted, users.preferred_name, users.avatar FROM comments

									LEFT JOIN files
									ON comments.file_id=files.file_id

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

		if err := rows.Scan(&c.ID, &c.UserID, &c.Content, &c.CreatedAt.Time, &c.PostID, &c.NullFile.ID, &c.NullFile.Key, &c.NullFile.ThumbnailKey, &c.NullFile.Store, &c.NullFile.Bucket, &c.CommentStats.Count, &c.CommentStats.IDsVoted, &c.PreferredName, &c.Avatar); err != nil {
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
		if c.NullFile.ID.Valid {
			c.File.ID = c.NullFile.ID.UUID
			c.File.Key = c.NullFile.Key.String
			c.File.ThumbnailKey = c.NullFile.ThumbnailKey.String
			c.File.Store = c.NullFile.Store.String
			c.File.Bucket = c.NullFile.Bucket.String
		}
		c.AvatarPath = users.ChooseAvatar(c.Avatar)

		comments = append(comments, c)
	}
	return comments, nil
}

func GetComment(commentID string, currentUser string) (Comment, error) {
	var c Comment

	// Protection is at 2 levels, if I wonder again in future.
	// 1) The UI conditional doesn't show the edit comment link if the comment owner is not currentUser.
	// 2) This SQL returns nothing found if the currentUser is not the owner of the comment.
	err := database.DB.QueryRow(`SELECT comments.comment_id, comments.user_id, comments.content, comments.created_at, comments.post_id, files.file_id, files.file_key, files.file_thumbnail_key, files.file_store, files.file_bucket, files.file_base_url
								FROM(SELECT *
									FROM comments
									WHERE comment_id=$1 AND user_id=$2) as comments
									LEFT JOIN files ON files.file_id=comments.file_id;`, commentID, currentUser).Scan(&c.ID, &c.UserID, &c.Content, &c.CreatedAt.Time, &c.PostID, &c.NullFile.ID, &c.NullFile.Key, &c.NullFile.ThumbnailKey, &c.NullFile.Store, &c.NullFile.Bucket, &c.NullFile.BaseURL)
	if err != nil {
		return c, err
	}
	if c.NullFile.ID.Valid {
		c.File.ID = c.NullFile.ID.UUID
		c.File.Key = c.NullFile.Key.String
		c.File.ThumbnailKey = c.NullFile.ThumbnailKey.String
		c.File.Store = c.NullFile.Store.String
		c.File.Bucket = c.NullFile.Bucket.String
		c.File.BaseURL = c.NullFile.BaseURL.String
	}
	return c, nil
}

func EditComment(c Comment) error {
	var err error
	var returnID uuid.UUID
	sanitizedHTML, err := sanitizeHTML(c.Content)
	if err != nil {
		return err
	}

	if c.File.File == nil && c.File.Key == "" {
		_, err = database.DB.Exec(`UPDATE comments SET content=$1 WHERE comment_id=$2 AND user_id=$3`, sanitizedHTML, c.ID, c.UserID)
		if err != nil {
			return err
		}
		fmt.Println("Successfully edited!")
		return nil
	}
	err = database.DB.QueryRow(`INSERT INTO files (user_id, file_key, file_thumbnail_key, file_store, file_bucket, file_base_url, uploaded_at) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING file_id`, c.UserID, c.File.Key, c.File.ThumbnailKey, c.File.Store, c.File.Bucket, c.File.BaseURL, time.Now()).Scan(&returnID)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec(`UPDATE comments SET content=$1, file_id=$2 WHERE comment_id=$3 AND user_id=$4`, sanitizedHTML, returnID.String(), c.ID, c.UserID)
	if err != nil {
		return err
	}

	fmt.Println("Successfully edited and updated files table!")
	return nil

	////////////////////////////////////////////////////////////
	// Insert into DB record if there's no uploaded file. By this time, the upload would have completed successfully.
	/* if c.File.File == nil && c.File.Key == "" {
		err := database.DB.QueryRow(`INSERT INTO comments (user_id, content, created_at, post_id) VALUES ($1, $2, NOW(), $3) RETURNING comment_id`, c.UserID, c.Content, c.PostID).Scan(&returnCommentID)
		if err != nil {
			return insertedCommentID, err
		}
		insertedCommentID = returnCommentID.String()
		return insertedCommentID, err
	}

	err = database.DB.QueryRow(`INSERT INTO files (file_key, file_store, file_bucket) VALUES($1, $2, $3) RETURNING file_id`, c.File.Key, c.File.Store, c.File.Bucket).Scan(&returnID)
	if err != nil {
		return insertedCommentID, err
	}
	err = database.DB.QueryRow(`INSERT INTO comments (user_id, content, created_at, post_id, file_id) VALUES ($1, $2, NOW(), $3, $4) RETURNING comment_id`, c.UserID, c.Content, c.PostID, returnID.String()).Scan(&returnCommentID)
	if err != nil {
		return insertedCommentID, err
	}

	insertedCommentID = returnCommentID.String()
	fmt.Println("Successfully inserted!") */
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
	var q string = `SELECT comments.comment_id, comments.user_id, comments.content, comments.created_at, comments.post_id, comments.file_id, files.file_key, files.file_thumbnail_key, files.file_store, files.file_bucket, files.file_base_url, cnt, ids_voted, users.preferred_name, users.avatar FROM comments 

					LEFT JOIN files
					ON comments.file_id=files.file_id

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

		if err := rows.Scan(&c.ID, &c.UserID, &c.Content, &c.CreatedAt.Time, &c.PostID, &c.NullFile.ID, &c.NullFile.Key, &c.NullFile.ThumbnailKey, &c.NullFile.Store, &c.NullFile.Bucket, &c.NullFile.BaseURL, &c.CommentStats.Count, &c.CommentStats.IDsVoted, &c.PreferredName, &c.Avatar); err != nil {
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
		if c.NullFile.ID.Valid {
			c.File.ID = c.NullFile.ID.UUID
			c.File.Key = c.NullFile.Key.String
			c.File.ThumbnailKey = c.NullFile.ThumbnailKey.String
			c.File.Store = c.NullFile.Store.String
			c.File.Bucket = c.NullFile.Bucket.String
			c.File.BaseURL = c.NullFile.BaseURL.String
		}
		c.AvatarPath = users.ChooseAvatar(c.Avatar)

		comments = append(comments, c)
	}

	return comments, nil
}
