package users

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"gorant/database"

	"github.com/rezakhademix/govalidator/v2"
)

type User struct {
	UserID          string `db:"user_id"`
	Email           string `db:"email"`
	PreferredName   string `db:"preferred_name"`
	ContactMe       int    `db:"contact_me"`
	ContactMeString string
	Avatar          string `db:"avatar"`
	AvatarPath      string
	SortComments    string `db:"sort_comments"`
}

type Settings struct {
	PreferredName string
	ContactMe     string
	Avatar        string
	SortComments  string
}

func GetSettings(username string) (User, error) {
	db, err := database.Connect()
	if err != nil {
		fmt.Println("Error connecting to DB", err)
	}

	var s User
	if err := db.QueryRow("SELECT * FROM users WHERE user_id=$1", username).Scan(&s.UserID, &s.Email, &s.PreferredName, &s.ContactMe, &s.Avatar, &s.SortComments); err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Weird, no rows found!")
			return s, err
		}
	}
	s.ContactMeString = strconv.Itoa(s.ContactMe)
	s.AvatarPath = ChooseAvatar(s.Avatar)

	return s, nil
}

func Validate(s Settings) map[string](string) {
	v := govalidator.New()

	var ok bool = false
	avatarVals := []string{"default", "shiba", "cat", "parrot", "bulldog"}
	for _, v := range avatarVals {
		if v == s.Avatar {
			ok = true
			break
		}
	}
	fmt.Println("Avatar check: ", ok)

	v.RequiredString(s.PreferredName, "preferred_name", "Please enter a preferred name").MaxString(s.PreferredName, 255, "preferred_name", "Message is more than 255 characters.")
	v.CustomRule(ok, "avatar", "Unrecognized avatar")

	if v.IsFailed() {
		return v.Errors()
	}

	return nil
}

func SaveSettings(username string, s Settings) error {
	db, err := database.Connect()
	if err != nil {
		return err
	}

	// ContactMe is the opposite of the form value:
	// - Checking box (on) = Don't contact me = 0
	// - Not checking box ("") = Contact me = 1 = Default
	if s.ContactMe == "on" {
		s.ContactMe = "0"
	} else {
		s.ContactMe = "1"
	}

	_, err = db.Exec("UPDATE users SET preferred_name=$1, contact_me=$2, avatar=$3, sort_comments=$4 WHERE user_id=$5;", s.PreferredName, s.ContactMe, s.Avatar, s.SortComments, username)
	if err != nil {
		return err
	}

	return nil
}

func SaveSortComments(username string, s string) (string, error) {
	db, err := database.Connect()
	if err != nil {
		return s, err
	}

	switch s {
	case "upvote;desc", "upvote;asc", "date;desc", "date;asc":
		_, err = db.Exec("UPDATE users SET sort_comments=$1 WHERE user_id=$2;", s, username)
		if err != nil {
			return s, err
		}

	default:
		err = errors.New("unknown value")
		return s, err

	}

	return s, nil
}

// TODO avatar choice

func ChooseAvatar(c string) string {
	var s string
	switch c {
	case "shiba":
		s = "/static/images/avatars/avatar-shiba.webp"
	case "cat":
		s = "/static/images/avatars/avatar-cat.webp"
	case "parrot":
		s = "/static/images/avatars/avatar-parrot.webp"
	case "bulldog":
		s = "/static/images/avatars/avatar-bulldog.webp"
	case "default":
		s = "/static/images/avatars/avatar-shiba.webp"
	}
	return s
}
