// Package helpers provides various helper functions
package helpers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"github.com/samkit-jain/go-blog/types"
)

// CheckPasswordHash checks whether hash can be decrypted as password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}

// ShiftPath splits the URL path component into two segments where first segment is the portion
// before first slash and rest is second
//
// Parameter: /this/is/a/url/
//
// Result: this /is/a/url
func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1

	if i <= 0 {
		return p[1:], "/"
	}

	return p[1:i], p[i:]
}

// HashPassword encrypts a string
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// GetAuthorIdFromHeader returns the value of key "token" in the request's header
func GetAuthorIdFromHeader(req *http.Request) string {
	// get session token from HEADER
	tokenString := req.Header.Get("token")

	// parse token
	token, _ := jwt.ParseWithClaims(tokenString, &types.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("GOBLOG_SIGNING_KEY")), nil
	})

	// checking for claims, not expired, valid, etc.
	if claims, ok := token.Claims.(*types.CustomClaims); ok && token.Valid {
		return claims.Id
	}

	// ParseWithClaims returned error
	return ""
}

// RangeIn returns a random number between low and hi
func RangeIn(low, hi int) int {
	// varying seed makes sure rand's sequence is not fixed
	seed := rand.NewSource(time.Now().UnixNano())
	tempRand := rand.New(seed)

	return low + tempRand.Intn(hi-low)
}

// CreateToken creates a JSON web token storing authorId that expires after 24 hours
func CreateToken(authorId string) (string, error) {
	claims := types.CustomClaims{
		Id: authorId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    "goblog",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("GOBLOG_SIGNING_KEY")))
}

// CreatePostWithoutAuthor removes the Author field from the array of Post
func CreatePostWithoutAuthor(content types.AuthorPosts) interface{} {
	type customAuthor struct {
		Username  string    `json:"username"`
		AuthorId  string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
	}

	type customPost struct {
		Id        string    `json:"id"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	type customStruct struct {
		AuthorInfo customAuthor `json:"author"`
		List       []customPost `json:"posts"`
	}

	var returnVal customStruct

	returnVal.AuthorInfo = customAuthor{
		Username:  content.AuthorInfo.Username,
		AuthorId:  content.AuthorInfo.AuthorId,
		CreatedAt: content.AuthorInfo.CreatedAt,
	}

	returnVal.List = make([]customPost, len(content.List))

	for index, item := range content.List {
		returnVal.List[index].Id = item.Id
		returnVal.List[index].Title = item.Title
		returnVal.List[index].Body = item.Body
		returnVal.List[index].CreatedAt = item.CreatedAt
		returnVal.List[index].UpdatedAt = item.UpdatedAt
	}

	return returnVal
}

// MethodNotAllowedResponse returns a JSON response indicating requested URL cannot be queried with
// the given method type
//
// Example - Doing a POST on a GET only URL
func MethodNotAllowedResponse(res http.ResponseWriter) {
	// set response header's content-type
	res.Header().Set("Content-Type", "application/json")

	res.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: "Method not allowed!"})

	return
}

// NotFoundResponse returns a JSON response indicating requested URL could not be found
func NotFoundResponse(res http.ResponseWriter) {
	// set response header's content-type
	res.Header().Set("Content-Type", "application/json")

	res.WriteHeader(http.StatusNotFound)
	json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: "URL not found!"})

	return
}

// ForbiddenResponse returns a JSON response indicating token provided was invalid or authorId in
// the token didn't match with the postId provided.
//
// Example - Session expired or post's author and logged in author mismatch
func ForbiddenResponse(res http.ResponseWriter) {
	// set response header's content-type
	res.Header().Set("Content-Type", "application/json")

	res.WriteHeader(http.StatusForbidden)
	json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: "Not logged in!"})

	return
}

// InternalServerError returns a JSON response indicating that some unexpected error occurred
func InternalServerErrorResponse(res http.ResponseWriter, message string) {
	// set response header's content-type
	res.Header().Set("Content-Type", "application/json")

	res.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: message})

	return
}

// BadRequestResponse returns a JSON response indicating that some unexpected error occurred
func BadRequestResponse(res http.ResponseWriter, message string) {
	// set response header's content-type
	res.Header().Set("Content-Type", "application/json")

	res.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(res).Encode(types.DefaultResponse{Status: "failure", Message: message})

	return
}
