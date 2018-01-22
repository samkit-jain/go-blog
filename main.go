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

type Author struct {
	Username string
	AuthorId string
}

type Post struct {
	Id         string
	Title      string
	Body       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	AuthorInfo Author
}

type Posts struct {
	IsEmpty bool
	List    []Post
}

// get information of all posts
func (dbc *DBConnection) getAllPosts() ([]Post, error) {
	result := make([]Post, 0)
	rows, err := dbc.db.Query("SELECT post_id, title, SUBSTRING(body FOR 100) AS body FROM posts ORDER BY updated_at DESC;")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			postId string
			title  string
			body   string
		)

		err = rows.Scan(&postId, &title, &body)

		if err != nil {
			return nil, err
		}

		result = append(result, Post{Id: postId, Title: title, Body: body})
	}

	// get any error encountered during iteration
	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return result, nil
}

// get information for a specific post
func (dbc *DBConnection) getPost(postId string) (Post, error) {
	result := Post{}
	row := dbc.db.QueryRow("SELECT authors.username, authors.author_id, posts.title, posts.body, posts.created_at, posts.updated_at FROM authors, posts WHERE posts.post_id=$1;", postId)

	var (
		authorName    string
		authorId      string
		postTitle     string
		postBody      string
		postCreatedAt time.Time
		postUpdatedAt time.Time
	)

	err := row.Scan(&authorName, &authorId, &postTitle, &postBody, &postCreatedAt, &postUpdatedAt)

	if err != nil {
		return Post{}, err
	}

	result = Post{Id: postId, Title: postTitle, Body: postBody, CreatedAt: postCreatedAt, UpdatedAt: postUpdatedAt, AuthorInfo: Author{Username: authorName, AuthorId: authorId}}

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

func (dbc *DBConnection) postHandler(w http.ResponseWriter, r *http.Request) {
	postId := r.URL.Path[len("/post/"):]
	content, err := dbc.getPost(postId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplateSingle(w, "post", &content)
}

var templates = template.Must(template.ParseGlob("templates/blog/*"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Posts) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderTemplateSingle(w http.ResponseWriter, tmpl string, p *Post) {
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
	http.HandleFunc("/post/", dbconn.postHandler)
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
