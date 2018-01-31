package models

import (
	"time"

	_ "github.com/lib/pq"

	"github.com/samkit-jain/go-blog/config"
	"github.com/samkit-jain/go-blog/helpers"
	"github.com/samkit-jain/go-blog/types"
	"strconv"
)

func GetAllPosts() ([]types.Post, error) {
	result := make([]types.Post, 0)
	rows, err := config.DB.Query("SELECT post_id, title, body, created_at, updated_at FROM posts ORDER BY updated_at DESC;")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			postId    string
			title     string
			body      string
			createdAt time.Time
			updatedAt time.Time
		)

		err = rows.Scan(&postId, &title, &body, &createdAt, &updatedAt)

		if err != nil {
			return nil, err
		}

		result = append(result, types.Post{Id: postId, Title: title, Body: body, CreatedAt: createdAt, UpdatedAt: updatedAt})
	}

	// get any error encountered during iteration
	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetPostById(postId string) (types.Post, error) {
	row := config.DB.QueryRow("SELECT authors.username, authors.author_id, authors.created_at, posts.title, posts.body, posts.created_at, posts.updated_at FROM authors JOIN posts ON(authors.author_id=posts.author_id) WHERE posts.post_id=$1;", postId)

	var (
		authorName      string
		authorId        string
		authorCreatedAt time.Time
		postTitle       string
		postBody        string
		postCreatedAt   time.Time
		postUpdatedAt   time.Time
	)

	err := row.Scan(&authorName, &authorId, &authorCreatedAt, &postTitle, &postBody, &postCreatedAt, &postUpdatedAt)

	if err != nil {
		return types.Post{}, err
	}

	result := types.Post{Id: postId, Title: postTitle, Body: postBody, CreatedAt: postCreatedAt, UpdatedAt: postUpdatedAt, AuthorInfo: types.Author{Username: authorName, AuthorId: authorId, CreatedAt: authorCreatedAt}}

	return result, nil
}

func CreatePost(title, body, author string) (string, error) {
	for {
		sqlStatement := `
		INSERT INTO posts (post_id, title, body, author_id)
		VALUES ($1, $2, $3, $4)
		RETURNING post_id;`

		var id string

		err := config.DB.QueryRow(sqlStatement, "500000"+strconv.Itoa(helpers.RangeIn(100000000, 999999999)), title, body, author).Scan(&id)

		if err == nil {
			return id, nil
		}
	}
}

func UpdatePost(postId, title, body, author string) (string, error) {
	sqlStatement := "UPDATE posts SET title=$1, body=$2 WHERE author_id=$3 AND post_id=$4 RETURNING post_id;"

	var id string

	err := config.DB.QueryRow(sqlStatement, title, body, author, postId).Scan(&id)

	if err == nil {
		return id, nil
	}

	return "", err
}
