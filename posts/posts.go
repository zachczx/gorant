package posts

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorant/database"

	"github.com/rezakhademix/govalidator/v2"
)

type Post struct {
	PostID      string `db:"post_id"`
	PostTitle   string `db:"post_title"`
	UserID      string `db:"user_id"`
	Description string `db:"description"`
	Protected   int    `db:"protected"`
	CreatedAt   string `db:"created_at"`
	Mood        string `db:"mood"`
}

type JoinPost struct {
	PostID              string `db:"post_id"`
	PostTitle           string `db:"post_title"`
	UserID              string `db:"user_id"`
	Description         string `db:"description"`
	Protected           int    `db:"protected"`
	CreatedAt           string `db:"created_at"`
	Mood                string `db:"mood"`
	PreferredName       string
	CommentsCount       sql.NullInt64 `db:"comments_cnt"`
	CommentsCountString string
}

const regex string = `^[A-Za-z0-9 _!.\$\/\\|()\[\]=` + "`" + `{<>?@#%^&*—:;'"+\-"]+$`

func ListPosts() ([]JoinPost, error) {
	db, err := database.Connect()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood, users.preferred_name, comments_cnt FROM posts
							LEFT JOIN users
							ON users.user_id = posts.user_id
              
							LEFT JOIN (SELECT comments.post_id, COUNT(1) AS comments_cnt FROM comments GROUP BY comments.post_id) AS comments
							ON comments.post_id = posts.post_id
							;`)
	if err != nil {
		fmt.Println("Error executing query: ", err)
		return nil, err
	}

	defer rows.Close()

	var posts []JoinPost

	for rows.Next() {
		var p JoinPost

		if err := rows.Scan(&p.PostID, &p.UserID, &p.PostTitle, &p.Description, &p.Protected, &p.CreatedAt, &p.Mood, &p.PreferredName, &p.CommentsCount); err != nil {
			fmt.Println("Error scanning")
			return nil, err
		}

		if p.CommentsCount.Valid {
			p.CommentsCountString = strconv.FormatInt(p.CommentsCount.Int64, 10)
		} else {
			p.CommentsCountString = "0"
		}

		posts = append(posts, p)
	}

	return posts, nil
}

func NewPost(postID string, postTitle string, username string) error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	t := time.Now().String()

	_, err = db.Exec("INSERT INTO posts (post_id, post_title, user_id, created_at) VALUES ($1, $2, $3, $4)", postID, postTitle, username, t)
	if err != nil {
		return err
	}

	return nil
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

	db, _ := database.Connect()

	// Generate ID to test
	ID, err := TitleToID(title)
	if err != nil {
		fmt.Println(err)
	}

	res, err := db.Query("SELECT post_id FROM posts WHERE post_id=$1;", ID)
	if err != nil {
		fmt.Println("Error executing query to verify post exists")
		fmt.Println(err)
	}

	defer res.Close()

	return res.Next(), ID
}

func GetPost(postID string, currentUser string) (Post, error) {
	db, err := database.Connect()
	if err != nil {
		return Post{}, err
	}

	var post Post
	if err := db.QueryRow("SELECT * FROM posts WHERE post_id=$1", postID).Scan(&post.PostID, &post.PostTitle, &post.UserID, &post.Description, &post.Protected, &post.CreatedAt, &post.Mood); err != nil {
		return post, err
	}

	return post, nil
}

func EditPostDescription(postID string, description string) error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE posts SET description=$1 WHERE post_id=$2", description, postID)
	if err != nil {
		return err
	}
	return nil
}

func DeletePost(postID string, username string) error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	var u string
	if err := db.QueryRow("SELECT user_id FROM posts WHERE post_id=$1", postID).Scan(&u); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("error: cannot find user_id with given postID")
		}
		return err
	}

	fmt.Println("Owner of post: ", u)

	if u != username {
		return errors.New("error: logged in user is not owner of post")
	}

	if _, err := db.Exec("DELETE FROM posts WHERE post_id=$1", postID); err != nil {
		return err
	}

	return nil
}

func EditMood(postID string, mood string) error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

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
		err = errors.New("new mood is not in allowed list")
		return err
	}

	_, err = db.Exec("UPDATE posts SET mood=$1 WHERE post_id=$2", mood, postID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func TitleToID(title string) (string, error) {
	ss := strings.Fields(title)
	title = strings.Join(ss, " ")
	r := strings.NewReplacer(
		" ", "-",
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
	)

	ID := r.Replace(strings.ToLower(title))
	fmt.Println(ID)

	return ID, nil
}
