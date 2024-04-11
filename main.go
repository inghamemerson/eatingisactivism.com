package main

import (
	"os"

	"eatingisactivism/app/router"

	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
)

func main() {
	godotenv.Load(".env")
	port := os.Getenv("PORT")
	mode := os.Getenv("GIN_MODE")

	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	if port == "" {
		port = "8080"
	}

	r := router.Router()
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})

	r.Run(":" + port)
}