package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	file string = "./starter.db"
)

func Connect() (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", file)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	fmt.Println("DB connected!")

	return db, nil
}
