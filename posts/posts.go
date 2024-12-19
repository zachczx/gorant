package posts

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorant/database"

	"github.com/jmoiron/sqlx"
	"github.com/rezakhademix/govalidator/v2"
)

type Post struct {
	PostID             string `db:"post_id"`
	PostTitle          string `db:"post_title"`
	UserID             string `db:"user_id"`
	Description        string `db:"description"`
	Protected          int    `db:"protected"`
	CreatedAt          string `db:"created_at"`
	Mood               string `db:"mood"`
	CreatedAtProcessed string
	Tags               string `db:"tags"`
}

type PostLike struct {
	ID     int    `db:"like_id"`
	UserID string `db:"user_id"`
	PostID string `db:"post_id"`
	Score  string `db:"score"`
}

type Tag struct {
	TagID int    `db:"tag_id"`
	Tag   string `db:"tag"`
}

type PostTag struct {
	PostsTagID int    `db:"posts_tags_id"`
	PostID     string `db:"post_id"`
	TagID      int    `db:"tag_id"`
	Tag        string
}

type JoinPost struct {
	PostID                string `db:"post_id"`
	PostTitle             string `db:"post_title"`
	UserID                string `db:"user_id"`
	Description           string `db:"description"`
	Protected             int    `db:"protected"`
	CreatedAt             string `db:"created_at"`
	Mood                  string `db:"mood"`
	PreferredName         string
	CommentsCount         sql.NullInt64 `db:"comments_cnt"`
	CommentsCountString   string
	CreatedAtProcessed    string
	LikesCount            sql.NullInt64 `db:"likes_cnt"`
	LikesCountString      string
	CurrentUserLike       sql.NullInt64 `db:"score"`
	CurrentUserLikeString string
	TagsNullString        sql.NullString `db:"tags"`
	Tags                  []string
}

type ZPost struct {
	PostID    string `db:"post_id"`
	PostTitle string `db:"post_title"`

	// Author and PreferredName
	UserID        string `db:"user_id"`
	PreferredName string
	Description   string `db:"description"`
	Protected     int    `db:"protected"`
	CreatedAt     CreatedAt
	Mood          string `db:"mood"`
	Tags          Tags
	PostStats     ZPostStats
}

type CreatedAt struct {
	CreatedAtString    string `db:"created_at"`
	CreatedAtProcessed string
}

type Tags struct {
	TagsNullString sql.NullString `db:"tags"`
	Tags           []string
}

type ZPostStats struct {
	CommentsCount         sql.NullInt64 `db:"comments_cnt"`
	CommentsCountString   string
	LikesCount            sql.NullInt64 `db:"likes_cnt"`
	LikesCountString      string
	CurrentUserLike       sql.NullInt64 `db:"score"`
	CurrentUserLikeString string
}

func StopSquigglyLines() {
	// PPP := ZPost{Tags: zTags{Tags: []string{"111", "222"}}}
	// fmt.Println(PPP.Tags.Tags)
}

type ListService interface {
	ListPosts() ([]JoinPost, error)
	ListTags() ([]string, error)
	ListPostsFilter(mood []string, tags []string) ([]JoinPost, error)
	ListComments(postID string, currentUser string) ([]JoinComment, error)
	ListCommentsFilterSort(postID string, currentUser string, sort string, filter string) ([]JoinComment, error)
}

const regex string = `^[A-Za-z0-9 _!.\$\/\\|()\[\]=` + "`" + `{<>?@#%^&*—:;'"+\-,"]+$`

func ListPosts() ([]ZPost, error) {
	rows, err := database.DB.Query(`SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood, users.preferred_name, comments_cnt, likes_cnt, tags
									FROM posts
										LEFT JOIN users ON users.user_id=posts.user_id
										LEFT JOIN(SELECT comments.post_id, COUNT(1) AS comments_cnt
												FROM comments
												GROUP BY comments.post_id) AS comments ON comments.post_id=posts.post_id
										LEFT JOIN(SELECT post_id, COUNT(1) as likes_cnt
												FROM posts_likes
												GROUP BY posts_likes.post_id) as posts_likes ON posts.post_id=posts_likes.post_id
										LEFT JOIN(SELECT posts_tags.post_id, string_agg(tags.tag, ',') as tags
												FROM posts_tags
														LEFT JOIN tags ON posts_tags.tag_id=tags.tag_id
												GROUP BY posts_tags.post_id) as posts_tags ON posts.post_id=posts_tags.post_id`)
	if err != nil {
		fmt.Println("Error executing query: ", err)
		return nil, err
	}

	defer rows.Close()

	var posts []ZPost

	for rows.Next() {
		var p ZPost

		if err := rows.Scan(&p.PostID, &p.UserID, &p.PostTitle, &p.Description, &p.Protected, &p.CreatedAt.CreatedAtString, &p.Mood, &p.PreferredName, &p.PostStats.CommentsCount, &p.PostStats.LikesCount, &p.Tags.TagsNullString); err != nil {
			fmt.Println("Error scanning")
			return nil, err
		}

		if p.PostStats.CommentsCount.Valid {
			p.PostStats.CommentsCountString = strconv.FormatInt(p.PostStats.CommentsCount.Int64, 10)
		} else {
			p.PostStats.CommentsCountString = "0"
		}

		if p.PostStats.LikesCount.Valid {
			p.PostStats.LikesCountString = strconv.FormatInt(p.PostStats.LikesCount.Int64, 10)
		} else {
			p.PostStats.LikesCountString = "0"
		}

		if p.Tags.TagsNullString.Valid {
			p.Tags.Tags = strings.Split(p.Tags.TagsNullString.String, ",")
		} else {
			p.Tags.Tags = []string{}
		}

		p.CreatedAt.CreatedAtProcessed, err = ConvertDate(p.CreatedAt.CreatedAtString)
		if err != nil {
			fmt.Println(err)
		}

		posts = append(posts, p)
	}

	return posts, nil
}

func ListTags() ([]string, error) {
	var tags []string
	var tagID string
	var tag string

	rows, err := database.DB.Query(`SELECT posts_tags.tag_id, tags.tag
									FROM(SELECT DISTINCT posts_tags.tag_id FROM posts_tags) as posts_tags
										LEFT JOIN tags ON posts_tags.tag_id=tags.tag_id
									ORDER BY tags.tag`)
	if err != nil {
		return tags, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&tagID, &tag); err != nil {
			return tags, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func ListPostsFilter(mood []string, tags []string) ([]ZPost, error) {
	query, args, err := sqlx.In(`SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood, users.preferred_name, comments_cnt, likes_cnt, tags
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
									LEFT JOIN(SELECT post_id, COUNT(1) as likes_cnt
											FROM posts_likes
											GROUP BY posts_likes.post_id) as posts_likes ON posts.post_id=posts_likes.post_id
									INNER JOIN(SELECT posts_tags.post_id, string_agg(tags.tag, ',') as tags
											FROM posts_tags
													LEFT JOIN tags ON posts_tags.tag_id=tags.tag_id
											GROUP BY posts_tags.post_id) as posts_tags ON posts.post_id=posts_tags.post_id`, mood, tags)
	if err != nil {
		fmt.Println("Error executing query: ", err)
		return nil, err
	}
	query = database.DB.Rebind(query)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		fmt.Println("Error executing query: ", err)
		return nil, err
	}

	defer rows.Close()

	var posts []ZPost

	for rows.Next() {
		var p ZPost

		if err := rows.Scan(&p.PostID, &p.UserID, &p.PostTitle, &p.Description, &p.Protected, &p.CreatedAt.CreatedAtString, &p.Mood, &p.PreferredName, &p.PostStats.CommentsCount, &p.PostStats.LikesCount, &p.Tags.TagsNullString); err != nil {
			fmt.Println("Error scanning")
			return nil, err
		}

		if p.PostStats.CommentsCount.Valid {
			p.PostStats.CommentsCountString = strconv.FormatInt(p.PostStats.CommentsCount.Int64, 10)
		} else {
			p.PostStats.CommentsCountString = "0"
		}

		if p.PostStats.LikesCount.Valid {
			p.PostStats.LikesCountString = strconv.FormatInt(p.PostStats.LikesCount.Int64, 10)
		} else {
			p.PostStats.LikesCountString = "0"
		}

		if p.Tags.TagsNullString.Valid {
			p.Tags.Tags = strings.Split(p.Tags.TagsNullString.String, ",")
		} else {
			p.Tags.Tags = []string{}
		}

		p.CreatedAt.CreatedAtProcessed, err = ConvertDate(p.CreatedAt.CreatedAtString)
		if err != nil {
			fmt.Println(err)
		}

		posts = append(posts, p)
	}

	return posts, nil
}

func NewPost(p Post, tags []string) error {
	t := time.Now().Format(time.RFC3339)

	var postID string
	err := database.DB.QueryRow(`INSERT INTO posts (post_id, post_title, user_id, created_at, mood) VALUES ($1, $2, $3, $4, $5) 
						RETURNING post_id`, p.PostID, p.PostTitle, p.UserID, t, p.Mood).Scan(&postID)
	if err != nil {
		return err
	}

	fmt.Println("Length: ", len(tags))
	if len(tags) == 0 {
		return nil
	}

	// Prepare tag struct for namedexec
	var tag Tag
	tagsStruct := []Tag{}
	for _, v := range tags {
		// Reusing TitleToID to sanitize input
		tag.Tag, err = TitleToID(v)
		if err != nil {
			return err
		}

		tagsStruct = append(tagsStruct, tag)
	}

	// insert tags into tags table
	_, err = database.DB.NamedExec(`INSERT INTO tags (tag) VALUES (:tag) ON CONFLICT (tag) DO NOTHING`, tagsStruct)
	if err != nil {
		// Print error instead of returning, duplicate value error is alright
		fmt.Println(err)
	}

	for _, v := range tags {
		// Copy postID and tag_id where the tag == tag value, and insert it into posts_tags
		_, err = database.DB.Exec(`INSERT INTO posts_tags (post_id, tag_id) SELECT $1, tag_id FROM tags WHERE tag=$2`, postID, v)
		if err != nil {
			// this err needs to be returned because it's not normal
			return err
		}
	}

	return nil
}

func GetTags(postID string) (JoinPost, error) {
	var t string
	var p JoinPost

	rows, err := database.DB.Query(`SELECT tags.tag
									FROM(SELECT posts_tags.post_id, posts_tags.tag_id
										FROM posts_tags
										WHERE posts_tags.post_id=$1) AS posts_tags
										LEFT JOIN tags ON posts_tags.tag_id=tags.tag_id`, postID)
	if err != nil {
		return p, err
	}

	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&t); err != nil {
			return p, err
		}

		p.Tags = append(p.Tags, t)
	}

	return p, nil
}

func EditTags(postID string, tags []string) error {
	var tag Tag
	validTags := []Tag{}
	var err error

	for _, v := range tags {
		if pass := ValidateTags(v); pass {
			tag.Tag = v
			validTags = append(validTags, tag)
		}
	}
	fmt.Println("Valid tags: ", validTags)
	fmt.Println("Number of valid tags: ", len(validTags))

	// Even if len(validTags) == 0, I can't end the func now, because I'll need to delete all the posts.
	// We can stop when needing to insert tags.

	if err = InsertTags(validTags); err != nil {
		// Print error instead of returning, duplicate value error is alright
		fmt.Println(err)
	}

	postsTags, err := GetPostIDTagIDTag(postID)
	if err != nil {
		return err
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

func InsertTags(validTags []Tag) error {
	if len(validTags) > 0 {
		_, err := database.DB.NamedExec(`INSERT INTO tags (tag) VALUES (:tag) ON CONFLICT (tag) DO NOTHING`, validTags)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertPostTags(postID string, validTags []Tag, postsTags []PostTag) error {
	var tagsToInsert []string

	for _, v := range validTags {
		if exists := containsPostTag(postsTags, v.Tag); !exists {
			tagsToInsert = append(tagsToInsert, v.Tag)
		}
	}

	fmt.Println("Tags to Insert: ", tagsToInsert)
	fmt.Println("Number to insert: ", len(tagsToInsert))

	// insert into posts_tags
	if len(tagsToInsert) > 0 {
		for _, v := range tagsToInsert {
			_, err := database.DB.Exec(`INSERT INTO posts_tags (post_id, tag_id) SELECT $1, tag_id FROM tags WHERE tag=$2`, postID, v)
			if err != nil {
				return err
			}
			fmt.Printf("Successfully saved: \"%s\" into \"%s\"\n", v, postID)
		}
	}
	return nil
}

func GetPostIDTagIDTag(postID string) ([]PostTag, error) {
	var pt PostTag
	var postsTags []PostTag
	// Grab all the existing tags from post using postID
	rows, err := database.DB.Query(`SELECT posts_tags.post_id, posts_tags.tag_id, tags.tag FROM posts_tags LEFT JOIN tags ON posts_tags.tag_id = tags.tag_id WHERE post_id=$1`, postID)
	if err != nil {
		return postsTags, err
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&pt.PostID, &pt.TagID, &pt.Tag); err != nil {
			return postsTags, err
		}

		postsTags = append(postsTags, pt)
		fmt.Println("Tags in post now: ", pt.Tag)
	}

	return postsTags, nil
}

var disallowed = [33]string{
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
}

func ValidateTags(tag string) (passed bool) {
	for _, v := range disallowed {
		if strings.Contains(tag, v) {
			return false
		}
	}
	return true
}

func DeleteUnwantedTags(inputTags []string, postsTags []PostTag) error {
	// Loop through postTags to find tags that are not in current user input, then mark those for deletion.
	var tagsToDeleteID []int
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
				return err
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

func containsPostTag(a []PostTag, s string) bool {
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
	ID, err := TitleToID(title)
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

func GetPost(postID string, currentUser string) (JoinPost, error) {
	var p JoinPost
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
		return p, err
	}
	defer row.Close()

	for row.Next() {
		if err := row.Scan(&p.PostID, &p.PostTitle, &p.UserID, &p.Description, &p.Protected, &p.CreatedAt, &p.Mood, &p.CurrentUserLike, &p.TagsNullString); err != nil {
			return p, err
		}

		if p.TagsNullString.Valid {
			p.Tags = strings.Split(p.TagsNullString.String, ",")
		} else {
			p.Tags = []string{}
		}

		if p.CurrentUserLike.Valid {
			p.CurrentUserLikeString = strconv.FormatInt(p.CurrentUserLike.Int64, 10)
		} else {
			p.CurrentUserLikeString = "0"
		}
	}

	p.CreatedAtProcessed, err = ConvertDate(p.CreatedAt)
	if err != nil {
		return p, err
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
				return score, err
			}
			score = 1
			return score, nil
		} else {
			return score, err
		}
	}

	_, err = database.DB.Exec("DELETE FROM posts_likes WHERE post_id=$1 AND user_id=$2", postID, currentUser)
	if err != nil {
		return score, err
	}
	score = 0

	return score, nil
}

func EditPostDescription(postID string, description string) error {
	_, err := database.DB.Exec("UPDATE posts SET description=$1 WHERE post_id=$2", description, postID)
	if err != nil {
		return err
	}
	return nil
}

func DeletePost(postID string, username string) error {
	var u string
	if err := database.DB.QueryRow("SELECT user_id FROM posts WHERE post_id=$1", postID).Scan(&u); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("error: cannot find user_id with given postID")
		}
		return err
	}

	fmt.Println("Owner of post: ", u)

	if u != username {
		return errors.New("error: logged in user is not owner of post")
	}

	if _, err := database.DB.Exec("DELETE FROM posts WHERE post_id=$1", postID); err != nil {
		return err
	}

	return nil
}

func EditMood(postID string, mood string) error {
	if err := ValidateMood(mood); err != nil {
		return err
	}

	_, err := database.DB.Exec("UPDATE posts SET mood=$1 WHERE post_id=$2", mood, postID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func ValidateMood(mood string) error {
	allowedMoods := [6]string{"Elated", "Happy", "Neutral", "Sad", "Upset", "Angry"}

	res := false
	for _, v := range allowedMoods {
		v = strings.ToUpper(v)

		if strings.Contains(v, strings.ToUpper(mood)) {
			res = true
			break
		}
	}
	if !res {
		err := errors.New("new mood is not in allowed list")
		return err
	}

	return nil
}

func TitleToID(title string) (string, error) {
	ss := strings.Fields(title)
	title = strings.Join(ss, " ")
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

	ID := r.Replace(strings.ToLower(title))
	fmt.Println(ID)

	return ID, nil
}
