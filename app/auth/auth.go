package auth

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	password string
	passwordHash string
	salt string
)

func init() {
	godotenv.Load(".env")

	password = os.Getenv("PASSWORD")
	salt = os.Getenv("SALT")

	if (password == "" || salt == "") {
		panic("PASSWORD or SALT not found in .env")
	}

	passwordHash = HashValue(password)
}

func HashValue(value string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value + salt)))
}

func IsPasswordValid(token string) bool {
	return token == passwordHash
}

func renderUnauthJSON(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"message": message,
	})
	c.Abort()
	return
}

func renderUnauthHTML(c *gin.Context, message string) {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
}


func isAuthed(c *gin.Context) bool {
	token, _ := c.Cookie("_token")

	// if we don't have a token from the cookie, check the URL params for "_token"
	if token == "" {
		token = c.Query("_token")
	}

	if (IsPasswordValid(token)) {
		return true
	}

	return false
}

func AuthHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/login" {
			c.Next()
			return
		}

		if c.Request.URL.Path == "/favicon.ico" {
			c.Next()
			return
		}

		if c.Request.URL.Path == "/public" {
			c.Next()
			return
		}

		if (isAuthed(c)) {
			token, _ := c.Cookie("_token")

			// if we don't have a token from the cookie, check the URL params for "_token"
			if token == "" {
				token = c.Query("_token")
			}

			c.SetCookie("_token", token, int(60 * 60 * 24), "/", "", false, true)
			c.Next()
			return
		}

		renderUnauthHTML(c, "Unauthorized")
	}
}

func AuthJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		if (isAuthed(c)) {
			c.Next()
			return
		}

		renderUnauthJSON(c, "Unauthorized")
	}
}