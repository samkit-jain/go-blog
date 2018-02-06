// Package models provides helper functions for querying database
package models

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/lib/pq"

	"github.com/samkit-jain/go-blog/config"
	"github.com/samkit-jain/go-blog/helpers"
	"github.com/samkit-jain/go-blog/types"
)

// GetAllPosts returns a list of all the posts
func GetAllPosts() ([]types.Post, error) {
	result := make([]types.Post, 0)
	rows, err := config.DB.Query("SELECT authors.username, authors.author_id, authors.created_at, posts.post_id, posts.title, posts.body, posts.created_at, posts.updated_at FROM authors JOIN posts ON(authors.author_id=posts.author_id) ORDER BY posts.updated_at DESC;")

	// ooh, an error
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// iterate through all the rows
	for rows.Next() {
		var (
			authorName      string
			authorId        string
			authorCreatedAt time.Time
			postId          string
			postTitle       string
			postBody        string
			postCreatedAt   time.Time
			postUpdatedAt   time.Time
		)

		err := rows.Scan(&authorName, &authorId, &authorCreatedAt, &postId, &postTitle, &postBody, &postCreatedAt, &postUpdatedAt)

		if err != nil {
			return nil, err
		}

		result = append(result, types.Post{Id: postId, Title: postTitle, Body: postBody, CreatedAt: postCreatedAt, UpdatedAt: postUpdatedAt, AuthorInfo: types.Author{Username: authorName, AuthorId: authorId, CreatedAt: authorCreatedAt}})
	}

	// get any error encountered during iteration
	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetPostById returns the post specified by postId
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

	// an error including sql.ErrNoRows
	if err != nil {
		return types.Post{}, err
	}

	result := types.Post{Id: postId, Title: postTitle, Body: postBody, CreatedAt: postCreatedAt, UpdatedAt: postUpdatedAt, AuthorInfo: types.Author{Username: authorName, AuthorId: authorId, CreatedAt: authorCreatedAt}}

	return result, nil
}

// CreatePost creates a post for author with title and body
func CreatePost(title, body, author string) (string, error) {
	i := 1

	// loop till unique ID created but not till infinity
	for {
		sqlStatement := `
		INSERT INTO posts (post_id, title, body, author_id)
		VALUES ($1, $2, $3, $4)
		RETURNING post_id;`

		var id string

		err := config.DB.QueryRow(sqlStatement, "500000"+strconv.Itoa(helpers.RangeIn(100000000, 999999999)), title, body, author).Scan(&id)

		if err == nil {
			return id, nil
		} else if i == 100 {
			return "", err
		} else if pgerr, ok := err.(*pq.Error); ok {
			// 23505 -> primary key exists (unique_violation)
			if pgerr.Code != "23505" {
				return "", err
			}
		}

		i += 1
	}
}

// UpdatePost modifies an author's post
func UpdatePost(postId, title, body, author string) (string, error) {
	sqlStatement := "UPDATE posts SET title=$1, body=$2 WHERE author_id=$3 AND post_id=$4 RETURNING post_id;"

	var id string

	err := config.DB.QueryRow(sqlStatement, title, body, author, postId).Scan(&id)

	if err == nil {
		return id, nil
	}

	return "", err
}

// DeletePost deletes an author's post
func DeletePost(postId, author string) error {
	sqlStatement := "DELETE FROM posts WHERE author_id=$1 AND post_id=$2;"

	err := config.DB.QueryRow(sqlStatement, author, postId).Scan()

	if err == sql.ErrNoRows {
		err = nil
	}

	return err
}

// DeletePost deletes an author's all posts
func DeletePosts(author string) error {
	sqlStatement := "DELETE FROM posts WHERE author_id=$1;"

	err := config.DB.QueryRow(sqlStatement, author).Scan()

	if err == sql.ErrNoRows {
		err = nil
	}

	return err
}
