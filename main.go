package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/inghamemerson/eatingisactivism.com/util"
)

var locations []util.Location
var mapboxToken string

func Router() *gin.Engine {
	r := gin.Default()

	// r.StaticFS("/public", http.FS(public))
	r.Static("/public", "./public")

	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		locations := util.GetLocations()
		locationJSON, _ := json.Marshal(locations)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"locations": locations,
			"locationsJSON": string(locationJSON),
			"mapboxToken": mapboxToken,
		})
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
	mapboxToken = os.Getenv("MAPBOX_TOKEN")


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