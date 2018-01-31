package types

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Author struct {
	Username  string    `json:"username"`
	AuthorId  string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type Post struct {
	Id         string    `json:"id"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	AuthorInfo Author    `json:"author"`
}

type AuthorPosts struct {
	AuthorInfo Author `json:"author"`
	List       []Post `json:"posts"`
}

type DefaultResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ValidResponse struct {
	Status  string      `json:"status"`
	Content interface{} `json:"content"`
}

type CustomClaims struct {
	Id string `json:"id"`
	jwt.StandardClaims
}
