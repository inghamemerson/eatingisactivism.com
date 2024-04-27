package router

import (
	"encoding/json"
	"net/http"
	"os"
	"io"
	"strings"

	"eatingisactivism/app/auth"
	"eatingisactivism/app/locations"

	healthcheck "github.com/RaMin0/gin-health-check"
	brotli "github.com/anargu/gin-brotli"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/semihalev/gin-stats"
)

var (
	mapboxToken string
)

func init() {
	godotenv.Load(".env")
	mapboxToken = os.Getenv("MAPBOX_TOKEN")

	if (mapboxToken == "") {
		panic("MAPBOX_TOKEN not found in .env")
	}
}

func staticCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if (strings.HasPrefix(c.Request.URL.Path, "/public")) {
			c.Header("Cache-Control", "public, max-age=31536000")
		}
		c.Next()
	}
}

func renderHTMLError(c *gin.Context, status int, message string) {
	c.HTML(status, "error.html.tmpl", gin.H{
		"status": status,
		"message": message,
	})
	c.Abort()
}

func renderJSONError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"status": status,
		"message": message,
	})
	c.Abort()
}

func LoadTemplates() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	baseTemplatePath := "./app/templates/layouts/base.html.tmpl"
	unauthedTemplatePath := "./app/templates/layouts/unauthed.html.tmpl"

	r.AddFromFiles("login.html.tmpl", unauthedTemplatePath, "./app/templates/pages/login.html.tmpl")
	r.AddFromFiles("home.html.tmpl", baseTemplatePath, "./app/templates/pages/home.html.tmpl")
	r.AddFromFiles("error.html.tmpl", baseTemplatePath, "./app/templates/pages/error.html.tmpl")
	r.AddFromFiles("location-single.html.tmpl", baseTemplatePath, "./app/templates/pages/location-single.html.tmpl")
	r.AddFromFiles("locations.html.tmpl", baseTemplatePath, "./app/templates/pages/locations.html.tmpl")

	return r
}

func Router() *gin.Engine {
	r := gin.Default()

	r.HTMLRender = LoadTemplates()

	r.Use(brotli.Brotli(brotli.DefaultCompression))
	r.Use(healthcheck.Default())

	r.NoRoute(func(c *gin.Context) {
		// of the request is to the /api path, return a JSON error
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			renderJSONError(c, http.StatusNotFound, "Page not found")
			return
		}

		renderHTMLError(c, http.StatusNotFound, "Page not found")
	})

	// if GIN_MODE is release, we need to compress static assets and set cache headers
	if gin.Mode() == gin.ReleaseMode {
		r.Use(staticCacheMiddleware())
	}

	r.Static("/public", "./public")

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html.tmpl", gin.H{})
	})

	r.POST("/login", func(c *gin.Context) {
		pass := c.PostForm("password")
		passHash := auth.HashValue(pass)

		if auth.IsPasswordValid(passHash) {
			c.SetCookie("_token", passHash, int(60 * 60 * 24), "/", "", false, true)
			c.Redirect(http.StatusFound, "/")
		} else {
			c.Redirect(http.StatusFound, "/login")
		}
	})

	authorized := r.Group("/", auth.AuthHTML())
	{
		authorized.GET("/", func(c *gin.Context) {
			locs := locations.GetLocations()
			locationJSON, _ := json.Marshal(locs)

			c.HTML(http.StatusOK, "home.html.tmpl", gin.H{
				"locations": locs,
				"standards": locations.GetStandards(),
				"tags": locations.GetTags(),
				"locationsJSON": string(locationJSON),
				"mapboxToken": mapboxToken,
			})
		})

		authorized.GET("/locations", func(c *gin.Context) {
			c.HTML(http.StatusOK, "locations.html.tmpl", gin.H{
				"locations": locations.GetLocations(),
			})
		})

		authorized.GET("/locations/:location", func(c *gin.Context) {
			locationSlug := c.Param("location")
			location := locations.GetLocationBySlug(locationSlug)

			if (locationSlug == "" || location.Slug == "") {
				renderHTMLError(c, http.StatusNotFound, "Page not found")
				return
			}

			c.HTML(http.StatusOK, "location-single.html.tmpl", gin.H{
				"location": location,
			})
		})
	}

	v1 := r.Group("/api/v1", auth.AuthJSON())
	{
		v1.Use(stats.RequestStats())

		v1.GET("/stats", func(c *gin.Context) {
			c.JSON(http.StatusOK, stats.Report())
		})

		v1.GET("/locations", func(c *gin.Context) {
			tagsParam := c.Query("tags")
			tags := []string{}

			standardsParam := c.Query("standards")
			standards := []string{}

			if (tagsParam != "") {
				tags = strings.Split(tagsParam, ",")
			}

			if (standardsParam != "") {
				standards = strings.Split(standardsParam, ",")
			}

			var locs = locations.LocationMap{}

			if (len(tags) != 0 || len(standards) != 0) {
				locs = locations.FilterLocations(standards, tags)
			} else {
				locs = locations.GetLocations()
			}

			c.JSON(http.StatusOK, locs)
		})

		// route to accept webhook from contentful
		v1.POST("/webhook", func(c *gin.Context) {

			jsonData, err := io.ReadAll(c.Request.Body)
			topic := c.GetHeader("X-Contentful-Topic")

			if err != nil {
				renderJSONError(c, http.StatusBadRequest, "Error reading request body")
				return
			}

			locations.HandleWebhook(topic, jsonData)
		})
	}

	return r
}
