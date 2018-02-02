package models

import (
	"strconv"
	"time"

	"github.com/lib/pq"

	"github.com/samkit-jain/go-blog/config"
	"github.com/samkit-jain/go-blog/helpers"
	"github.com/samkit-jain/go-blog/types"
)

// GetAllAuthors returns a list of all the authors
func GetAllAuthors() ([]types.Author, error) {
	result := make([]types.Author, 0)
	rows, err := config.DB.Query("SELECT author_id, username, created_at FROM authors ORDER BY username;")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// iterating over result set
	for rows.Next() {
		var (
			authorId  string
			username  string
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

// GetAuthorById searches by authorId and returns information of the author and all the its posts
func GetAuthorById(authorId string) (types.AuthorPosts, error) {
	// getting author's info
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

	// getting author's posts
	rows, err := config.DB.Query("SELECT post_id, title, body, created_at, updated_at FROM posts WHERE author_id=$1 ORDER BY created_at DESC;", authorId)

	if err != nil {
		return result, err
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
			return result, err
		}

		result.List = append(result.List, types.Post{Id: postId, Title: title, Body: body, CreatedAt: createdAt, UpdatedAt: updatedAt, AuthorInfo: types.Author{Username: authorName, AuthorId: authorId, CreatedAt: authorCreatedAt}})
	}

	// get any error encountered during iteration
	err = rows.Err()

	if err != nil {
		return result, err
	}

	return result, nil
}

// GetAuthorById searches by username and returns author's ID
func GetAuthorIdByUsername(username string) (string, error) {
	row := config.DB.QueryRow("SELECT author_id FROM authors WHERE username=$1;", username)

	var authorId string

	err := row.Scan(&authorId)

	if err != nil {
		return "", err
	}

	return authorId, nil
}

// GetPasswordHash returns the encrypted password of an author
func GetPasswordHash(username string) (string, error) {
	row := config.DB.QueryRow("SELECT password FROM authors WHERE username=$1;", username)

	var password string

	err := row.Scan(&password)

	if err != nil {
		return "", err
	}

	return password, nil
}

// CreateAuthor creates an author
func CreateAuthor(un, ps string) (string, error) {
	hash, err := helpers.HashPassword(ps)

	if err != nil {
		return "", err
	}

	i := 1

	// loop till unique ID created but not till infinity
	for {
		sqlStatement := `
		INSERT INTO authors (author_id, username, password)
		VALUES ($1, $2, $3)
		RETURNING author_id`

		var id string

		err = config.DB.QueryRow(sqlStatement, "100000"+strconv.Itoa(helpers.RangeIn(100000000, 999999999)), un, hash).Scan(&id)

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
