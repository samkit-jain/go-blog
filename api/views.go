// Package api provides handlers to route and handle <base>/api/... HTTP calls
package api

import (
	"encoding/json"
	"net/http"

	"database/sql"
	"github.com/samkit-jain/go-blog/helpers"
	"github.com/samkit-jain/go-blog/models"
	"github.com/samkit-jain/go-blog/types"
)

// Base handler for <base>/api/... calls
type ApiHandler struct {
	AuthorHandler *AuthorHandler
	LoginHandler  *LoginHandler
	PostHandler   *PostHandler
}

// ApiHandler's constructor
func NewApiHandler() *ApiHandler {
	return &ApiHandler{
		AuthorHandler: &AuthorHandler{
			AuthorIdPresentHandler:    new(AuthorIdPresentHandler),
			AuthorIdNotPresentHandler: new(AuthorIdNotPresentHandler),
		},
		PostHandler: &PostHandler{
			PostIdPresentHandler:    new(PostIdPresentHandler),
			PostIdNotPresentHandler: new(PostIdNotPresentHandler),
		},
		LoginHandler: new(LoginHandler),
	}
}

// ApiHandler's ServeHTTP receives <base>/api/:profile/... calls and based on the profile routes to
// best matched handler
func (h *ApiHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string

	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	switch head {
	case "authors": // <base>/api/authors/...
		h.AuthorHandler.ServeHTTP(res, req)
	case "posts": // <base>/api/posts/...
		h.PostHandler.ServeHTTP(res, req)
	case "login": // <base>/api/login/...
		h.LoginHandler.ServeHTTP(res, req)
	default: // all other
		helpers.NotFoundResponse(res)
	}

	return
}

// Handler for <base>/api/authors/... calls
type AuthorHandler struct {
	AuthorIdPresentHandler    *AuthorIdPresentHandler
	AuthorIdNotPresentHandler *AuthorIdNotPresentHandler
}

// AuthorHandler's ServeHTTP serves URLs of authors profile
func (h *AuthorHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// author's ID
	var authorId string

	// Only GET method allowed
	if req.Method != "GET" {
		helpers.MethodNotAllowedResponse(res)
		return
	}

	// get authorId from path
	authorId, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	// URL not empty even after removing authorId
	if req.URL.Path != "/" {
		helpers.NotFoundResponse(res)
		return
	}

	switch authorId {
	case "":
		// path /authors/
		h.AuthorIdNotPresentHandler.ServeHTTP(res, req)
	default:
		// path /authors/:authorId
		h.AuthorIdPresentHandler.Handler(authorId).ServeHTTP(res, req)
	}

	return
}

// AuthorIdPresentHandler handles author URLs with authorId
type AuthorIdPresentHandler struct {
}

// AuthorIdPresentHandler's ServeHTTP returns information of author (including posts) whose id is authorId
//
// GET	<base>/api/authors/:authorId
func (h *AuthorIdPresentHandler) Handler(authorId string) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// set response header's content-type
		res.Header().Set("Content-Type", "application/json")

		if content, err := models.GetAuthorById(authorId); err != nil {
			if err == sql.ErrNoRows {
				helpers.BadRequestResponse(res, "Author does not exist!")
			} else {
				helpers.InternalServerErrorResponse(res, err.Error())
			}
		} else {
			res.WriteHeader(http.StatusOK)
			json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: helpers.CreatePostWithoutAuthor(content)})
		}

		return
	})
}

// AuthorIdNotPresentHandler handles author URLs without authorId
type AuthorIdNotPresentHandler struct {
}

// AuthorIdNotPresentHandler's ServeHTTP returns information of all authors (excluding posts)
//
// GET	<base>/api/authors/
func (h *AuthorIdNotPresentHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	// set response header's content-type
	res.Header().Set("Content-Type", "application/json")

	if content, err := models.GetAllAuthors(); err != nil {
		if err == sql.ErrNoRows {
			res.WriteHeader(http.StatusOK)
			json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: content})
		} else {
			helpers.InternalServerErrorResponse(res, err.Error())
		}
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: content})
	}

	return
}

// Handler for <base>/api/posts/... calls
type PostHandler struct {
	PostIdPresentHandler    *PostIdPresentHandler
	PostIdNotPresentHandler *PostIdNotPresentHandler
}

// PostHandler's ServeHTTP serves URLs of posts profile
func (h *PostHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var postId string

	res.Header().Set("Content-Type", "application/json")

	postId, req.URL.Path = helpers.ShiftPath(req.URL.Path)
	authorId := helpers.GetAuthorIdFromHeader(req)

	// URL not empty even after removing postId
	if req.URL.Path != "/" {
		helpers.NotFoundResponse(res)
		return
	}

	switch postId {
	case "":
		// path /posts/
		h.PostIdNotPresentHandler.Handler(authorId).ServeHTTP(res, req)
	default:
		// path /posts/:postId
		h.PostIdPresentHandler.Handler(postId, authorId).ServeHTTP(res, req)
	}

	return
}

// PostIdPresentHandler handles post URLs with postId
type PostIdPresentHandler struct {
}

// PostIdPresentHandler's method to handle URLs of type
//
// GET  	<base>/api/posts/:postId	Info of a specific post
//
// PUT  	<base>/api/posts/:postId	Update a specific post
//
// DELETE  	<base>/api/posts/:postId	Delete a specific post
func (h *PostIdPresentHandler) Handler(postId, authorId string) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// set response header's content-type
		res.Header().Set("Content-Type", "application/json")

		switch req.Method {
		case "GET":
			if content, err := models.GetPostById(postId); err != nil {
				if err == sql.ErrNoRows {
					helpers.BadRequestResponse(res, "Post does not exist!")
				} else {
					helpers.InternalServerErrorResponse(res, err.Error())
				}
			} else {
				res.WriteHeader(http.StatusOK)
				json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: content})
			}
		case "PUT":
			if authorId != "" {
				title := req.FormValue("title")
				body := req.FormValue("body")

				if postId, err := models.UpdatePost(postId, title, body, authorId); err != nil {
					if err == sql.ErrNoRows {
						helpers.BadRequestResponse(res, "You don't have write access to the post!")
					} else {
						helpers.InternalServerErrorResponse(res, err.Error())
					}
				} else {
					res.WriteHeader(http.StatusOK)
					json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: postId})
				}
			} else {
				helpers.ForbiddenResponse(res)
			}
		case "DELETE":
			if authorId != "" {
				if err := models.DeletePost(postId, authorId); err != nil {
					if err == sql.ErrNoRows {
						helpers.BadRequestResponse(res, "You don't have write access to the post!")
					} else {
						helpers.InternalServerErrorResponse(res, err.Error())
					}
				} else {
					res.WriteHeader(http.StatusOK)
					json.NewEncoder(res).Encode(types.DefaultResponse{Status: "success", Message: "Post deleted!"})
				}
			} else {
				helpers.ForbiddenResponse(res)
				return
			}
		default:
			helpers.MethodNotAllowedResponse(res)
		}

		return
	})
}

// PostIdNotPresentHandler handles post URLs without postId
type PostIdNotPresentHandler struct {
}

// PostIdPresentHandler's method to handle URLs of type
//
// GET		/posts/			Info of all posts
//
// POST 	/posts/			Create a new post
//
// DELETE	/posts/			Delete all posts of a specific author
func (h *PostIdNotPresentHandler) Handler(authorId string) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// set response header's content-type
		res.Header().Set("Content-Type", "application/json")

		switch req.Method {
		case "GET":
			if content, err := models.GetAllPosts(); err != nil {
				if err == sql.ErrNoRows {
					res.WriteHeader(http.StatusOK)
					json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: content})
				} else {
					helpers.InternalServerErrorResponse(res, err.Error())
				}
			} else {
				res.WriteHeader(http.StatusOK)
				json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: content})
			}
		case "POST":
			if authorId != "" {
				// read passed form parameters
				title := req.FormValue("title")
				body := req.FormValue("body")

				if postId, err := models.CreatePost(title, body, authorId); err != nil {
					if err == sql.ErrNoRows {
						helpers.BadRequestResponse(res, "You don't have write access to the post!")
					} else {
						helpers.InternalServerErrorResponse(res, err.Error())
					}
				} else {
					res.WriteHeader(http.StatusOK)
					json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: postId})
				}
			} else {
				helpers.ForbiddenResponse(res)
			}
		case "DELETE":
			if authorId != "" {
				if err := models.DeletePosts(authorId); err != nil {
					if err == sql.ErrNoRows {
						helpers.BadRequestResponse(res, "You don't have write access to the post!")
					} else {
						helpers.InternalServerErrorResponse(res, err.Error())
					}
				} else {
					res.WriteHeader(http.StatusOK)
					json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: "Deleted!"})
				}
			} else {
				helpers.ForbiddenResponse(res)
			}
		default:
			helpers.MethodNotAllowedResponse(res)
		}

		return
	})
}

// Handler for <base>/api/login/ call
type LoginHandler struct {
}

// LoginHandler's ServeHTTP logs in a user and returns a session token
//
// POST	<base>/api/login/
func (h *LoginHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	_, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	if req.URL.Path != "/" {
		helpers.NotFoundResponse(res)
		return
	}

	if req.Method == "POST" {
		un := req.FormValue("username")
		ps := req.FormValue("password")

		encryptedPassword, err := models.GetPasswordHash(un)

		if err != nil {
			helpers.InternalServerErrorResponse(res, err.Error())
			return
		}

		psMatched := helpers.CheckPasswordHash(ps, encryptedPassword)

		if psMatched {
			authorId, err := models.GetAuthorIdByUsername(un)

			if err != nil {
				helpers.InternalServerErrorResponse(res, err.Error())
				return
			}

			tokenString, err := helpers.CreateToken(authorId)

			if err != nil {
				helpers.InternalServerErrorResponse(res, err.Error())
				return
			}

			res.WriteHeader(http.StatusOK)
			json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: tokenString})
		} else {
			res.WriteHeader(http.StatusOK)
			json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: "Invalid credentials"})
		}
	} else {
		helpers.MethodNotAllowedResponse(res)
	}

	return
}
