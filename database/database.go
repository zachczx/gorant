package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sqlx.DB

func Reset() error {
	var err error
	// Postgres
	_, err = DB.Exec(`DROP TABLE IF EXISTS comments CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: comments")
		return err
	}
	_, err = DB.Exec(`DROP TABLE IF EXISTS comments_votes CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: comments_votes")
		return err
	}
	_, err = DB.Exec(`DROP TABLE IF EXISTS posts CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: posts")
		return err
	}
	_, err = DB.Exec(`DROP TABLE IF EXISTS files CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: files")
		return err
	}
	_, err = DB.Exec(`DROP TABLE IF EXISTS users CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: users")
		return err
	}
	_, err = DB.Exec(`DROP TABLE IF EXISTS posts_likes CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: posts_likes")
		return err
	}

	_, err = DB.Exec(`DROP TABLE IF EXISTS tags CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: tags")
		return err
	}

	_, err = DB.Exec(`DROP TABLE IF EXISTS posts_tags CASCADE;`)
	if err != nil {
		fmt.Println("Error dropping table: posts_tags")
		return err
	}

	_, err = DB.Exec(`DROP TABLE IF EXISTS instant_posts CASCADE`)
	if err != nil {
		fmt.Println("Error dropping table: instant_posts")
		return err
	}

	_, err = DB.Exec(`DROP TABLE IF EXISTS instant_comments CASCADE`)
	if err != nil {
		fmt.Println("Error dropping table: instant_comments")
		return err
	}

	// Users

	_, err = DB.Exec(`CREATE TABLE users (user_id VARCHAR(255) PRIMARY KEY, email VARCHAR(100) NOT NULL, preferred_name VARCHAR(255) DEFAULT '', contact_me INT DEFAULT 1, avatar VARCHAR(255) DEFAULT 'default', sort_comments VARCHAR(15) DEFAULT 'upvote;desc');`)
	if err != nil {
		fmt.Println("Error creating table: users")
		return err
	}
	fmt.Println("Created table: users")

	// Posts

	_, err = DB.Exec(`CREATE TABLE posts (post_id VARCHAR(255) PRIMARY KEY, post_title VARCHAR(255) NOT NULL, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, description VARCHAR(255) DEFAULT '', protected INT DEFAULT 0, created_at TIMESTAMPTZ, mood VARCHAR(15) DEFAULT 'neutral');`)
	if err != nil {
		fmt.Println("Error creating table: posts")
		return err
	}
	fmt.Println("Created table: posts")

	// Files

	_, err = DB.Exec(`CREATE TABLE files (file_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), file_key VARCHAR(2000), file_store VARCHAR(255),file_bucket VARCHAR(255));`)
	if err != nil {
		fmt.Println("Error creating table: files")
		return err
	}
	fmt.Println("Created table: files")

	// Comments

	_, err = DB.Exec(`CREATE TABLE comments (comment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, content TEXT, created_at TIMESTAMPTZ, post_id VARCHAR(255) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE, file_id UUID REFERENCES files(file_id) ON DELETE SET NULL);`)
	if err != nil {
		fmt.Println("Error creating table: comments")
		return err
	}
	fmt.Println("Created table: comments")

	_, err = DB.Exec(`CREATE INDEX idx_comments_post_id ON comments (post_id);`)
	if err != nil {
		fmt.Println("Error creating index: comments.post_id ")
		return err
	}
	fmt.Println("Created index: comments")

	// Posts Likes

	_, err = DB.Exec(`CREATE TABLE posts_likes (like_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE, post_id VARCHAR(255) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE, score INT);`)
	if err != nil {
		fmt.Println("Error creating table: posts_likes")
		return err
	}
	fmt.Println("Created table: posts_likes")

	_, err = DB.Exec(`CREATE INDEX idx_posts_likes_post_id ON posts_likes (post_id);`)
	if err != nil {
		fmt.Println("Error creating index: idx_posts_likes_post_id")
		return err
	}
	fmt.Println("Created index: idx_posts_likes_post_id")

	// Comments Votes
	_, err = DB.Exec(`CREATE TABLE comments_votes (vote_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, comment_id UUID REFERENCES comments(comment_id) ON DELETE CASCADE ON UPDATE CASCADE, score INT);`)
	if err != nil {
		fmt.Println("Error creating table: comments_votes")
		return err
	}
	fmt.Println("Created table: comments_votes")

	_, err = DB.Exec(`CREATE INDEX idx_comments_votes_comment_id ON comments_votes (comment_id);`)
	if err != nil {
		fmt.Println("Error creating index: comments_votes.comment_id")
		return err
	}
	fmt.Println("Created index: idx_comments_votes_comment_id")

	// Tags
	_, err = DB.Exec(`CREATE TABLE tags (tag_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), tag VARCHAR(30) UNIQUE NOT NULL);`)
	if err != nil {
		fmt.Println("Error creating table: tags")
		return err
	}
	fmt.Println("Created table: tags")

	// Intermediate Tags table
	_, err = DB.Exec(`CREATE TABLE posts_tags (posts_tags_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), post_id VARCHAR(255) REFERENCES posts(post_id) ON DELETE CASCADE ON UPDATE CASCADE, tag_id UUID REFERENCES tags(tag_id) ON DELETE CASCADE ON UPDATE CASCADE);`)
	if err != nil {
		fmt.Println("Error creating table: posts_tags")
		return err
	}
	fmt.Println("Created table: posts_tags")

	// Instant Posts
	_, err = DB.Exec(`CREATE TABLE instant_posts (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), title VARCHAR(255) NOT NULL, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE SET NULL ON UPDATE CASCADE, created_at TIMESTAMPTZ)`)
	if err != nil {
		return err
	}
	fmt.Println("Created table: instant_posts")

	// Instant Comments
	_, err = DB.Exec(`CREATE TABLE instant_comments (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), instant_post_id UUID REFERENCES instant_posts(id) ON DELETE CASCADE, title VARCHAR(255) DEFAULT '', content TEXT NOT NULL, user_id VARCHAR(255) REFERENCES users(user_id) ON DELETE CASCADE, created_at TIMESTAMPTZ)`)
	if err != nil {
		return err
	}
	fmt.Println("Created table: instant_comments")

	_, err = DB.Exec(`INSERT INTO users (user_id, email, preferred_name) VALUES ('anonymous@rantkit.com', 'anonymous@rantkit.com', 'anonymous')`)
	if err != nil {
		fmt.Println("Error creating user: anonymous")
		return err
	}
	fmt.Println("Created user: anonymous")

	return nil
}
