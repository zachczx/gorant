package models

import "time"

type Posts struct {
	PostID      string `db:"post_id"`
	PostTitle   string `db:"post_title"`
	UserID      string `db:"user_id"`
	Description string `db:"description"`
	Protected   int    `db:"protected"`
	CreatedAt   time.Time
	Mood        string `db:"mood"`
}

type PostsLikes struct {
	LikeID int    `db:"like_id"`
	UserID string `db:"user_id"`
	PostID string `db:"post_id"`
	Score  int    `db:"score"`
}

type PostsTags struct {
	PostsTagsID int    `db:"posts_tags_id"`
	PostID      string `db:"post_id"`
	TagID       int    `db:"tag_id"`
}

type Tags struct {
	TagID int    `db:"tag_id"`
	Tag   string `db:"tag"`
}

type Comments struct {
	CommentID int       `db:"comment_id"`
	UserID    string    `db:"user_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	PostID    string    `db:"post_id"`
}

type CommentsVotes struct {
	VoteID    int `db:"vote_id"`
	UserID    int `db:"user_id"`
	CommentID int `db:"comment_id"`
	Score     int `db:"score"`
}

type User struct {
	UserID        string `db:"user_id"`
	Email         string `db:"email"`
	PreferredName string `db:"preferred_name"`
	ContactMe     int    `db:"contact_me"`
	Avatar        string `db:"avatar"`
	SortComments  string `db:"sort_comments"`
}

type InstantPost struct {
	ID        int       `db:"id"`
	Title     string    `db:"title"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type InstantComment struct {
	ID            int       `db:"id"`
	InstantPostID int       `db:"instant_post_id"`
	Title         string    `db:"title"`
	Content       string    `db:"content"`
	UserID        string    `db:"user_id"`
	CreatedAt     time.Time `db:"created_at"`
}
