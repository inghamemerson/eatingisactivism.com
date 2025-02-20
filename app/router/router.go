package router

import (
	"encoding/json"
	"net/http"
	"os"
	"io"
	"strings"
	"slices"
	"strconv"
	"html/template"

	"eatingisactivism/app/auth"
	"eatingisactivism/app/locations"
	"eatingisactivism/app/seasons"

	healthcheck "github.com/RaMin0/gin-health-check"
	brotli "github.com/anargu/gin-brotli"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/semihalev/gin-stats"
	"github.com/unrolled/render"
)

var (
	mapboxToken string
	environment string
)

func init() {
	godotenv.Load(".env")
	mapboxToken = os.Getenv("MAPBOX_TOKEN")
	environment = os.Getenv("GIN_MODE")

	if (mapboxToken == "") {
		panic("MAPBOX_TOKEN not found in .env")
	}
}

// function takes a string and returns HTML
// This will be used inside of templates
func safeHTML(s string) template.HTML {
	return template.HTML(s)
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
	c.HTML(status, "pages/error", gin.H{
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

func Router() *gin.Engine {
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"safeHTML": safeHTML,
	})
	r.LoadHTMLGlob("templates/**/*.tmpl")

	renderer := render.New(render.Options{
		Extensions: []string{".tmpl"},
		Funcs: []template.FuncMap{
			{
				"safeHTML": safeHTML,
			},
		},
		IndentJSON: true,
		IsDevelopment: environment != "release",
		Layout: "layout",
	})

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
		renderer.HTML(c.Writer, http.StatusOK, "pages/login", gin.H{})
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

			renderer.HTML(c.Writer, http.StatusOK, "pages/home", gin.H{
				"locations": locs,
				"states": seasons.States,
				"seasons": seasons.Seasons,
				"standards": locations.GetStandards(),
				"tags": locations.GetTags(),
				"locationsJSON": string(locationJSON),
				"mapboxToken": mapboxToken,
			})
		})

		authorized.GET("/locations", func(c *gin.Context) {
			renderer.HTML(c.Writer, http.StatusOK, "pages/locations", gin.H{
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

			renderer.HTML(c.Writer, http.StatusOK, "pages/location-single", gin.H{
				"location": location,
			})
		})

		authorized.GET("/foods", func(c *gin.Context) {
			// log the request
			state := c.Query("state") // string
			season := c.Query("season") // int
			nextSeasonInt := 0;

			inSeason := map[string]string{}
			nextSeason := map[string]string{}

			// if we don't have a season or a state param, we return a 204
			if (state != "" && season != "") {
				// cast season to int
				seasonInt, err := strconv.Atoi(season)

				if (err != nil) {
					seasonInt = 0
				}

				// if season is 24, we need to set next season to 1
				if (seasonInt == 24) {
					nextSeasonInt = 1
				} else {
					nextSeasonInt = seasonInt + 1
				}

				inSeasonFoods := seasons.GetFoodsByStateAndSeason(state, seasonInt)
				nextSeasonFoods := seasons.GetFoodsByStateAndSeason(state, nextSeasonInt)

				for _, food := range inSeasonFoods {
					inSeason[food.Slug] = food.Name
				}

				for _, food := range nextSeasonFoods {
					nextSeason[food.Slug] = food.Name
				}
			}

			tmplFile := "food-items.tmpl"

			c.HTML(http.StatusOK, tmplFile, gin.H{
				"inSeason": inSeason,
				"nextSeason": nextSeason,
			})
		})
	}

	v1 := r.Group("/api/v1", auth.AuthJSON())
	{
		v1.Use(stats.RequestStats())

		v1.GET("/stats", func(c *gin.Context) {
			renderer.JSON(c.Writer, http.StatusOK, stats.Report())
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

			renderer.JSON(c.Writer, http.StatusOK, locs)
		})

		v1.GET("/foods", func(c *gin.Context) {
			renderer.JSON(c.Writer, http.StatusOK, seasons.GetFoods())
		})

		v1.GET("/seasons/:season", func(c *gin.Context) {
			season := c.Param("season")

			if (season == "") {
				renderJSONError(c, http.StatusBadRequest, "Season not provided")
				return
			}

			seasonInt, err := strconv.Atoi(season)

			if (err != nil || seasonInt < 1 || seasonInt > 24) {
				renderJSONError(c, http.StatusBadRequest, "Invalid state")
				return
			}

			foods := seasons.GetFoodsBySeason(seasonInt)

			renderer.JSON(c.Writer, http.StatusOK, foods)
		})

		v1.GET("/states/:state", func(c *gin.Context) {
			state := c.Param("state")

			if (state == "") {
				renderJSONError(c, http.StatusBadRequest, "Season not provided")
				return
			}

			foods := seasons.GetFoodsByState(state)

			renderer.JSON(c.Writer, http.StatusOK, foods)
		})

		v1.GET("/states/:state/seasons/:season", func(c *gin.Context) {
			state := c.Param("state")
			season := c.Param("season")

			if (state == "" || season == "") {
				renderJSONError(c, http.StatusBadRequest, "Invalid state or season")
				return
			}

			seasonInt, err := strconv.Atoi(season)

			if (err != nil || seasonInt < 1 || seasonInt > 24) {
				renderJSONError(c, http.StatusBadRequest, "Invalid state")
				return
			}

			// check if state is in the list of Valid States
			// first uppercase the state
			state = strings.ToUpper(state)
			validStates := seasons.ValidStates()

			validState := slices.Contains(validStates, state)

			if (!validState) {
				renderJSONError(c, http.StatusBadRequest, "Invalid state")
				return
			}

			foods := seasons.GetFoodsByStateAndSeason(state, seasonInt)

			renderer.JSON(c.Writer, http.StatusOK, foods)
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
