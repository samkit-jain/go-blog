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
	Username  string
	AuthorId  string
	CreatedAt time.Time
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

type AuthorPosts struct {
	AuthorInfo Author
	List       []Post
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

	result := Post{Id: postId, Title: postTitle, Body: postBody, CreatedAt: postCreatedAt, UpdatedAt: postUpdatedAt, AuthorInfo: Author{Username: authorName, AuthorId: authorId}}

	return result, nil
}

// get information for a specific author
func (dbc *DBConnection) getAuthor(authorId string) (AuthorPosts, error) {
	// Getting Author's info
	row := dbc.db.QueryRow("SELECT username, created_at FROM authors WHERE author_id=$1;", authorId)

	var (
		authorName      string
		authorCreatedAt time.Time
	)

	err := row.Scan(&authorName, &authorCreatedAt)

	if err != nil {
		return AuthorPosts{}, err
	}

	posts := make([]Post, 0)
	result := AuthorPosts{AuthorInfo: Author{Username: authorName, AuthorId: authorId, CreatedAt: authorCreatedAt}, List: posts}

	// Getting Author's posts
	rows, err := dbc.db.Query("SELECT post_id, title, SUBSTRING(body FOR 100) AS body FROM posts WHERE author_id=$1 ORDER BY created_at DESC;", authorId)

	if err != nil {
		return result, err
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
			return result, err
		}

		result.List = append(result.List, Post{Id: postId, Title: title, Body: body})
	}

	// get any error encountered during iteration
	err = rows.Err()

	if err != nil {
		return result, err
	}

	return result, nil
}

// ........................................................................
// .. Use https://stackoverflow.com/a/35967196/7760998 to remove isEmpty ..
// ........................................................................
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

func (dbc *DBConnection) authorHandler(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Path[len("/author/"):]
	content, err := dbc.getAuthor(authorId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplateAuthor(w, "author", &content)
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

func renderTemplateAuthor(w http.ResponseWriter, tmpl string, ap *AuthorPosts) {
	err := templates.ExecuteTemplate(w, tmpl+".html", ap)

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
	http.HandleFunc("/author/", dbconn.authorHandler)
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
