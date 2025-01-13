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
	ID                 uuid.UUID `db:"comment_id"`
	UserID             string    `db:"user_id"`
	Content            string    `db:"content"`
	CreatedAt          CreatedAt `db:"created_at"`
	PostID             string    `db:"post_id"`
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

func (c *Comment) ThumbnailURL() string {
	// Trying to guess the thumbnail just in case something went wrong (e.g. didn't save successfully or accidentally deleted)
	if c.File.ThumbnailKey == "" {
		var fn []string
		if strings.Contains(c.File.Key, ".") {
			fn = strings.Split(c.File.Key, ".")
		}
		return c.File.BaseURL + "/" + c.File.ID.String() + "-" + fn[0] + "-tn." + fn[1]
	}
	return c.File.BaseURL + "/" + c.File.ID.String() + "-" + c.File.ThumbnailKey
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

type SearchComment struct {
	ID                 uuid.UUID `db:"comment_id"`
	UserID             string    `db:"user_id"`
	Content            string    `db:"content"`
	CreatedAt          CreatedAt `db:"created_at"`
	PostID             string    `db:"post_id"`
	PostTitle          string    `db:"post_title"`
	Initials           string
	PreferredName      string `db:"preferred_name"`
	Avatar             string `db:"avatar"`
	CommentStats       CommentStats
	CreatedAtProcessed string
	AvatarPath         string
	File               upload.LookupFile
	NullFile           upload.NullFile
}

func (c *SearchComment) IDString() string {
	return c.ID.String()
}

func sanitizeHTML(HTML string) (string, error) {
	sanitizer := htmlsanitizer.NewHTMLSanitizer()
	sanitizer.AllowList.Tags = allowedHTMLTags
	sHTML, err := sanitizer.SanitizeString(HTML)
	if err != nil {
		return "", fmt.Errorf("error with SanitizeString(): %w", err)
	}
	return sHTML, nil
}

var allowedHTMLTags = []*htmlsanitizer.Tag{
	{Name: "a", Attr: nil, URLAttr: []string{"href"}},
	{Name: "h1", Attr: []string{"style"}, URLAttr: nil},
	{Name: "h2", Attr: []string{"style"}, URLAttr: nil},
	{Name: "li", Attr: []string{"style"}, URLAttr: nil},
	{Name: "strong", Attr: nil, URLAttr: nil},
	{Name: "ol", Attr: []string{"style"}, URLAttr: nil},
	{Name: "p", Attr: []string{"style"}, URLAttr: nil},
	{Name: "ul", Attr: []string{"style"}, URLAttr: nil},
	{Name: "b", Attr: nil, URLAttr: nil},
	{Name: "span", Attr: []string{"style"}, URLAttr: nil},
	{Name: "i", Attr: nil, URLAttr: nil},
	{Name: "u", Attr: nil, URLAttr: nil},
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
			return insertedCommentID, fmt.Errorf("error inserting comment without files: %v", err)
		}
		insertedCommentID = returnCommentID.String()
		return insertedCommentID, nil
	}

	_, err = database.DB.Exec(`INSERT INTO files (file_id, user_id, file_key, file_thumbnail_key, file_store, file_bucket, file_base_url, uploaded_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`, c.File.ID, c.UserID, c.File.Key, c.File.ThumbnailKey, c.File.Store, c.File.Bucket, c.File.BaseURL, time.Now())
	if err != nil {
		return insertedCommentID, fmt.Errorf("error inserting file (before inserting comment): %v", err)
	}
	err = database.DB.QueryRow(`INSERT INTO comments (user_id, content, created_at, post_id, file_id) VALUES ($1, $2, NOW(), $3, $4) RETURNING comment_id`, c.UserID, c.Content, c.PostID, c.File.ID).Scan(&returnCommentID)
	if err != nil {
		return insertedCommentID, fmt.Errorf("error inserting comment (after file inserted): %v", err)
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
		return comments, fmt.Errorf("error querying comments: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c Comment

		if err := rows.Scan(&c.ID, &c.UserID, &c.Content, &c.CreatedAt.Time, &c.PostID, &c.NullFile.ID, &c.NullFile.Key, &c.NullFile.ThumbnailKey, &c.NullFile.Store, &c.NullFile.Bucket, &c.CommentStats.Count, &c.CommentStats.IDsVoted, &c.PreferredName, &c.Avatar); err != nil {
			return comments, fmt.Errorf("error scanning comments: %v", err)
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
		return c, fmt.Errorf("error querying row to getcomment(): %w", err)
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
		return fmt.Errorf("error using sanitizeHTML for edit comment: %v", err)
	}

	if c.File.Key == "" {
		_, err = database.DB.Exec(`UPDATE comments SET content=$1 WHERE comment_id=$2 AND user_id=$3`, sanitizedHTML, c.ID, c.UserID)
		if err != nil {
			return fmt.Errorf("error updating comments table without file: %v", err)
		}
		fmt.Println("Successfully edited!")
		return nil
	}
	err = database.DB.QueryRow(`INSERT INTO files (user_id, file_key, file_thumbnail_key, file_store, file_bucket, file_base_url, uploaded_at) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING file_id`, c.UserID, c.File.Key, c.File.ThumbnailKey, c.File.Store, c.File.Bucket, c.File.BaseURL, time.Now()).Scan(&returnID)
	if err != nil {
		return fmt.Errorf("error inserting file for edit comment: %v", err)
	}
	_, err = database.DB.Exec(`UPDATE comments SET content=$1, file_id=$2 WHERE comment_id=$3 AND user_id=$4`, sanitizedHTML, returnID.String(), c.ID, c.UserID)
	if err != nil {
		return fmt.Errorf("error editing comment after inserting file: %v", err)
	}

	fmt.Println("Successfully edited and updated files table!")
	return nil
}

func Delete(commentID string, username string) error {
	_, err := database.DB.Exec(`DELETE FROM comments WHERE comment_id=$1 AND user_id=$2`, commentID, username)
	if err != nil {
		return fmt.Errorf("error deleting comment: %v", err)
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
		return fmt.Errorf("error fetching comment_id for upvote: %v", err)
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
		return fmt.Errorf("error inserting upvote value: %v", err)
	}

	return nil
}

func ConvertDate(date string) (string, error) {
	var s string
	var suffix string
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return s, fmt.Errorf("error parsing time to convert date: %v", err)
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
	return s, nil
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

	var rows *sql.Rows
	var err error
	// Useful resource for the join - https://stackoverflow.com/questions/2215754/sql-left-join-count
	// I considered left join for post description, but it was stupid to append description to every comment.
	// Decided to just do a separate query for that instead.
	if filter != "" {
		rows, err = database.DB.Query(q, postID, filter)
		if err != nil {
			return comments, fmt.Errorf("error querying filtered comments: %v", err)
		}
	} else {
		rows, err = database.DB.Query(q, postID)
		if err != nil {
			return comments, fmt.Errorf("error querying (without filter) comments: %v", err)
		}
	}
	defer rows.Close()
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.UserID, &c.Content, &c.CreatedAt.Time, &c.PostID, &c.NullFile.ID, &c.NullFile.Key, &c.NullFile.ThumbnailKey, &c.NullFile.Store, &c.NullFile.Bucket, &c.NullFile.BaseURL, &c.CommentStats.Count, &c.CommentStats.IDsVoted, &c.PreferredName, &c.Avatar); err != nil {
			return comments, fmt.Errorf("error scanning for ListCommentsFilterSort(): %w", err)
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

func SearchComments(query string, sort string) ([]SearchComment, error) {
	var results []SearchComment
	var rows *sql.Rows
	var err error
	fmt.Println("Sort: .....", sort)

	if sort == "recent" {
		rows, err = database.DB.Query(`SELECT comments.comment_id, comments.user_id, users.preferred_name, comments.content, comments.created_at, comments.post_id, posts.post_title, comments.file_id FROM comments
										LEFT JOIN users
										ON comments.user_id = users.user_id
										LEFT JOIN posts
										ON posts.post_id = comments.post_id
										WHERE ts @@ websearch_to_tsquery('english', $1)
										ORDER BY comments.created_at DESC`, query)
	} else {
		rows, err = database.DB.Query(`SELECT comments.comment_id, comments.user_id, users.preferred_name, comments.content, comments.created_at, comments.post_id, posts.post_title, comments.file_id FROM comments
										LEFT JOIN users
										ON comments.user_id = users.user_id
										LEFT JOIN posts
										ON posts.post_id = comments.post_id
										WHERE ts @@ websearch_to_tsquery('english', $1)`, query)
	}
	if err != nil {
		return results, fmt.Errorf("error querying db for search comments: %w", err)
	}
	defer rows.Close()
	var row SearchComment
	for rows.Next() {
		if err := rows.Scan(&row.ID, &row.UserID, &row.PreferredName, &row.Content, &row.CreatedAt.Time, &row.PostID, &row.PostTitle, &row.NullFile.ID); err != nil {
			return results, fmt.Errorf("error scanning db for search comments: %w", err)
		}
		if row.NullFile.ID.Valid {
			row.File.ID = row.NullFile.ID.UUID
		}
		results = append(results, row)
	}
	return results, nil
}
