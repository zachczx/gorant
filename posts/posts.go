package posts

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorant/database"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rezakhademix/govalidator/v2"
)

type PostLike struct {
	LikeID uuid.UUID `db:"like_id"`
	UserID string    `db:"user_id"`
	PostID string    `db:"post_id"`
	Score  string    `db:"score"`
}

type JunctionPostTag struct {
	PostsTagsID uuid.UUID `db:"posts_tags_id"`
	PostID      string    `db:"post_id"`
	TagID       uuid.UUID `db:"tag_id"`
	Tag         string
}

type Post struct {
	ID            string `db:"post_id"`
	Title         string `db:"post_title"`
	UserID        string `db:"user_id"`
	Description   string `db:"description"`
	Protected     int    `db:"protected"`
	CreatedAt     CreatedAt
	Mood          string `db:"mood"`
	Similarity    float32
	Tags          Tags
	PostStats     PostStats
	PreferredName string // PreferredName of Author
}

type CreatedAt struct {
	Time time.Time
}

type Tags struct {
	TagsNullString sql.NullString `db:"tags"`
	Tags           []string
}

func (t *Tags) Count() string {
	return strconv.Itoa(len(t.Tags))
}

type Tag struct {
	TagID uuid.UUID `db:"tag_id"`
	Tag   string    `db:"tag"`
}

type PostStats struct {
	CommentsCount         sql.NullInt64 `db:"comments_cnt"`
	RepliesCount          sql.NullInt64
	LikesCount            sql.NullInt64 `db:"likes_cnt"`
	CurrentUserLike       sql.NullInt64 `db:"score"`
	CurrentUserLikeString string
}

func (ps *PostStats) RepliesCountString() string {
	if ps.RepliesCount.Valid {
		return strconv.FormatInt(ps.RepliesCount.Int64, 10)
	}
	return "0"
}

func (ps *PostStats) CommentsCountString() string {
	if ps.CommentsCount.Valid {
		return strconv.FormatInt(ps.CommentsCount.Int64, 10)
	}
	return "0"
}

func (ps *PostStats) LikesCountString() string {
	if ps.LikesCount.Valid {
		return strconv.FormatInt(ps.LikesCount.Int64, 10)
	}
	return "0"
}

func (ps *PostStats) CommentsRepliesCountString() string {
	var comments int64
	var replies int64
	if ps.CommentsCount.Valid {
		comments = ps.CommentsCount.Int64
	}
	if ps.RepliesCount.Valid {
		replies = ps.RepliesCount.Int64
	}
	return strconv.FormatInt(comments+replies, 10)
}

type PostCollection []Post

var allowedMoods = [6]string{"Elated", "Happy", "Neutral", "Sad", "Upset", "Angry"}

const regex string = `^[A-Za-z0-9 _!.\$\/\\|()\[\]=` + "`" + `{<>?@#%^&*—:;'"+\-,"]+$`

func (c *CreatedAt) Process() string {
	var s string
	var suffix string
	singleDay := 24.00
	singleHour := 1.00
	n := time.Now()
	diff := n.Sub(c.Time).Hours()
	switch {
	case diff < singleHour:
		if n.Sub(c.Time).Minutes() < 2 {
			suffix = " minute ago"
		} else {
			suffix = " minutes ago"
		}
		// Mins
		s = strconv.Itoa(int(n.Sub(c.Time).Minutes())) + suffix
	case diff >= singleHour && diff < singleDay:
		if diff < 2 {
			suffix = " hour ago"
		} else {
			suffix = " hours ago"
		}
		// Hours
		s = strconv.Itoa(int(diff)) + suffix
	case diff >= singleDay:
		if diff < 2*singleDay {
			suffix = " day ago"
		} else {
			suffix = " days ago"
		}
		// Days
		s = strconv.Itoa(int(n.Sub(c.Time).Hours()/singleDay)) + suffix
	default:
		fmt.Println("Something went wrong")
	}
	return s
}

var listPostsQuery = `SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood, users.preferred_name, comments_cnt, replies_cnt, likes_cnt, tags
									FROM posts
										LEFT JOIN users ON users.user_id=posts.user_id
										LEFT JOIN(SELECT comments.post_id, COUNT(1) AS comments_cnt
												FROM comments
												GROUP BY comments.post_id) AS comments ON comments.post_id=posts.post_id
										LEFT JOIN(SELECT replies.post_id, COUNT(1) AS replies_cnt
												FROM replies 
												GROUP BY replies.post_id) AS replies ON replies.post_id=posts.post_id
										LEFT JOIN(SELECT post_id, COUNT(1) AS likes_cnt
												FROM posts_likes
												GROUP BY posts_likes.post_id) AS posts_likes ON posts.post_id=posts_likes.post_id
										LEFT JOIN(SELECT posts_tags.post_id, string_agg(tags.tag, ',') AS tags
												FROM posts_tags
														LEFT JOIN tags ON posts_tags.tag_id=tags.tag_id
												GROUP BY posts_tags.post_id) AS posts_tags ON posts.post_id=posts_tags.post_id
									ORDER BY posts.created_at DESC`

func ListPosts() (PostCollection, error) {
	return scanPosts(listPostsQuery)
}

func ListPostsFilter(mood []string, tags []string) (PostCollection, error) {
	var query string
	var args []interface{}
	var err error
	// For when there's a reset of the form
	if len(mood) == 0 && len(tags) == 0 {
		p, err := ListPosts()
		if err != nil {
			return p, fmt.Errorf("error: fetching posts: %w", err)
		}
		return p, nil
	}
	// IN clause fails if the tags slice fed to the INNER JOIN is empty, so if it's empty, grab all the possible moods
	if len(mood) != 0 && len(tags) == 0 {
		// This is different because 2 INNER JOINs are now LEFT JOINs
		// Specifically: (SELECT DISTINCT... and SELECT posts_tags.post_id...)
		query, args, err = sqlx.In(`SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood, users.preferred_name, comments.comments_cnt, replies.replies_cnt, posts_likes.likes_cnt, posts_tags.tags
								FROM(SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood
									FROM posts
									WHERE mood IN (?)) AS posts
									LEFT JOIN(SELECT DISTINCT posts_tags.post_id
											from posts_tags
													INNER JOIN(SELECT tags.tag_id, tags.tag FROM tags) AS tags ON posts_tags.tag_id=tags.tag_id) AS selected_tags ON selected_tags.post_id=posts.post_id
									LEFT JOIN users ON users.user_id=posts.user_id
									LEFT JOIN(SELECT comments.post_id, COUNT(1) AS comments_cnt
											FROM comments
											GROUP BY comments.post_id) AS comments ON comments.post_id=posts.post_id
									LEFT JOIN(SELECT replies.post_id, COUNT(1) AS replies_cnt
											FROM replies 
											GROUP BY replies.post_id) AS replies ON replies.post_id=posts.post_id
									LEFT JOIN(SELECT post_id, COUNT(1) AS likes_cnt
											FROM posts_likes
											GROUP BY posts_likes.post_id) AS posts_likes ON posts.post_id=posts_likes.post_id
									LEFT JOIN(SELECT posts_tags.post_id, string_agg(tags.tag, ',') AS tags
											FROM posts_tags
													LEFT JOIN tags ON posts_tags.tag_id=tags.tag_id
											GROUP BY posts_tags.post_id) AS posts_tags ON posts.post_id=posts_tags.post_id
								ORDER BY posts.created_at DESC`, mood)
	} else if len(mood) == 0 && len(tags) != 0 {
		query, args, err = sqlx.In(`SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood, users.preferred_name, comments.comments_cnt, replies.replies_cnt, posts_likes.likes_cnt, posts_tags.tags
									FROM(SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood
										FROM posts
											INNER JOIN(SELECT DISTINCT posts_tags.post_id
														FROM posts_tags
															INNER JOIN(SELECT tags.tag_id, tags.tag FROM tags WHERE tags.tag IN (?)) AS tags ON posts_tags.tag_id=tags.tag_id) AS selected_tags ON selected_tags.post_id=posts.post_id) AS posts
										LEFT JOIN users ON users.user_id=posts.user_id
										LEFT JOIN(SELECT comments.post_id, COUNT(1) AS comments_cnt
												FROM comments
												GROUP BY comments.post_id) AS comments ON comments.post_id=posts.post_id
										LEFT JOIN(SELECT replies.post_id, COUNT(1) AS replies_cnt
												FROM replies
												GROUP BY replies.post_id) AS replies ON replies.post_id=posts.post_id
										LEFT JOIN(SELECT post_id, COUNT(1) AS likes_cnt
												FROM posts_likes
												GROUP BY posts_likes.post_id) AS posts_likes ON posts.post_id=posts_likes.post_id
										INNER JOIN(SELECT posts_tags.post_id, string_agg(tags.tag, ',') AS tags
												FROM posts_tags
														LEFT JOIN tags ON posts_tags.tag_id=tags.tag_id
												GROUP BY posts_tags.post_id) AS posts_tags ON posts.post_id=posts_tags.post_id
									ORDER BY posts.created_at DESC`, tags)
	} else {
		query, args, err = sqlx.In(`SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood, users.preferred_name, comments.comments_cnt, replies.replies_cnt, posts_likes.likes_cnt, posts_tags.tags
								FROM(SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood
									FROM posts
									WHERE mood IN (?)) AS posts
									INNER JOIN(SELECT DISTINCT posts_tags.post_id
											from posts_tags
													INNER JOIN(SELECT tags.tag_id, tags.tag FROM tags WHERE tags.tag IN (?)) AS tags ON posts_tags.tag_id=tags.tag_id) AS selected_tags ON selected_tags.post_id=posts.post_id
									LEFT JOIN users ON users.user_id=posts.user_id
									LEFT JOIN(SELECT comments.post_id, COUNT(1) AS comments_cnt
											FROM comments
											GROUP BY comments.post_id) AS comments ON comments.post_id=posts.post_id
									LEFT JOIN(SELECT replies.post_id, COUNT(1) AS replies_cnt
											FROM replies 
											GROUP BY replies.post_id) AS replies ON replies.post_id=posts.post_id
									LEFT JOIN(SELECT post_id, COUNT(1) AS likes_cnt
											FROM posts_likes
											GROUP BY posts_likes.post_id) AS posts_likes ON posts.post_id=posts_likes.post_id
									INNER JOIN(SELECT posts_tags.post_id, string_agg(tags.tag, ',') AS tags
											FROM posts_tags
													LEFT JOIN tags ON posts_tags.tag_id=tags.tag_id
											GROUP BY posts_tags.post_id) AS posts_tags ON posts.post_id=posts_tags.post_id
								ORDER BY posts.created_at DESC`, mood, tags)
	}
	if err != nil {
		return nil, fmt.Errorf("error executing sqlx.In: %w", err)
	}
	query = database.DB.Rebind(query)
	return scanPosts(query, args...)
}

func scanPosts(query string, args ...interface{}) (PostCollection, error) {
	rows, err := database.DB.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing list-post-filter query: %w", err)
	}
	defer rows.Close()
	var posts PostCollection
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Description, &p.Protected, &p.CreatedAt.Time, &p.Mood, &p.PreferredName, &p.PostStats.CommentsCount, &p.PostStats.RepliesCount, &p.PostStats.LikesCount, &p.Tags.TagsNullString); err != nil {
			return nil, fmt.Errorf("error scanning list-post-filter: %w", err)
		}
		if p.Tags.TagsNullString.Valid {
			p.Tags.Tags = strings.Split(p.Tags.TagsNullString.String, ",")
		} else {
			p.Tags.Tags = []string{}
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func ListTags() ([]string, error) {
	var tags []string
	var tagID uuid.UUID
	var tag string
	rows, err := database.DB.Query(`SELECT posts_tags.tag_id, tags.tag
									FROM(SELECT DISTINCT posts_tags.tag_id FROM posts_tags) AS posts_tags
										LEFT JOIN tags ON posts_tags.tag_id=tags.tag_id
									ORDER BY tags.tag`)
	if err != nil {
		return tags, fmt.Errorf("error querying posts_tags for listtags(): %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&tagID, &tag); err != nil {
			return tags, fmt.Errorf("error scanning posts_tags for listtags(): %w", err)
		}
		tags = append(tags, tag)
	}
	fmt.Println(tags)
	return tags, nil
}

func NewPost(p Post, tags []string) error {
	var postID string
	err := database.DB.QueryRow(`INSERT INTO posts (post_id, post_title, user_id, created_at, mood) VALUES ($1, $2, $3, NOW(), $4) 
						RETURNING post_id`, p.ID, p.Title, p.UserID, p.Mood).Scan(&postID)
	if err != nil {
		return fmt.Errorf("error inserting post: %w", err)
	}

	fmt.Println("Length: ", len(tags))
	if len(tags) == 0 {
		return nil
	}

	// Insert individual tags into tags table
	if _, err = InsertTags(tags); err != nil {
		// Print error instead of returning, duplicate value error is alright
		fmt.Println(err)
	}

	for _, v := range tags {
		// Copy postID and tag_id where the tag == tag value, and insert it into posts_tags
		_, err = database.DB.Exec(`INSERT INTO posts_tags (post_id, tag_id) SELECT $1, tag_id FROM tags WHERE tag=$2`, postID, v)
		if err != nil {
			// this err needs to be returned because it's not normal
			return fmt.Errorf("error inserting tags into post_tag table: %w", err)
		}
	}

	return nil
}

func GetTags(postID string) (Post, error) {
	var t string
	var p Post

	// Splitting into 2 queries because its easier instead of querying by posts then left join
	if err := database.DB.QueryRow(`SELECT posts.post_id, posts.user_id FROM posts WHERE posts.post_id=$1`, postID).Scan(&p.ID, &p.UserID); err != nil {
		if err == sql.ErrNoRows {
			return p, fmt.Errorf("error: postID no rows: %w", err)
		}
		return p, fmt.Errorf("error querying for UserID: %w", err)
	}
	rows, err := database.DB.Query(`SELECT tags.tag FROM (SELECT posts_tags.post_id, posts_tags.tag_id
															FROM posts_tags
															WHERE posts_tags.post_id=$1) AS posts_tags
									LEFT JOIN tags ON posts_tags.tag_id=tags.tag_id`, postID)
	if err != nil {
		return p, fmt.Errorf("error querying for gettags(): %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&p.UserID, &t); err != nil {
			return p, fmt.Errorf("error scanning for gettags(): %w", err)
		}

		p.Tags.Tags = append(p.Tags.Tags, t)
	}
	return p, nil
}

func EditTags(postID string, tags []string) error {
	var err error

	// Even if len(validTags) == 0, I can't end the func now, because I'll need to delete all the posts.
	// We can stop when needing to insert tags.
	validTags, err := InsertTags(tags)
	if err != nil {
		// Print error instead of returning, duplicate value error is alright
		fmt.Println(err)
	}

	postsTags, err := GetPostIDTagIDTag(postID)
	if err != nil {
		return fmt.Errorf("error from GetPostIDTagIDTag() junction table: %w", err)
	}

	if err := DeleteUnwantedTags(tags, postsTags); err != nil {
		fmt.Println(err)
	}

	// Now I can end the func if len(validTags) == 0, since we've already deleted what we needed
	if len(validTags) == 0 {
		fmt.Println("Ending func since no valid tags were found!")
		return nil
	}

	if err := InsertPostTags(postID, validTags, postsTags); err != nil {
		fmt.Println(err)
	}

	return nil
}

func InsertTags(tags []string) ([]Tag, error) {
	var validTags []Tag
	if len(tags) == 0 {
		fmt.Println("No valid tags found!")
		return validTags, errors.New("no valid tag found")
	}

	var err error

	var tag Tag
	for _, v := range tags {
		// Reusing SanitizeTitleToID to sanitize input
		tag.Tag, err = SanitizeTitleToID(v)
		if err != nil {
			return validTags, fmt.Errorf("error sanitizing tags: %w", err)
		}
		validTags = append(validTags, tag)
	}

	fmt.Println("Valid tags: ", validTags)
	fmt.Println("Number of valid tags: ", len(validTags))

	_, err = database.DB.NamedExec(`INSERT INTO tags (tag) VALUES (:tag) ON CONFLICT (tag) DO NOTHING`, validTags)
	if err != nil {
		return validTags, fmt.Errorf("error inserting tags: %w", err)
	}

	return validTags, nil
}

func InsertPostTags(postID string, validTags []Tag, postsTags []JunctionPostTag) error {
	var tagsToInsert []string

	for _, v := range validTags {
		if exists := containsJunctionPostTag(postsTags, v.Tag); !exists {
			tagsToInsert = append(tagsToInsert, v.Tag)
		}
	}

	fmt.Println("Insert into PostTags: ", len(tagsToInsert), " - ", tagsToInsert)

	// insert into posts_tags
	if len(tagsToInsert) > 0 {
		for _, v := range tagsToInsert {
			_, err := database.DB.Exec(`INSERT INTO posts_tags (post_id, tag_id) SELECT $1, tag_id FROM tags WHERE tag=$2`, postID, v)
			if err != nil {
				return fmt.Errorf("error inserting tags into posts_tags: %w", err)
			}
			fmt.Printf("Successfully saved: \"%s\" into \"%s\"\n", v, postID)
		}
	}
	return nil
}

func GetPostIDTagIDTag(postID string) ([]JunctionPostTag, error) {
	var pt JunctionPostTag
	var postsTags []JunctionPostTag
	// Grab all the existing tags from post using postID
	rows, err := database.DB.Query(`SELECT posts_tags.post_id, posts_tags.tag_id, tags.tag FROM posts_tags LEFT JOIN tags ON posts_tags.tag_id = tags.tag_id WHERE post_id=$1`, postID)
	if err != nil {
		return postsTags, fmt.Errorf("error selecting posts_tags for GetPostIDTagIDTag(): %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&pt.PostID, &pt.TagID, &pt.Tag); err != nil {
			return postsTags, fmt.Errorf("error scanning posts_tags for GetPostIDTagIDTag(): %w", err)
		}
		postsTags = append(postsTags, pt)
		fmt.Println("Tags in post now: ", pt.Tag)
	}
	return postsTags, nil
}

/* var disallowed = [33]string{
	".",
	" ",
	"_",
	"!",
	".",
	"$",
	"/",
	"\\",
	"|",
	"(",
	")",
	"[",
	"]",
	"=",
	"`",
	"{",
	"}",
	"<",
	">",
	"?",
	"@",
	"#",
	"%",
	"^",
	"&",
	"*",
	"—",
	":",
	"'",
	";",
	"\"",
	"+",
	",",
} */

var validTagCharacters = `^[A-Za-z0-9-]+$`

func ValidateTags(tag string) (passed bool, err error) {
	regex, err := regexp.Compile(validTagCharacters)
	if err != nil {
		return false, fmt.Errorf("regexp compilation error: %w", err)
	}

	if !regex.MatchString(tag) {
		return false, nil
	}
	return true, nil
}

func DeleteUnwantedTags(inputTags []string, postsTags []JunctionPostTag) error {
	// Loop through postTags to find tags that are not in current user input, then mark those for deletion.
	var tagsToDeleteID []uuid.UUID
	var exists bool
	for _, v := range postsTags {
		if exists = contains(inputTags, v.Tag); !exists {
			tagsToDeleteID = append(tagsToDeleteID, v.TagID)
		}
	}

	fmt.Println("Tags to Delete: ", tagsToDeleteID)

	// Delete from posts_tags if tagsToDelete is not empty
	if len(tagsToDeleteID) > 0 {
		for _, v := range tagsToDeleteID {
			_, err := database.DB.Exec(`DELETE FROM posts_tags WHERE tag_id=$1`, v)
			if err != nil {
				fmt.Println("Error in deleting")
				return fmt.Errorf("error deleting tags from posts_tags: %w", err)
			}
		}
	}

	return nil
}

func contains(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

func containsJunctionPostTag(a []JunctionPostTag, s string) bool {
	for _, v := range a {
		if v.Tag == s {
			return true
		}
	}
	return false
}

func ValidatePost(postTitle string) map[string](string) {
	v := govalidator.New()

	v.RequiredString(postTitle, "postTitle", "Please enter an ID").RegexMatches(postTitle, regex, "postTitle", "Special characters not allowed").MaxString(postTitle, 255, "postTitle", "That's too long! Max length of title is 255 characters.")

	if v.IsFailed() {
		return v.Errors()
	}

	return nil
}

func VerifyPostID(title string) (bool, string) {
	var ID string

	// Generate ID to test
	ID, err := SanitizeTitleToID(title)
	if err != nil {
		fmt.Println(err)
	}

	res, err := database.DB.Query("SELECT post_id FROM posts WHERE post_id=$1;", ID)
	if err != nil {
		fmt.Println("Error executing query to verify post exists")
		fmt.Println(err)
	}

	defer res.Close()

	return res.Next(), ID
}

func GetPost(postID string, currentUser string) (Post, error) {
	var p Post
	row, err := database.DB.Query(`SELECT posts.post_id, posts.post_title, posts.user_id, posts.description, posts.protected, posts.created_at, posts.mood, posts_likes.score, STRING_AGG(posts_tags.tag, ',') AS tags
									FROM posts
										LEFT JOIN (SELECT * FROM posts_likes) AS posts_likes ON posts.post_id = posts_likes.post_id
										AND posts_likes.user_id = $2
										LEFT JOIN (SELECT posts_tags.post_id, tags.tag
											FROM posts_tags
											LEFT JOIN tags ON posts_tags.tag_id = tags.tag_id) AS posts_tags ON posts_tags.post_id = posts.post_id
									WHERE posts.post_id = $1
									GROUP BY posts.post_id, posts.post_title, posts.user_id, posts.description, posts.protected, posts.created_at, posts.mood, posts_likes.score;`, postID, currentUser)
	if err != nil {
		return p, fmt.Errorf("error querying getpost(): %w", err)
	}
	defer row.Close()

	for row.Next() {
		if err := row.Scan(&p.ID, &p.Title, &p.UserID, &p.Description, &p.Protected, &p.CreatedAt.Time, &p.Mood, &p.PostStats.CurrentUserLike, &p.Tags.TagsNullString); err != nil {
			return p, fmt.Errorf("error scanning getpost(): %w", err)
		}

		if p.Tags.TagsNullString.Valid {
			p.Tags.Tags = strings.Split(p.Tags.TagsNullString.String, ",")
		} else {
			p.Tags.Tags = []string{}
		}

		if p.PostStats.CurrentUserLike.Valid {
			p.PostStats.CurrentUserLikeString = strconv.FormatInt(p.PostStats.CurrentUserLike.Int64, 10)
		} else {
			p.PostStats.CurrentUserLikeString = "0"
		}
	}

	return p, nil
}

func LikePost(postID string, currentUser string) (int, error) {
	var score int
	var exists string
	err := database.DB.QueryRow("SELECT score FROM posts_likes WHERE post_id=$1 AND user_id=$2", postID, currentUser).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			if _, err = database.DB.Exec("INSERT INTO posts_likes (user_id, post_id, score) VALUES ($1, $2, 1)", currentUser, postID); err != nil {
				return score, fmt.Errorf("error inserting into posts_likes: %w", err)
			}
			score = 1
			return score, nil
		}
		return score, fmt.Errorf("error querying row from posts_likes: %w", err)
	}
	_, err = database.DB.Exec("DELETE FROM posts_likes WHERE post_id=$1 AND user_id=$2", postID, currentUser)
	if err != nil {
		return score, fmt.Errorf("error deleting from posts_likes: %w", err)
	}
	score = 0
	return score, nil
}

func EditPostDescription(postID string, description string) error {
	_, err := database.DB.Exec("UPDATE posts SET description=$1 WHERE post_id=$2", description, postID)
	if err != nil {
		return fmt.Errorf("error updating description in posts table: %w", err)
	}
	return nil
}

func DeletePost(postID string, username string) error {
	var u string
	if err := database.DB.QueryRow("SELECT user_id FROM posts WHERE post_id=$1", postID).Scan(&u); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("error cannot find user_id with given postID: %w", err)
		}
		return fmt.Errorf("error with querying user_id to delete post: %w", err)
	}
	if u != username {
		return errors.New("error: logged in user is not owner of post")
	}
	if _, err := database.DB.Exec("DELETE FROM posts WHERE post_id=$1", postID); err != nil {
		return fmt.Errorf("error deleting post: %w", err)
	}
	return nil
}

func EditMood(postID string, mood string) error {
	if err := ValidateMood(mood); err != nil {
		return fmt.Errorf("error from validate mood: %w", err)
	}

	_, err := database.DB.Exec("UPDATE posts SET mood=$1 WHERE post_id=$2", mood, postID)
	if err != nil {
		return fmt.Errorf("error updating posts for edit mode: %w", err)
	}

	return nil
}

func ValidateMood(mood string) error {
	for _, v := range allowedMoods {
		if strings.EqualFold(v, mood) {
			return nil
		}
	}
	return errors.New("new mood is not in allowed list")
}

// Stopwords from Bow (libbow, rainbow, arrow, crossbow).
// List from https://github.com/igorbrigadir/stopwords?tab=readme-ov-file
var stopWords = [48]string{
	"about",
	"all",
	"am",
	"an",
	"and",
	"are",
	"as",
	"at",
	"be",
	"been",
	"but",
	"by",
	"can",
	"cannot",
	"did",
	"do",
	"does",
	"doing",
	"done",
	"for",
	"from",
	"had",
	"has",
	"have",
	"having",
	"if",
	"in",
	"is",
	"it",
	"its",
	"of",
	"on",
	"that",
	"the",
	"these",
	"they",
	"this",
	"those",
	"to",
	"too",
	"want",
	"wants",
	"was",
	"what",
	"which",
	"will",
	"with",
	"would",
}

func SanitizeTitleToID(inputTitle string) (string, error) {
	title := strings.Fields(inputTitle)
	var titleNoStopWords []string

out:
	for _, titleWord := range title {
		for _, word := range stopWords {
			if strings.EqualFold(titleWord, word) {
				continue out
			}
		}
		titleNoStopWords = append(titleNoStopWords, titleWord)
	}

	// If the title is purely stop words only, then revert to the title with stopwords.
	if len(titleNoStopWords) == 0 {
		titleNoStopWords = title
	}
	titleJoined := strings.Join(titleNoStopWords, " ")
	r := strings.NewReplacer(
		" ", "-",
		".", "",
		"_", "-",
		"!", "",
		".", "",
		"$", "",
		"/", "",
		"\\", "",
		"|", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"=", "",
		"`", "",
		"{", "",
		"}", "",
		"<", "",
		">", "",
		"?", "",
		"@", "",
		"#", "",
		"%", "",
		"^", "",
		"&", "",
		"*", "",
		"—", "",
		":", "",
		"'", "",
		";", "",
		"\"", "",
		"+", "",
		",", "",
	)
	ID := r.Replace(strings.ToLower(titleJoined))
	if len(ID) > 60 {
		ID = ID[:60]
	}
	return ID, nil
}

func NullIntToString(n sql.NullInt64) string {
	var s string
	if n.Valid {
		s = strconv.FormatInt(n.Int64, 10)
	} else {
		s = "0"
	}
	return s
}

func RelatedPosts(post Post, results int) (PostCollection, error) {
	limit := strconv.Itoa(results)
	query := `SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.created_at, posts.mood, similarity(post_title, 'chinese astrology') AS similarity, comments_cnt, replies_cnt, likes_cnt, tags
				FROM posts
					LEFT JOIN(SELECT comments.post_id, COUNT(1) AS comments_cnt
							FROM comments
							GROUP BY comments.post_id) AS comments ON comments.post_id=posts.post_id
					LEFT JOIN(SELECT replies.post_id, COUNT(1) AS replies_cnt
							FROM replies
							GROUP BY replies.post_id) AS replies ON replies.post_id=posts.post_id
					LEFT JOIN(SELECT post_id, COUNT(1) AS likes_cnt
							FROM posts_likes
							GROUP BY posts_likes.post_id) AS posts_likes ON posts.post_id=posts_likes.post_id
					LEFT JOIN(SELECT posts_tags.post_id, string_agg(tags.tag, ',') AS tags
							FROM posts_tags
									LEFT JOIN tags ON posts_tags.tag_id=tags.tag_id
							GROUP BY posts_tags.post_id) AS posts_tags ON posts.post_id=posts_tags.post_id
				WHERE posts.post_title % $1 AND NOT posts.post_id=$2
				ORDER BY similarity DESC
				LIMIT $3;`
	rows, err := database.DB.Query(query, post.Title, post.ID, limit)
	if err != nil {
		return nil, fmt.Errorf("error: relatedposts query: %w", err)
	}
	defer rows.Close()
	var posts PostCollection
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Description, &p.CreatedAt.Time, &p.Mood, &p.Similarity, &p.PostStats.CommentsCount, &p.PostStats.RepliesCount, &p.PostStats.LikesCount, &p.Tags.TagsNullString); err != nil {
			return nil, fmt.Errorf("error scanning list-post-filter: %w", err)
		}
		if p.Tags.TagsNullString.Valid {
			p.Tags.Tags = strings.Split(p.Tags.TagsNullString.String, ",")
		} else {
			p.Tags.Tags = []string{}
		}
		posts = append(posts, p)
	}
	return posts, nil
}
