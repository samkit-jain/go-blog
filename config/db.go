package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "samkit"
	dbname = "goblog"
)

var DB *sql.DB

func InitDB() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, os.Getenv("GOBLOG_PS"), dbname)

	// since DB is global, using := creates a local DB
	var err error

	// open connection to DB
	DB, err = sql.Open("postgres", psqlInfo)

	// opening connection failed
	if err != nil {
		panic(err)
	}

	// defer DB.Close()

	// can't connect to DB
	if err = DB.Ping(); err != nil {
		panic(err)
	}
}
