package live

import (
	"fmt"
	"strconv"
	"time"

	"gorant/database"

	"github.com/google/uuid"
)

type InstantPost struct {
	ID        uuid.UUID `db:"id"`
	Title     string    `db:"title"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type InstantComment struct {
	ID            uuid.UUID `db:"id"`
	InstantPostID uuid.UUID `db:"instant_post_id"`
	Title         string    `db:"title"`
	Content       string    `db:"content"`
	UserID        string    `db:"user_id"`
	CreatedAt     time.Time `db:"created_at"`
	PreferredName string
}

func (instP *InstantPost) TitleInitials() string {
	return instP.Title[:2]
}

func (instC *InstantComment) PreferredNameInitials() string {
	return instC.PreferredName[:2]
}

func (instP *InstantPost) DateString() string {
	var s string
	var suffix string
	singleDay := 24.00
	singleHour := 1.00
	n := time.Now()
	diff := n.Sub(instP.CreatedAt).Hours()
	switch {
	case diff < singleHour:
		if n.Sub(instP.CreatedAt).Minutes() < 2 {
			suffix = " minute ago"
		} else {
			suffix = " minutes ago"
		}
		// Mins
		s = strconv.Itoa(int(n.Sub(instP.CreatedAt).Minutes())) + suffix
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
		s = strconv.Itoa(int(n.Sub(instP.CreatedAt).Hours()/singleDay)) + suffix
	default:
		fmt.Println("Something went wrong")
	}
	return s
}

func ListLivePosts() ([]InstantPost, error) {
	var instPosts []InstantPost
	var instP InstantPost
	rows, err := database.DB.Query(`SELECT * FROM instant_posts`)
	if err != nil {
		return instPosts, fmt.Errorf("error with querying for listlistposts(): %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&instP.ID, &instP.Title, &instP.UserID, &instP.CreatedAt); err != nil {
			return instPosts, fmt.Errorf("error with scanning listlistposts(): %w", err)
		}
		instPosts = append(instPosts, instP)
	}

	return instPosts, nil
}

func ListLiveComments() ([]InstantComment, error) {
	rows, err := database.DB.Query(`SELECT * FROM instant_comments ORDER BY created_at DESC`)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var instComments []InstantComment
	var instC InstantComment

	for rows.Next() {
		if err := rows.Scan(&instC.ID, &instC.InstantPostID, &instC.Title, &instC.Content, &instC.UserID, &instC.CreatedAt); err != nil {
			return instComments, fmt.Errorf("error with scanning listlivecomments(): %w", err)
		}
		instComments = append(instComments, instC)
	}
	return instComments, nil
}

func GetInstantPost(id uuid.UUID) (InstantPost, error) {
	var p InstantPost
	if err := database.DB.QueryRow(`SELECT id, title, user_id, created_at FROM instant_posts WHERE id=$1`, id).Scan(&p.ID, &p.Title, &p.UserID, &p.CreatedAt); err != nil {
		return p, fmt.Errorf("error fetching instant post: %w", err)
	}
	return p, nil
}

func ViewLivePost(id uuid.UUID) ([]InstantComment, error) {
	rows, err := database.DB.Query(`SELECT instant_comments.id, instant_comments.instant_post_id, instant_comments.title, instant_comments.content, instant_comments.user_id, instant_comments.created_at, users.preferred_name FROM instant_comments 
										LEFT JOIN users
										ON instant_comments.user_id = users.user_id
									WHERE instant_post_id=$1
									ORDER BY created_at DESC`, id)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var instComments []InstantComment
	var instC InstantComment

	for rows.Next() {
		if err := rows.Scan(&instC.ID, &instC.InstantPostID, &instC.Title, &instC.Content, &instC.UserID, &instC.CreatedAt, &instC.PreferredName); err != nil {
			return instComments, fmt.Errorf("error with scanning viewlivepost(): %w", err)
		}
		instComments = append(instComments, instC)
	}
	return instComments, nil
}

func CreateInstantPost(instP InstantPost) error {
	_, err := database.DB.Exec(`INSERT INTO instant_posts (title, user_id, created_at) VALUES ($1, $2, NOW())`, instP.Title, instP.UserID)
	if err != nil {
		return fmt.Errorf("error with inserting into instant_posts: %w", err)
	}
	return nil
}

func CreateInstantComment(instC InstantComment) error {
	_, err := database.DB.Exec(`INSERT INTO instant_comments (instant_post_id, content, user_id, created_at) VALUES ($1, $2, $3, NOW())`, instC.InstantPostID, instC.Content, instC.UserID)
	if err != nil {
		return fmt.Errorf("error with inserting into instant_comments: %w", err)
	}
	return nil
}
