package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/renders/multitemplate"
	"github.com/joho/godotenv"
	"github.com/inghamemerson/eatingisactivism.com/util"
)

var locations []util.Location
var mapboxToken string

func renderError(c *gin.Context, status int, message string) {
	c.HTML(status, "error.html.tmpl", gin.H{
		"message": message,
	})
}

// func Authenticated() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		key := c.Query("key")
// 	}
// }

func Router() *gin.Engine {
	r := gin.Default()
	baseTemplatePath := "./templates/layouts/base.html.tmpl"
	templates := multitemplate.New()
	templates.AddFromFiles("home.html.tmpl", baseTemplatePath, "./templates/pages/home.html.tmpl")
	templates.AddFromFiles("error.html.tmpl", baseTemplatePath, "./templates/pages/error.html.tmpl")
	templates.AddFromFiles("location-single.html.tmpl", baseTemplatePath, "./templates/pages/location-single.html.tmpl")

	r.HTMLRender = templates

	// r.StaticFS("/public", http.FS(public))
	r.Static("/public", "./public")

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
	mapboxToken = os.Getenv("MAPBOX_TOKEN")
	locations = util.GetLocations()

	if (url == "" || key == "") {
		panic("URL or KEY not found in .env")
	}

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