package main

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
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

type AuthorPosts struct {
	AuthorInfo Author
	List       []Post
}

// https://stackoverflow.com/questions/38616687/which-way-to-name-a-function-in-go-camelcase-or-semi-camelcase
func rangeIn(low, hi int) int {
	seed := rand.NewSource(time.Now().UnixNano())
	tempRand := rand.New(seed)

	return low + tempRand.Intn(hi-low)
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1

	if i <= 0 {
		return p[1:], "/"
	}

	return p[1:i], p[i:]
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

func getPostById(postId string) (Post, error) {
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

func getAuthorById(authorId string) (AuthorPosts, error) {
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

func getPasswordHash(username string) (string, error) {
	row := db.QueryRow("SELECT password FROM authors WHERE username=$1;", username)

	var password string

	err := row.Scan(&password)

	if err != nil {
		return "", err
	}

	return password, nil
}

func createAuthor(un, ps string) (string, error) {
	hash, err := HashPassword(ps)

	if err != nil {
		return "", err
	}

	for {
		sqlStatement := `
		INSERT INTO authors (author_id, username, password)
		VALUES ($1, $2, $3)
		RETURNING author_id`

		var id string

		err = db.QueryRow(sqlStatement, "100000"+strconv.Itoa(rangeIn(100000000, 999999999)), un, hash).Scan(&id)

		if err == nil {
			return id, nil
		}
	}
}

type App struct {
	AuthorHandler *AuthorHandler
	AuthHandler   *AuthHandler
	PostHandler   *PostHandler
	RootHandler   *RootHandler
}

func (h *App) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string

	head, req.URL.Path = ShiftPath(req.URL.Path)

	switch head {
	case "":
		h.RootHandler.ServeHTTP(res, req)
	case "auth":
		h.AuthHandler.ServeHTTP(res, req)
	case "author":
		h.AuthorHandler.ServeHTTP(res, req)
	case "post":
		h.PostHandler.ServeHTTP(res, req)
	default:
		http.Error(res, "Not Found", http.StatusNotFound)
	}

	return
}

type AuthorHandler struct {
}

func (h *AuthorHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var authorId string
	authorId, req.URL.Path = ShiftPath(req.URL.Path)

	content, err := getAuthorById(authorId)

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

	content, err := getPostById(postId)

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

type AuthHandler struct {
	SignupHandler *SignupHandler
	SigninHandler *SigninHandler
}

func (h *AuthHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = ShiftPath(req.URL.Path)

	switch head {
	case "signup":
		h.SignupHandler.ServeHTTP(res, req)
	case "signin":
		h.SigninHandler.ServeHTTP(res, req)
	default:
		http.Error(res, "Not Found", http.StatusNotFound)
	}

	return
}

type SignupHandler struct {
	SignupStartHandler *SignupStartHandler
	SignupEndHandler   *SignupEndHandler
}

func (h *SignupHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = ShiftPath(req.URL.Path)

	switch head {
	case "":
		h.SignupStartHandler.ServeHTTP(res, req)
	case "finish":
		h.SignupEndHandler.ServeHTTP(res, req)
	default:
		http.Error(res, "Not Found", http.StatusNotFound)
	}

	return
}

type SignupStartHandler struct {
}

func (h *SignupStartHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	renderTemplate(res, "signup", nil)
}

type SignupEndHandler struct {
}

func (h *SignupEndHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = ShiftPath(req.URL.Path)

	if head != "" {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	if req.Method == "POST" {
		un := req.FormValue("username")
		ps := req.FormValue("password")

		authorId, err := createAuthor(un, ps)

		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(res, req, "/author/"+authorId, http.StatusFound)
	} else {
		http.Error(res, "Only POST is allowed", http.StatusMethodNotAllowed)
	}
}

type SigninHandler struct {
	SigninStartHandler *SigninStartHandler
	SigninEndHandler   *SigninEndHandler
}

func (h *SigninHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = ShiftPath(req.URL.Path)

	switch head {
	case "":
		h.SigninStartHandler.ServeHTTP(res, req)
	case "finish":
		h.SigninEndHandler.ServeHTTP(res, req)
	default:
		http.Error(res, "Not Found", http.StatusNotFound)
	}

	return
}

type SigninStartHandler struct {
}

func (h *SigninStartHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	renderTemplate(res, "signin", nil)
}

type SigninEndHandler struct {
}

func (h *SigninEndHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = ShiftPath(req.URL.Path)

	if head != "" {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	if req.Method == "POST" {
		un := req.FormValue("username")
		ps := req.FormValue("password")

		encryptedPassword, err := getPasswordHash(un)

		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		psMatched := CheckPasswordHash(ps, encryptedPassword)
		
		if psMatched {
			//http.Redirect(res, req, "/author/"+authorId, http.StatusFound)
			http.Error(res, "Valid credentials", http.StatusFound)
		} else {
			http.Error(res, "Invalid credentials", http.StatusInternalServerError)
		}
	} else {
		http.Error(res, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}
}

type RootHandler struct {
}

func (h *RootHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = ShiftPath(req.URL.Path)

	if head == "" {
		content, _ := getAllPosts()

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

	app := &App{
		RootHandler:   new(RootHandler),
		PostHandler:   new(PostHandler),
		AuthorHandler: new(AuthorHandler),
		AuthHandler:   &AuthHandler{SignupHandler: new(SignupHandler), SigninHandler: new(SigninHandler)},
	}

	http.ListenAndServe(":8080", app)
}

