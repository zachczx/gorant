package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib" // Pg driver
)

var DB *sqlx.DB

type resetTables []resetSQL

type resetIndexes []resetSQL

type insertBase []resetSQL

type resetSQL struct {
	name  string
	query string
}

// Structure:
//
// Posts + Likes
// --> Comments + Votes + Files
// -----> Replies

var tables = resetTables{
	{name: "users", query: `CREATE TABLE users (user_id VARCHAR(255) PRIMARY KEY, email VARCHAR(100) NOT NULL, preferred_name VARCHAR(255) DEFAULT '', contact_me INT DEFAULT 1, avatar VARCHAR(255) DEFAULT 'default', sort_comments VARCHAR(15) DEFAULT 'upvote;desc');`},
	{name: "posts", query: `CREATE TABLE posts (post_id VARCHAR(255) PRIMARY KEY, post_title VARCHAR(255) NOT NULL, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, description VARCHAR(255) DEFAULT '', protected INT DEFAULT 0, created_at TIMESTAMPTZ, mood VARCHAR(15) DEFAULT 'neutral', ts tsvector GENERATED ALWAYS AS (to_tsvector('english', post_title)) STORED);`},
	{name: "files", query: `CREATE TABLE files (file_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, file_key VARCHAR(2000) NOT NULL, file_thumbnail_key VARCHAR(2000), file_store VARCHAR(255) NOT NULL, file_bucket VARCHAR(255) NOT NULL, file_base_url VARCHAR(2000) NOT NULL, uploaded_at TIMESTAMPTZ);`},
	{name: "comments", query: `CREATE TABLE comments (comment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, content TEXT, created_at TIMESTAMPTZ, post_id VARCHAR(255) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE, file_id UUID REFERENCES files(file_id) ON DELETE SET NULL, ts tsvector GENERATED ALWAYS AS (to_tsvector('english', content)) STORED);`},
	{name: "posts_likes", query: `CREATE TABLE posts_likes (like_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE, post_id VARCHAR(255) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE, score INT);`},
	{name: "comments_votes", query: `CREATE TABLE comments_votes (vote_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, comment_id UUID REFERENCES comments(comment_id) ON DELETE CASCADE ON UPDATE CASCADE, score INT);`},
	{
		name:  "replies",
		query: `CREATE TABLE replies (reply_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, content TEXT, created_at TIMESTAMPTZ, post_id VARCHAR(255) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL, comment_id UUID REFERENCES comments(comment_id) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL, file_id UUID REFERENCES files(file_id) ON DELETE SET NULL, ts tsvector GENERATED ALWAYS AS (to_tsvector('english', content)) STORED);`,
	},
	{name: "tags", query: `CREATE TABLE tags (tag_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), tag VARCHAR(30) UNIQUE NOT NULL);`},
	{name: "posts_tags", query: `CREATE TABLE posts_tags (posts_tags_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), post_id VARCHAR(255) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE, tag_id UUID REFERENCES tags(tag_id) ON DELETE CASCADE ON UPDATE CASCADE);`},
	{name: "instant_posts", query: `CREATE TABLE instant_posts (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), title VARCHAR(255) NOT NULL, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, created_at TIMESTAMPTZ)`},
	{name: "instant_comments", query: `CREATE TABLE instant_comments (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), instant_post_id UUID REFERENCES instant_posts(id) ON DELETE CASCADE, title VARCHAR(255) DEFAULT '', content TEXT NOT NULL, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE CASCADE, created_at TIMESTAMPTZ)`},
}

var indexes = resetIndexes{
	{name: "GIN_index_posts_ts-tsvector", query: `CREATE INDEX idx_posts_ts ON posts USING GIN (ts);`},
	{name: "GIN_index_posts_gist_trgm", query: `CREATE INDEX idx_posts_gist_trgm ON posts USING GIST (post_title gist_trgm_ops, post_id);`},
	{name: "idx_comments_post_id", query: `CREATE INDEX idx_comments_post_id ON comments (post_id);`},
	{name: "GIN_index_comments_ts-tsvector", query: `CREATE INDEX idx_comments_ts ON comments USING GIN (ts);`},
	{name: "GIN_index_replies_ts-tsvector", query: `CREATE INDEX idx_replies_ts ON replies USING GIN (ts);`},
	{name: "idx_posts_likes_post_id", query: `CREATE INDEX idx_posts_likes_post_id ON posts_likes (post_id);`},
	{name: "idx_comments_votes_comment_id", query: `CREATE INDEX idx_comments_votes_comment_id ON comments_votes (comment_id);`},
}

var pgExtensions = insertBase{
	{name: "pg_trgm", query: `CREATE EXTENSION IF NOT EXISTS pg_trgm;`},
	{name: "btree_gist", query: `CREATE EXTENSION btree_gist;`},
}

var baseQueries = insertBase{
	{name: "created_user_anonymous", query: `INSERT INTO users (user_id, email, preferred_name) VALUES ('anonymous@rantkit.com', 'anonymous@rantkit.com', 'anonymous')`},
}

func (t *resetTables) dropTables() error {
	for _, v := range *t {
		q := fmt.Sprintf(`DROP TABLE IF EXISTS %v CASCADE;`, v.name)
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("error dropping table: %v: %w", v.name, err)
		}
		fmt.Printf("Dropped table: %v\r\n", v.name)
	}
	return nil
}

func executeSQL(rs []resetSQL, operation string) error {
	for _, v := range rs {
		if _, err := DB.Exec(v.query); err != nil {
			return fmt.Errorf("error %v table: %v: %w", operation, v.name, err)
		}
		fmt.Printf("Created table: %v\r\n", v.name)
	}
	return nil
}

func (t resetTables) create() error {
	return executeSQL(t, "creating")
}

func (t resetIndexes) create() error {
	return executeSQL(t, "creating")
}

func (t insertBase) create() error {
	return executeSQL(t, "creating")
}

func Reset() error {
	if err := pgExtensions.create(); err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	if err := tables.dropTables(); err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	if err := tables.create(); err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	if err := indexes.create(); err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	if err := baseQueries.create(); err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	fmt.Println("Reset completed successfully!")
	return nil
}
