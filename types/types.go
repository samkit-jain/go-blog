// Package types provides objects used multiple times in the project
package types

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Object containing author's properties
type Author struct {
	Username  string    `json:"username"`   // author's username
	AuthorId  string    `json:"id"`         // author's ID
	CreatedAt time.Time `json:"created_at"` // author's creation date
}

// Object containing post's properties
type Post struct {
	Id         string    `json:"id"`         // post's ID
	Title      string    `json:"title"`      // post's title
	Body       string    `json:"body"`       // post's body
	CreatedAt  time.Time `json:"created_at"` // post's creation date
	UpdatedAt  time.Time `json:"updated_at"` // post's modification date
	AuthorInfo Author    `json:"author"`     // post's author
}

// Object containing author's properties including all his/her posts
type AuthorPosts struct {
	AuthorInfo Author `json:"author"` // author
	List       []Post `json:"posts"`  // author's posts
}

// Default JSON response object
type DefaultResponse struct {
	Status  string `json:"status"`  // status field
	Message string `json:"message"` // message field
}

// Not default JSON response object
type ValidResponse struct {
	Status  string      `json:"status"`  // status field
	Content interface{} `json:"content"` // content field
}

// Custom claims for the JSON web token
type CustomClaims struct {
	Id                 string `json:"id"` // author's ID
	jwt.StandardClaims        // standard JWT claims (issuer, expiry, etc.)
}
