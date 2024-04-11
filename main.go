package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/inghamemerson/eatingisactivism.com/util"
	"github.com/joho/godotenv"
)

var (
	locations []util.Location
	mapboxToken string
	password string
	passwordHash string
	salt string
)

func hashValue(value string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value + salt)))
}

func renderError(c *gin.Context, status int, message string) {
	c.HTML(status, "error.html.tmpl", gin.H{
		"message": message,
	})
}

func isPasswordValid(token string) bool {
	return token == passwordHash
}

func Authenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/login" {
			c.Next()
			return
		}

		token, _ := c.Cookie("_token")

		// if we don't have a token from the cookie, check the URL params for "_token"
		if token == "" {
			token = c.Query("_token")
		}

		if token == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if !isPasswordValid(token) {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// since we have a valid token, set the cookie
		c.SetCookie("_token", token, 3600, "/", "", false, true)

		c.Next()
	}
}

func Router() *gin.Engine {
	r := gin.Default()
	baseTemplatePath := "./templates/layouts/base.html.tmpl"
	unauthedTemplatePath := "./templates/layouts/unauthed.html.tmpl"
	templates := multitemplate.New()
	templates.AddFromFiles("login.html.tmpl", unauthedTemplatePath, "./templates/pages/login.html.tmpl")
	templates.AddFromFiles("home.html.tmpl", baseTemplatePath, "./templates/pages/home.html.tmpl")
	templates.AddFromFiles("error.html.tmpl", baseTemplatePath, "./templates/pages/error.html.tmpl")
	templates.AddFromFiles("location-single.html.tmpl", baseTemplatePath, "./templates/pages/location-single.html.tmpl")

	r.HTMLRender = templates

	// r.StaticFS("/public", http.FS(public))
	r.Static("/public", "./public")

	r.Use(Authenticated())

	r.NoRoute(func(c *gin.Context) {
		renderError(c, http.StatusNotFound, "Page not found")
	})

	r.GET("/", func(c *gin.Context) {
		locations := locations
		locationJSON, _ := json.Marshal(locations)

		c.HTML(http.StatusOK, "home.html.tmpl", gin.H{
			"locations": locations,
			"locationsJSON": string(locationJSON),
			"mapboxToken": mapboxToken,
		})
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html.tmpl", gin.H{})
	})

	r.POST("/login", func(c *gin.Context) {
		pass := c.PostForm("password")
		passHash := hashValue(pass)

		if isPasswordValid(passHash) {
			c.SetCookie("_token", passHash, int(60 * 60 * 24), "/", "", false, true)
			c.Redirect(http.StatusFound, "/")
		} else {
			c.Redirect(http.StatusFound, "/login")
		}
	})

	r.GET("/locations/:location", func(c *gin.Context) {
		locationSlug := c.Param("location")

		if (locationSlug == "") {
			renderError(c, http.StatusNotFound, "Page not found")
		}

		for _, location := range locations {
			if location.Slug == locationSlug {
				c.HTML(http.StatusOK, "location-single.html.tmpl", gin.H{
					"location": location,
				})

				return
			}
		}

		renderError(c, http.StatusNotFound, "Page not found")
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/locations", func(c *gin.Context) {
			c.JSON(http.StatusOK, &locations)
		})
	}

	return r
}

func main() {
	godotenv.Load(".env")
	port := os.Getenv("PORT")
	mode := os.Getenv("GIN_MODE")
	url := os.Getenv("URL")
	key := os.Getenv("KEY")

	if (url == "" || key == "") {
		panic("URL or KEY not found in .env")
	}

	mapboxToken = os.Getenv("MAPBOX_TOKEN")

	if (mapboxToken == "") {
		panic("MAPBOX_TOKEN not found in .env")
	}

	locations = util.GetLocations()
	password = os.Getenv("PASSWORD")
	salt = os.Getenv("SALT")

	if (password == "" || salt == "") {
		panic("PASSWORD or SALT not found in .env")
	}

	passwordHash = hashValue(password)

	// poll for locations every 5 seconds
	go func() {
		for {
			newLocations := util.GetLocations()

			if len(newLocations) > 0 {
				locations = newLocations
			}

			time.Sleep(5 * time.Second)
		}
	}()

	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	if port == "" {
		port = "8080"
	}

	r := Router()
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})

	r.Run(":" + port)
}