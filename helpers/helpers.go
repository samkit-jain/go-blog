package helpers

import (
	"math/rand"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/samkit-jain/go-blog/types"
)

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1

	if i <= 0 {
		return p[1:], "/"
	}

	return p[1:i], p[i:]
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func RangeIn(low, hi int) int {
	seed := rand.NewSource(time.Now().UnixNano())
	tempRand := rand.New(seed)

	return low + tempRand.Intn(hi-low)
}

func InvalidMethod() types.DefaultResponse {
	return types.DefaultResponse{Status: "failure", Message: "Method not allowed!"}
}

func NotFound() types.DefaultResponse {
	return types.DefaultResponse{Status: "failure", Message: "URL not found!"}
}
