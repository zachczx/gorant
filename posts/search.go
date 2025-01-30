package posts

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"gorant/database"
	"gorant/upload"
	"gorant/users"

	"github.com/google/uuid"
)

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

type UserStats struct {
	PostsCount    int
	CommentsCount int
	RepliesCount  int
}

func (u UserStats) PostsCountString() string {
	return strconv.Itoa(u.PostsCount)
}

func (u UserStats) CommentsCountString() string {
	return strconv.Itoa(u.CommentsCount)
}

func (u UserStats) RpliesCountString() string {
	return strconv.Itoa(u.RepliesCount)
}

func Search(query string, coverage string, sort string) ([]SearchComment, error) {
	var results []SearchComment
	var rows *sql.Rows
	var err error

	sql := `SELECT comments.comment_id, comments.user_id, users.preferred_name, comments.content, comments.created_at, comments.post_id, posts.post_title, comments.file_id FROM comments
			LEFT JOIN users
			ON comments.user_id = users.user_id
			LEFT JOIN posts
			ON posts.post_id = comments.post_id` + ` ` // Adding space here to be more obvious in future there's a space
	if sort == "recent" {
		if coverage == "posts" {
			sql = sql + `WHERE posts.ts @@ websearch_to_tsquery('english', $1) ORDER BY comments.created_at DESC`
		}
		if coverage == "comments" {
			sql = sql + `WHERE comments.ts @@ websearch_to_tsquery('english', $1) ORDER BY comments.created_at DESC`
		}
	}
	if sort == "relevance" {
		if coverage == "posts" {
			sql = sql + `WHERE posts.ts @@ websearch_to_tsquery('english', $1)`
		}
		if coverage == "comments" {
			sql = sql + `WHERE comments.ts @@ websearch_to_tsquery('english', $1)`
		}
	}
	rows, err = database.DB.Query(sql, query)
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

// Displaying user posts.
func GetUserPosts(userID string, p int) (PostCollection, bool, error) {
	var posts PostCollection
	var endOfList bool

	limit := 5
	var offset int
	if p > 0 {
		offset = (p - 1) * limit
	} else {
		offset = 0
	}

	query := `SELECT posts.post_id, posts.user_id, posts.post_title, posts.description, posts.protected, posts.created_at, posts.mood, users.preferred_name, comments.comments_cnt, replies.replies_cnt, posts_likes.likes_cnt, posts_tags.tags
			FROM posts
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
			WHERE posts.user_id=$1
			ORDER BY posts.created_at DESC
			LIMIT $2 OFFSET $3`

	rows, err := database.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, endOfList, fmt.Errorf("error executing list-post-filter query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.Description, &p.Protected, &p.CreatedAt.Time, &p.Mood, &p.PreferredName, &p.PostStats.CommentsCount, &p.PostStats.RepliesCount, &p.PostStats.LikesCount, &p.Tags.TagsNullString); err != nil {
			return nil, endOfList, fmt.Errorf("error scanning list-post-filter: %w", err)
		}
		if p.Tags.TagsNullString.Valid {
			p.Tags.Tags = strings.Split(p.Tags.TagsNullString.String, ",")
		} else {
			p.Tags.Tags = []string{}
		}
		posts = append(posts, p)
	}
	endOfList = checkEndOfList(posts, limit)

	return posts, endOfList, nil
}

func checkEndOfList(posts PostCollection, limit int) bool {
	return len(posts) < limit
}

func GetUserEngagementCount(currentUser *users.User) (UserStats, error) {
	var userStats UserStats
	var err error
	userStats.PostsCount, err = GetUserPostCount(currentUser)
	if err != nil {
		return userStats, fmt.Errorf("error: %w", err)
	}
	userStats.CommentsCount, err = GetUserCommentCount(currentUser)
	if err != nil {
		return userStats, fmt.Errorf("error: %w", err)
	}
	return userStats, nil
}

func GetUserPostCount(currentUser *users.User) (int, error) {
	var postsCount int
	err := database.DB.QueryRow(`SELECT COUNT(1) AS posts_cnt 
								FROM posts 
								WHERE posts.user_id=$1 
								GROUP BY posts.user_id;`, currentUser.UserID).Scan(&postsCount)
	if err != nil {
		if err == sql.ErrNoRows {
			postsCount = 0
			return postsCount, fmt.Errorf("no user posts found: %w", err)
		}
		return postsCount, fmt.Errorf("error: GetUserPostCount: %w", err)
	}
	return postsCount, nil
}

func GetUserCommentCount(currentUser *users.User) (int, error) {
	var commentsCount int
	err := database.DB.QueryRow(`SELECT COUNT(1) AS comment_cnt 
								FROM comments 
								WHERE comments.user_id=$1 
								GROUP BY comments.user_id;`, currentUser.UserID).Scan(&commentsCount)
	if err != nil {
		if err == sql.ErrNoRows {
			commentsCount = 0
			return commentsCount, fmt.Errorf("no user comments found: %w", err)
		}
		return commentsCount, fmt.Errorf("error: GetUserCommentCount: %w", err)
	}
	return commentsCount, nil
}
