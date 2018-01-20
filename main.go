package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"net/http"
	//"regexp"
	"html/template"
	"time"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "samkit"
	dbname = "goblog"
)

type DBConnection struct {
	db *sql.DB
}

type Post struct {
	Id        string
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Author    string
}

type Posts struct {
	IsEmpty bool
	List    []Post
}

func (dbc *DBConnection) getAllPosts() ([]Post, error) {
	result := make([]Post, 0)
	rows, err := dbc.db.Query("SELECT post_id, title, body, created_at, updated_at, author_id FROM posts")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			postId    string
			title     string
			body      string
			authorId  string
			createdAt time.Time
			updatedAt time.Time
		)

		err = rows.Scan(&postId, &title, &body, &createdAt, &updatedAt, &authorId)

		if err != nil {
			return nil, err
		}

		result = append(result, Post{Id: postId, Title: title, Body: body, CreatedAt: createdAt, UpdatedAt: updatedAt, Author: authorId})
	}

	// get any error encountered during iteration
	err = rows.Err()

	if err != nil {
		panic(err)
	}

	return result, nil
}

func (dbc *DBConnection) homePageHandler(w http.ResponseWriter, r *http.Request) {
	allPosts, err := dbc.getAllPosts()

	content := &Posts{IsEmpty: false, List: allPosts}

	if err != nil && len(allPosts) > 0 {
		content.IsEmpty = true
	}

	renderTemplate(w, "home", content)
}

var templates = template.Must(template.ParseGlob("templates/blog/*"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Posts) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//var validPath = regexp.MustCompile("^/$")
//
//func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		m := validPath.FindStringSubmatch(r.URL.Path)
//
//		// 404
//		if m == nil {
//			http.NotFound(w, r)
//			return
//		}
//
//		fn(w, r, m[2])
//	}
//}

func main() {
	// string containing info to connect to Postgres
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, os.Getenv("GOBLOG_PS"), dbname)

	// open connection to DB
	db, err := sql.Open("postgres", psqlInfo)

	// opening connection failed
	if err != nil {
		panic(err)
	}

	defer db.Close()

	// connect to DB
	err = db.Ping()

	// can't connect to DB
	if err != nil {
		panic(err)
	}

	dbconn := &DBConnection{db: db}

	http.HandleFunc("/", dbconn.homePageHandler)
	http.ListenAndServe(":8080", nil)

	/*
			sqlStatement := `
		INSERT INTO users (age, email, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

			id := 0
			err = db.QueryRow(sqlStatement, 30, "jon@calhoun.io", "Jonathan", "Calhoun").Scan(&id)
			if err != nil {
				panic(err)
			}
			fmt.Println("New record ID is:", id)
	*/
}
