package models

import (
	"time"

	_ "github.com/lib/pq"

	"github.com/samkit-jain/go-blog/config"
	"github.com/samkit-jain/go-blog/types"
)

func GetAllPosts() ([]types.Post, error) {
	result := make([]types.Post, 0)
	rows, err := config.DB.Query("SELECT post_id, title, SUBSTRING(body FOR 100) AS body FROM posts ORDER BY updated_at DESC;")

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

		result = append(result, types.Post{Id: postId, Title: title, Body: body})
	}

	// get any error encountered during iteration
	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetPostById(postId string) (types.Post, error) {
	row := config.DB.QueryRow("SELECT authors.username, authors.author_id, posts.title, posts.body, posts.created_at, posts.updated_at FROM authors, posts WHERE posts.post_id=$1;", postId)

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
		return types.Post{}, err
	}

	result := types.Post{Id: postId, Title: postTitle, Body: postBody, CreatedAt: postCreatedAt, UpdatedAt: postUpdatedAt, AuthorInfo: types.Author{Username: authorName, AuthorId: authorId}}

	return result, nil
}
