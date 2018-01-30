package api

import (
	"encoding/json"
	"net/http"

	"fmt"
	"github.com/samkit-jain/go-blog/helpers"
	"github.com/samkit-jain/go-blog/models"
	"github.com/samkit-jain/go-blog/types"
)

type ApiHandler struct {
	AuthorHandler *AuthorHandler
	//AuthHandler   *AuthHandler
	PostHandler *PostHandler
}

func (h *ApiHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string

	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	switch head {
	//case "auth":
	//	h.AuthHandler.ServeHTTP(res, req)
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

	res.Header().Set("Content-Type", "application/json")

	authorId, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	if req.URL.Path != "/" {
		res.WriteHeader(http.StatusNotFound)
		json.NewEncoder(res).Encode(helpers.NotFound())

		return
	}

	if req.Method == "GET" {
		if authorId == "" {
			content, _ := models.GetAllAuthors()

			res.WriteHeader(http.StatusOK)
			json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: content})
		} else {
			content, err := models.GetAuthorById(authorId)

			if err != nil {
				// Change response based on status, InternalServerError, etc.
				res.WriteHeader(http.StatusOK)
				json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: "ID does not exist!"})
			} else {
				res.WriteHeader(http.StatusOK)
				json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: content})
			}
		}
	} else {
		res.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(res).Encode(helpers.InvalidMethod())
	}

	return
}

type PostHandler struct {
}

func (h *PostHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var postId string

	res.Header().Set("Content-Type", "application/json")

	postId, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	if req.URL.Path != "/" {
		res.WriteHeader(http.StatusNotFound)
		json.NewEncoder(res).Encode(helpers.NotFound())

		return
	}

	if req.Method == "GET" {
		if postId == "" {
			content, _ := models.GetAllPosts()

			fmt.Println(content)

			res.WriteHeader(http.StatusOK)
			json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: content})
		} else {
			content, err := models.GetPostById(postId)

			if err != nil {
				// Change response based on status, InternalServerError, etc.
				res.WriteHeader(http.StatusOK)
				json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: "ID does not exist!"})
			} else {
				res.WriteHeader(http.StatusOK)
				json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: content})
			}
		}
	} else {
		res.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(res).Encode(helpers.InvalidMethod())
	}

	return
}

/*
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
	renderTemplate(res, "signup", nil)
}

type SignupEndHandler struct {
}

func (h *SignupEndHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

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
	renderTemplate(res, "signin", nil)
}

type SigninEndHandler struct {
}

func (h *SigninEndHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

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
*/
