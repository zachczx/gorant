package posts

import (
	"fmt"

	"gorant/database"
	"gorant/users"

	"github.com/google/uuid"
)

type Reply struct {
	ID            uuid.UUID
	UserID        string
	PreferredName string
	Avatar        string
	Content       string
	CreatedAt     CreatedAt
	PostID        string
	CommentID     uuid.UUID
	FileID        uuid.UUID
}

type ReplyCollection []Reply

func (reply *Reply) AvatarPath() string {
	avatar := users.ChooseAvatar(reply.Avatar)
	return avatar
}

func (replyCollection ReplyCollection) Map() map[uuid.UUID]ReplyCollection {
	replyMap := make(map[uuid.UUID]ReplyCollection)
	for _, v := range replyCollection {
		replyMap[v.CommentID] = append(replyMap[v.CommentID], v)
	}
	return replyMap
}

func (reply *Reply) Insert() (string, error) {
	var replyID string
	q := `INSERT INTO replies (user_id, content, created_at, post_id, comment_id) VALUES ($1, $2, NOW(), $3, $4) RETURNING reply_id`
	if err := database.DB.QueryRow(q, reply.UserID, reply.Content, reply.PostID, reply.CommentID).Scan(&replyID); err != nil {
		return replyID, fmt.Errorf("error: insert reply: %w", err)
	}
	return replyID, nil
}

func GetReplies(postID string) (ReplyCollection, error) {
	var replyCollection ReplyCollection
	q := `SELECT replies.reply_id, replies.user_id, users.preferred_name, users.avatar, replies.content, replies.created_at, replies.post_id, replies.comment_id FROM replies 
			LEFT JOIN users
			ON replies.user_id = users.user_id
			WHERE post_id=$1
			ORDER BY replies.created_at DESC;`
	rows, err := database.DB.Query(q, postID)
	if err != nil {
		return replyCollection, fmt.Errorf("error: replies query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var reply Reply
		if err := rows.Scan(&reply.ID, &reply.UserID, &reply.PreferredName, &reply.Avatar, &reply.Content, &reply.CreatedAt.Time, &reply.PostID, &reply.CommentID); err != nil {
			return replyCollection, fmt.Errorf("error: replies scan: %w", err)
		}

		replyCollection = append(replyCollection, reply)
	}
	return replyCollection, nil
}
