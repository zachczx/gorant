package users

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
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

var avatars = []string{"bird", "bird2", "bulldog", "cat", "cat2", "cat3", "cat4", "cat5", "cat6", "corgi", "dog2", "dog3", "dog4", "dog5", "goldfish", "hamster", "husky", "owl", "pigeon", "poodle", "rabbit", "shiba", "snake", "turtle", "whitehamster"}

func (u *User) GetSettings(username string) error {
	if err := database.DB.QueryRow("SELECT * FROM users WHERE user_id=$1", username).Scan(&u.UserID, &u.Email, &u.PreferredName, &u.ContactMe, &u.Avatar, &u.SortComments); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no user settings found: %w", err)
		}
		return fmt.Errorf("error fetching settings getsettings() from db: %w", err)
	}
	u.ContactMeString = strconv.Itoa(u.ContactMe)
	u.AvatarPath = ChooseAvatar(u.Avatar)
	return nil
}

const regex string = `^[0-9A-Za-z -_+()[]|@\.]+$`

func Validate(s Settings) map[string](string) {
	v := govalidator.New()
	ok := false
	for _, v := range avatars {
		if v == s.Avatar {
			ok = true
			break
		}
	}
	v.RequiredString(s.PreferredName, "preferred_name", "Please enter a preferred name").RegexMatches(s.PreferredName, regex, "preferred_name", "No special characters allowed! (Use only A-Z, a-z, 0-9, -, _, brackets, +)").MaxString(s.PreferredName, 255, "preferred_name", "Message is more than 255 characters.")
	v.CustomRule(ok, "avatar", "Unrecognized avatar")
	if v.IsFailed() {
		return v.Errors()
	}
	return nil
}

func SaveSettings(username string, s Settings) error {
	// ContactMe is the opposite of the form value:
	// - Checking box (on) = Don't contact me = 0
	// - Not checking box ("") = Contact me = 1 = Default
	if s.ContactMe == "on" {
		s.ContactMe = "0"
	} else {
		s.ContactMe = "1"
	}

	_, err := database.DB.Exec("UPDATE users SET preferred_name=$1, contact_me=$2, avatar=$3, sort_comments=$4 WHERE user_id=$5;", s.PreferredName, s.ContactMe, s.Avatar, s.SortComments, username)
	if err != nil {
		return fmt.Errorf("error updating users table to save settings: %w", err)
	}

	return nil
}

func SaveSortComments(username string, s string) (string, error) {
	switch s {
	case "upvote;desc", "upvote;asc", "date;desc", "date;asc":
		_, err := database.DB.Exec("UPDATE users SET sort_comments=$1 WHERE user_id=$2;", s, username)
		if err != nil {
			return s, fmt.Errorf("error updating users table to save sort comments: %w", err)
		}
	default:
		return s, fmt.Errorf("error: unknown value")
	}
	return s, nil
}

// TODO avatar choice.
func ReturnAvatars() []string {
	return avatars
}

func ChooseAvatar(c string) string {
	s := "/static/images/avatars/%s.svg"
	for _, v := range avatars {
		if c == v {
			s = fmt.Sprintf(s, v)
			return s
		}
	}
	// Default value if it's default value
	s = fmt.Sprintf(s, "shiba")
	return s
}

var emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

func CheckUsername(username string) (exists bool, err error) {
	regex, err := regexp.Compile(emailRegex)
	if err != nil {
		fmt.Println(err)
	}
	if !regex.MatchString(username) {
		err = errors.New("not an email string received")
		return exists, err
	}
	var dbUserID string
	err = database.DB.QueryRow(`SELECT user_id FROM users WHERE user_id=$1`, username).Scan(&dbUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		fmt.Println(err)
	}
	return true, nil
}
