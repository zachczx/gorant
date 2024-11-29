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
	PostID             string `db:"post_id"`
	PostTitle          string `db:"post_title"`
	UserID             string `db:"user_id"`
	Description        string `db:"description"`
	Protected          int    `db:"protected"`
	CreatedAt          string `db:"created_at"`
	Mood               string `db:"mood"`
	CreatedAtProcessed string
}

type PostLike struct {
	ID     int    `db:"like_id"`
	UserID string `db:"user_id"`
	PostID string `db:"post_id"`
	Score  string `db:"score"`
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
}

const regex string = `^[A-Za-z0-9 _!.\$\/\\|()\[\]=` + "`" + `{<>?@#%^&*—:;'"+\-"]+$`

func ListPosts() ([]JoinPost, error) {
	db, err := database.Connect()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood, users.preferred_name, comments_cnt, likes_cnt FROM posts
							LEFT JOIN users
							ON users.user_id = posts.user_id
              
							LEFT JOIN (SELECT comments.post_id, COUNT(1) AS comments_cnt FROM comments GROUP BY comments.post_id) AS comments
							ON comments.post_id = posts.post_id

							LEFT JOIN (SELECT post_id, COUNT(1) as likes_cnt FROM posts_likes GROUP BY posts_likes.post_id) as posts_likes
							ON posts.post_id = posts_likes.post_id;`)
	if err != nil {
		fmt.Println("Error executing query: ", err)
		return nil, err
	}

	defer rows.Close()

	var posts []JoinPost

	for rows.Next() {
		var p JoinPost

		if err := rows.Scan(&p.PostID, &p.UserID, &p.PostTitle, &p.Description, &p.Protected, &p.CreatedAt, &p.Mood, &p.PreferredName, &p.CommentsCount, &p.LikesCount); err != nil {
			fmt.Println("Error scanning")
			return nil, err
		}

		if p.CommentsCount.Valid {
			p.CommentsCountString = strconv.FormatInt(p.CommentsCount.Int64, 10)
		} else {
			p.CommentsCountString = "0"
		}

		if p.LikesCount.Valid {
			p.LikesCountString = strconv.FormatInt(p.LikesCount.Int64, 10)
		} else {
			p.LikesCountString = "0"
		}

		p.CreatedAtProcessed, err = ConvertDate(p.CreatedAt)
		if err != nil {
			fmt.Println(err)
		}

		posts = append(posts, p)
	}

	return posts, nil
}

func NewPost(postID string, postTitle string, username string, mood string) error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	t := time.Now().Format(time.RFC3339)

	_, err = db.Exec("INSERT INTO posts (post_id, post_title, user_id, created_at, mood) VALUES ($1, $2, $3, $4, $5)", postID, postTitle, username, t, mood)
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

func GetPost(postID string, currentUser string) (JoinPost, error) {
	db, err := database.Connect()
	if err != nil {
		return JoinPost{}, err
	}

	var p JoinPost
	row, err := db.Query(`SELECT posts.post_id, posts.post_title, posts.user_id, posts.description, posts.protected, posts.created_at, posts.mood, posts_likes.score FROM posts  
							LEFT JOIN (SELECT * FROM posts_likes) as posts_likes 
							ON posts.post_id = posts_likes.post_id
							AND posts_likes.user_id=$2

							WHERE posts.post_id=$1`, postID, currentUser)
	if err != nil {
		return p, err
	}
	defer row.Close()

	for row.Next() {
		if err := row.Scan(&p.PostID, &p.PostTitle, &p.UserID, &p.Description, &p.Protected, &p.CreatedAt, &p.Mood, &p.CurrentUserLike); err != nil {
			return p, err
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
	db, err := database.Connect()

	var score int

	if err != nil {
		return score, err
	}

	var exists string
	err = db.QueryRow("SELECT score FROM posts_likes WHERE post_id=$1 AND user_id=$2", postID, currentUser).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			if _, err = db.Exec("INSERT INTO posts_likes (user_id, post_id, score) VALUES ($1, $2, 1)", currentUser, postID); err != nil {
				return score, err
			}
			score = 1
			return score, nil
		} else {
			return score, err
		}
	}

	_, err = db.Exec("DELETE FROM posts_likes WHERE post_id=$1 AND user_id=$2", postID, currentUser)
	if err != nil {
		return score, err
	}
	score = 0

	return score, nil
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
