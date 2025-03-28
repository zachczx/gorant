package posts

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"gorant/database"
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
}

type SearchPost struct {
	ID                 string    `db:"post_id"`
	Title              string    `db:"post_title"`
	UserID             string    `db:"user_id"`
	Description        string    `db:"description"`
	CreatedAt          CreatedAt `db:"created_at"`
	Initials           string
	PreferredName      string `db:"preferred_name"`
	Avatar             string `db:"avatar"`
	CommentStats       CommentStats
	CreatedAtProcessed string
	AvatarPath         string
}

type SearchItem interface {
	GetPostID() string
	GetCommentID() uuid.UUID
	GetPostTitle() string
	GetContent() string
	GetCreatedAt() string
	GetPreferredName() string
}

func (s SearchPost) GetPostID() string {
	return s.ID
}

func (s SearchPost) GetCommentID() uuid.UUID {
	return uuid.Nil
}

func (s SearchPost) GetPostTitle() string {
	return s.Title
}

func (s SearchPost) GetContent() string {
	return ""
}

func (s SearchPost) GetCreatedAt() string {
	return s.CreatedAt.Process()
}

func (s SearchPost) GetPreferredName() string {
	return s.PreferredName
}

func (s SearchComment) GetPostID() string {
	return s.PostID
}

func (s SearchComment) GetCommentID() uuid.UUID {
	return s.ID
}

func (s SearchComment) GetPostTitle() string {
	return s.PostTitle
}

func (s SearchComment) GetContent() string {
	return s.Content
}

func (s SearchComment) GetCreatedAt() string {
	return s.CreatedAt.Process()
}

func (s SearchComment) GetPreferredName() string {
	return s.PreferredName
}

func (c SearchComment) IDString() string {
	return c.ID.String()
}

type UserStats struct {
	UserID        string
	PostsCount    sql.NullInt64
	CommentsCount sql.NullInt64
	RepliesCount  sql.NullInt64
}

func (u UserStats) PostsCountString() string {
	if !u.PostsCount.Valid {
		return "0"
	}
	return strconv.FormatInt(u.PostsCount.Int64, 10)
}

func (u UserStats) CommentsCountString() string {
	if !u.CommentsCount.Valid {
		return "0"
	}
	return strconv.FormatInt(u.CommentsCount.Int64, 10)
}

func (u UserStats) RepliesCountString() string {
	if !u.RepliesCount.Valid {
		return "0"
	}
	return strconv.FormatInt(u.RepliesCount.Int64, 10)
}

func SearchPosts(query string, sort string) ([]SearchItem, error) {
	sql := `SELECT DISTINCT ON (posts.post_id) posts.post_id, posts.post_title, posts.user_id, posts.description, posts.created_at, users.preferred_name FROM posts
				LEFT JOIN users
				ON posts.user_id = users.user_id
				WHERE posts.ts @@ websearch_to_tsquery('english', $1)` + ` ` // Adding space here to be more obvious in future there's a space
	if sort == "recent" {
		sql = sql + `ORDER BY posts.post_id, posts.created_at DESC`
	}
	rows, err := database.DB.Query(sql, query)
	if err != nil {
		return nil, fmt.Errorf("error querying db for search post: %w", err)
	}
	defer rows.Close()
	var results []SearchItem
	var row SearchPost
	for rows.Next() {
		if err := rows.Scan(&row.ID, &row.Title, &row.UserID, &row.Description, &row.CreatedAt.Time, &row.PreferredName); err != nil {
			return results, fmt.Errorf("error scanning for search posts: %w", err)
		}
		results = append(results, row)
	}
	return results, nil
}

func SearchComments(query string, sort string) ([]SearchItem, error) {
	sql := `SELECT comments.comment_id, comments.user_id, users.preferred_name, comments.content, comments.created_at, comments.post_id, posts.post_title FROM comments
			LEFT JOIN users
			ON comments.user_id = users.user_id
			LEFT JOIN posts
			ON posts.post_id = comments.post_id
			WHERE comments.ts @@ websearch_to_tsquery('english', $1)` + ` ` // Adding space here to be more obvious in future there's a space
	if sort == "recent" {
		sql = sql + `ORDER BY comments.created_at DESC`
	}
	rows, err := database.DB.Query(sql, query)
	if err != nil {
		return nil, fmt.Errorf("error querying db for search comments: %w", err)
	}
	defer rows.Close()
	var results []SearchItem
	var row SearchComment
	for rows.Next() {
		if err := rows.Scan(&row.ID, &row.UserID, &row.PreferredName, &row.Content, &row.CreatedAt.Time, &row.PostID, &row.PostTitle); err != nil {
			return results, fmt.Errorf("error scanning db for search comments: %w", err)
		}
		results = append(results, row)
	}
	return results, nil
}

// Displaying user posts.
func GetUserPosts(userID string, p int) (PostCollection, bool, error) {
	var posts PostCollection
	var endOfList bool

	limit := 10
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

func GetEngagementStats(currentUser *users.User) (UserStats, error) {
	var stats UserStats

	err := database.DB.QueryRow(`SELECT posts.user_id, posts.posts_cnt, comments.comments_cnt, replies.replies_cnt
							FROM(SELECT user_id, COUNT(1) AS posts_cnt
								FROM posts
								WHERE user_id=$1
								GROUP BY user_id) AS posts
								LEFT JOIN(SELECT comments.user_id, COUNT(1) AS comments_cnt
										FROM comments
										GROUP BY comments.user_id) AS comments ON posts.user_id=comments.user_id
								LEFT JOIN(SELECT replies.user_id, COUNT(1) as replies_cnt
										FROM replies
										GROUP BY replies.user_id) AS replies ON posts.user_id=replies.user_id`, currentUser.UserID).
		Scan(&stats.UserID, &stats.PostsCount, &stats.CommentsCount, &stats.RepliesCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return stats, fmt.Errorf("error: no engagement stats: %w", err)
		}
		return stats, fmt.Errorf("error: %w", err)
	}
	return stats, nil
}
