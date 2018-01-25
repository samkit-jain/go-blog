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
	//"golang.org/x/crypto/bcrypt"
	//"math/rand"
	//"strconv"
	"path"
	"strings"
)

var db *sql.DB

const (
	host   = "localhost"
	port   = 5432
	user   = "samkit"
	dbname = "goblog"
)

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

func getAllPosts() ([]Post, error) {
	result := make([]Post, 0)
	rows, err := db.Query("SELECT post_id, title, SUBSTRING(body FOR 100) AS body FROM posts ORDER BY updated_at DESC;")

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

func getPost(postId string) (Post, error) {
	row := db.QueryRow("SELECT authors.username, authors.author_id, posts.title, posts.body, posts.created_at, posts.updated_at FROM authors, posts WHERE posts.post_id=$1;", postId)

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

func getAuthor(authorId string) (AuthorPosts, error) {
	// Getting Author's info
	row := db.QueryRow("SELECT username, created_at FROM authors WHERE author_id=$1;", authorId)

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
	rows, err := db.Query("SELECT post_id, title, SUBSTRING(body FOR 100) AS body FROM posts WHERE author_id=$1 ORDER BY created_at DESC;", authorId)

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

func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1

	if i <= 0 {
		return p[1:], "/"
	}

	return p[1:i], p[i:]
}

type App struct {
	AuthorHandler *AuthorHandler
	PostHandler *PostHandler
	RootHandler *RootHandler
}

func (h *App) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string

	head, req.URL.Path = ShiftPath(req.URL.Path)

	switch head {
	case "":
		h.RootHandler.ServeHTTP(res, req)
		return
	case "author":
		h.AuthorHandler.ServeHTTP(res, req)
		return
	case "post":
		h.PostHandler.ServeHTTP(res, req)
		return
	}

	http.Error(res, "Not Found", http.StatusNotFound)
}

type AuthorHandler struct {
}

func (h *AuthorHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var authorId string
	authorId, req.URL.Path = ShiftPath(req.URL.Path)

	content, err := getAuthor(authorId)

	if err != nil {
		http.Error(res, fmt.Sprintf("Invalid author ID %q", authorId), http.StatusBadRequest)
		return
		//OR
		//http.Error(res, err.Error(), http.StatusInternalServerError)
		//return
		// depending on the error
	}

	renderTemplate(res, "author", content)
}

type PostHandler struct {
}

func (h *PostHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var postId string
	postId, req.URL.Path = ShiftPath(req.URL.Path)

	content, err := getPost(postId)

	if err != nil {
		http.Error(res, fmt.Sprintf("Invalid post ID %q", postId), http.StatusBadRequest)
		return
		//OR
		//http.Error(res, err.Error(), http.StatusInternalServerError)
		//return
		// depending on the error
	}

	renderTemplate(res, "post", content)
}

type RootHandler struct {
}

func (h *RootHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = ShiftPath(req.URL.Path)

	if head == "" {
		allPosts, err := getAllPosts()

		content := &Posts{IsEmpty: false, List: allPosts}

		if err != nil && len(allPosts) > 0 {
			content.IsEmpty = true
		}

		renderTemplate(res, "home", content)
	}

	http.Error(res, "Not Found", http.StatusNotFound)
}

var templates = template.Must(template.ParseGlob("templates/blog/*"))

func renderTemplate(res http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(res, tmpl+".html", data)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, os.Getenv("GOBLOG_PS"), dbname)

	// since db is global, using := creates a local db
	var err error

	// open connection to DB
	db, err = sql.Open("postgres", psqlInfo)

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

	app := &App {
		RootHandler: new(RootHandler),
		PostHandler: new(PostHandler),
		AuthorHandler: new(AuthorHandler),
	}

	http.ListenAndServe(":8080", app)
}
/*
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

func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

//func CheckPasswordHash(password, hash string) bool {
//	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
//	return err == nil
//}

// create author if not exists
func (dbc *DBConnection) createAuthor(un, ps string) (string, error) {
	hash, err := HashPassword(ps)

	if err != nil {
		return "", err
	}

	// ADD A FOR LOOP THAT RUNS TILL AUTHOR_ID IS UNIQUE
	sqlStatement := `
		INSERT INTO authors (author_id, username, password)
		VALUES ($1, $2, $3)
		RETURNING author_id`

	var id string

	err = dbc.db.QueryRow(sqlStatement, "100000" + strconv.Itoa(rangeIn(100000000, 999999999)), un, hash).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
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

	renderTemplatePosts(w, "home", content)
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

func (dbc *DBConnection) authorRegisterHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "signup")
}

func (dbc *DBConnection) authorRegisterCompleteHandler(w http.ResponseWriter, r *http.Request) {
	un := r.FormValue("username")
	ps := r.FormValue("password")

	authorId, err := dbc.createAuthor(un, ps)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/author/" + authorId, http.StatusFound)
}

func (dbc *DBConnection) authorLoginHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "signin")
}

func (dbc *DBConnection) authorLoginCompleteHandler(w http.ResponseWriter, r *http.Request) {
	un := r.FormValue("username")
	ps := r.FormValue("password")

	authorId, err := dbc.createAuthor(un, ps)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/author/" + authorId, http.StatusFound)
}

var templates = template.Must(template.ParseGlob("templates/blog/*"))

func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderTemplatePosts(w http.ResponseWriter, tmpl string, p *Posts) {
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
	http.HandleFunc("/signup/", dbconn.authorRegisterHandler)
	http.HandleFunc("/register/author/", dbconn.authorRegisterCompleteHandler)
	http.HandleFunc("/signin/", dbconn.authorLoginHandler)
	http.ListenAndServe(":8080", nil)
}
*/
