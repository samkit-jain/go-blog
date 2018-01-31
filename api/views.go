package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/samkit-jain/go-blog/helpers"
	"github.com/samkit-jain/go-blog/models"
	"github.com/samkit-jain/go-blog/types"
)

type ApiHandler struct {
	AuthorHandler *AuthorHandler
	LoginHandler  *LoginHandler
	PostHandler   *PostHandler
}

func (h *ApiHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string

	head, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	switch head {
	case "authors":
		h.AuthorHandler.ServeHTTP(res, req)
	case "posts":
		h.PostHandler.ServeHTTP(res, req)
	case "login":
		h.LoginHandler.ServeHTTP(res, req)
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

	switch req.Method {
	case "GET":
		if postId == "" {
			content, _ := models.GetAllPosts()

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
	case "POST":
		if postId == "" {
			tokenString := req.Header.Get("token")

			token, _ := jwt.ParseWithClaims(tokenString, &types.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("GOBLOG_SIGNING_KEY")), nil
			})

			if claims, ok := token.Claims.(*types.CustomClaims); ok && token.Valid {
				title := req.FormValue("title")
				body := req.FormValue("body")

				postId, err := models.CreatePost(title, body, claims.Id)

				if err != nil {
					res.WriteHeader(http.StatusOK)
					json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: err.Error()})

					return
				}

				res.WriteHeader(http.StatusOK)
				json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: postId})
			} else {
				res.WriteHeader(http.StatusForbidden)
				json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: "Unauthorized!"})
			}
		} else {
			// update post
		}
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(res).Encode(helpers.InvalidMethod())
	}

	return
}

type LoginHandler struct {
}

func (h *LoginHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")

	_, req.URL.Path = helpers.ShiftPath(req.URL.Path)

	if req.URL.Path != "/" {
		res.WriteHeader(http.StatusNotFound)
		json.NewEncoder(res).Encode(helpers.NotFound())

		return
	}

	if req.Method == "POST" {
		un := req.FormValue("username")
		ps := req.FormValue("password")

		encryptedPassword, err := models.GetPasswordHash(un)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: err.Error()})

			return
		}

		psMatched := helpers.CheckPasswordHash(ps, encryptedPassword)

		if psMatched {
			authorId, err := models.GetAuthorIdByUsername(un)

			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: err.Error()})

				return
			}

			claims := types.CustomClaims{
				Id: authorId,
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
					Issuer:    "goblog",
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(os.Getenv("GOBLOG_SIGNING_KEY")))

			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: err.Error()})

				return
			}

			res.WriteHeader(http.StatusOK)
			json.NewEncoder(res).Encode(types.ValidResponse{Status: "success", Content: tokenString})
		} else {
			res.WriteHeader(http.StatusOK)
			json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: "Invalid credentials"})

			return
		}
	} else {
		res.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(res).Encode(helpers.InvalidMethod())
	}

	return
}
