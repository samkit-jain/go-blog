package website

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/samkit-jain/go-blog/helpers"
	"github.com/samkit-jain/go-blog/models"
)

type WebsiteHandler struct {
	AuthorHandler *AuthorHandler
	AuthHandler   *AuthHandler
	PostHandler   *PostHandler
	RootHandler   *RootHandler
}

func NewWebsiteHandler() *WebsiteHandler {
	return &WebsiteHandler{
		AuthorHandler: new(AuthorHandler),
		PostHandler:   new(PostHandler),
		RootHandler:   new(RootHandler),
		AuthHandler: &AuthHandler{
			SignupHandler: new(SignupHandler),
			SigninHandler: new(SigninHandler),
		},
	}
}

func (h *WebsiteHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string

	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

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
	authorId, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	if req.URL.Path != "/" {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	content, err := models.GetAuthorById(authorId)

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
	postId, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	if req.URL.Path != "/" {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	content, err := models.GetPostById(postId)

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
	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

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
	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

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
	if req.URL.Path != "/" {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	renderTemplate(res, "signup", nil)
}

type SignupEndHandler struct {
}

func (h *SignupEndHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	if req.URL.Path != "/" {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	if head != "" {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	if req.Method == "POST" {
		un := req.FormValue("username")
		ps := req.FormValue("password")

		authorId, err := models.CreateAuthor(un, ps)

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
	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

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
	if req.URL.Path != "/" {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	renderTemplate(res, "signin", nil)
}

type SigninEndHandler struct {
}

func (h *SigninEndHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	if req.URL.Path != "/" {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	if head != "" {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	if req.Method == "POST" {
		un := req.FormValue("username")
		ps := req.FormValue("password")

		encryptedPassword, err := models.GetPasswordHash(un)

		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		psMatched := helpers.CheckPasswordHash(ps, encryptedPassword)

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
	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	if req.URL.Path != "/" {
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}

	if head == "" {
		content, _ := models.GetAllPosts()

		renderTemplate(res, "home", content)
		return
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
