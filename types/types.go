package types

import "time"

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

type DefaultResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ValidResponse struct {
	Status  string      `json:"status"`
	Content interface{} `json:"content"`
}
