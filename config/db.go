// Package config initialises a postgresql database instance
package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// database credentials
const (
	// host to connect to
	host = "localhost"

	// host to connect to
	port = 5432

	// user to sign in as
	user = "samkit"

	// name of the database to connect to
	dbname = "goblog"
)

// database connection instance
var DB *sql.DB

func InitDB() {
	// creating the database connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, os.Getenv("GOBLOG_PS"), dbname)

	var err error

	// since DB is global, using := creates a local DB
	// validating information provided
	DB, err = sql.Open("postgres", psqlInfo)

	// information provided incorrect
	if err != nil {
		panic(err)
	}

	// defer DB.Close()

	// can't connect to DB
	if err = DB.Ping(); err != nil {
		panic(err)
	}
}
