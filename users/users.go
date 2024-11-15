package users

import (
	"database/sql"
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
}

type Settings struct {
	PreferredName string
	ContactMe     string
}

func GetSettings(username string) (User, error) {
	db, err := database.Connect()
	if err != nil {
		fmt.Println("Error connecting to DB", err)
	}

	var s User
	if err := db.QueryRow("SELECT * FROM users WHERE user_id=$1", username).Scan(&s.UserID, &s.Email, &s.PreferredName, &s.ContactMe); err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Weird, no rows found!")
			return s, err
		}
		s.ContactMeString = strconv.Itoa(s.ContactMe)
	}

	return s, nil
}

func Validate(s Settings) map[string](string) {
	v := govalidator.New()

	// v.RequiredString(c.Name, "name", "Please enter a name")
	v.RequiredString(s.PreferredName, "preferred_name", "Please enter a preferred name").MaxString(s.PreferredName, 255, "preferred_name", "Message is more than 255 characters.")

	if v.IsFailed() {
		return v.Errors()
	}

	return nil
}

func SaveSettings(username string, s Settings) error {
	db, err := database.Connect()
	if err != nil {
		fmt.Println(err)
	}

	// ContactMe is the opposite of the form value:
	// - Checking box (on) = Don't contact me = 0
	// - Not checking box ("") = Contact me = 1 = Default
	if s.ContactMe == "on" {
		s.ContactMe = "0"
	} else {
		s.ContactMe = "1"
	}

	_, err = db.Exec("UPDATE users SET preferred_name=$1, contact_me=$2 WHERE user_id=$3;", s.PreferredName, s.ContactMe, username)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}
