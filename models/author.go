package models

import (
	"strconv"
	"time"

	_ "github.com/lib/pq"

	"github.com/samkit-jain/go-blog/config"
	"github.com/samkit-jain/go-blog/helpers"
	"github.com/samkit-jain/go-blog/types"
)

func GetAllAuthors() ([]types.Author, error) {
	result := make([]types.Author, 0)
	rows, err := config.DB.Query("SELECT author_id, username, created_at FROM authors ORDER BY username;")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			authorId string
			username string
			createdAt time.Time
		)

		err = rows.Scan(&authorId, &username, &createdAt)

		if err != nil {
			return nil, err
		}

		result = append(result, types.Author{AuthorId: authorId, Username: username, CreatedAt: createdAt})
	}

	// get any error encountered during iteration
	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetAuthorById(authorId string) (types.AuthorPosts, error) {
	// Getting Author's info
	row := config.DB.QueryRow("SELECT username, created_at FROM authors WHERE author_id=$1;", authorId)

	var (
		authorName      string
		authorCreatedAt time.Time
	)

	err := row.Scan(&authorName, &authorCreatedAt)

	if err != nil {
		return types.AuthorPosts{}, err
	}

	posts := make([]types.Post, 0)
	result := types.AuthorPosts{AuthorInfo: types.Author{Username: authorName, AuthorId: authorId, CreatedAt: authorCreatedAt}, List: posts}

	// Getting Author's posts
	rows, err := config.DB.Query("SELECT post_id, title, SUBSTRING(body FOR 100) AS body FROM posts WHERE author_id=$1 ORDER BY created_at DESC;", authorId)

	if err != nil {
		return result, err
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
			return result, err
		}

		result.List = append(result.List, types.Post{Id: postId, Title: title, Body: body})
	}

	// get any error encountered during iteration
	err = rows.Err()

	if err != nil {
		return result, err
	}

	return result, nil
}

func GetPasswordHash(username string) (string, error) {
	row := config.DB.QueryRow("SELECT password FROM authors WHERE username=$1;", username)

	var password string

	err := row.Scan(&password)

	if err != nil {
		return "", err
	}

	return password, nil
}

func CreateAuthor(un, ps string) (string, error) {
	hash, err := helpers.HashPassword(ps)

	if err != nil {
		return "", err
	}

	for {
		sqlStatement := `
		INSERT INTO authors (author_id, username, password)
		VALUES ($1, $2, $3)
		RETURNING author_id`

		var id string

		err = config.DB.QueryRow(sqlStatement, "100000"+strconv.Itoa(helpers.RangeIn(100000000, 999999999)), un, hash).Scan(&id)

		if err == nil {
			return id, nil
		}
	}
}
