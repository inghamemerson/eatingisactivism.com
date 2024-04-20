package main

import (
	"bytes"
	"net/http"
	"os"
	"fmt"

	"eatingisactivism/app/router"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func purgeCloudflare() {
	fmt.Println("Purging Cloudflare cache")
	godotenv.Load(".env")
	token := os.Getenv("CLOUDFLARE_TOKEN")
	url := os.Getenv("CLOUDFLARE_CACHE_URL")

	if (token == "" || url == "") {
		fmt.Println("CLOUDFLARE_TOKEN or CLOUDFLARE_CACHE_URL not found in .env")
		return
	}

	reqBody := []byte(`{"purge_everything":true}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))

	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + token)

	client := &http.Client{}
	_, err = client.Do(req)

	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	return
}

func main() {
	godotenv.Load(".env")
	port := os.Getenv("PORT")
	mode := os.Getenv("GIN_MODE")

	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
		purgeCloudflare()
	}

	if port == "" {
		port = "8080"
	}

	r := router.Router()
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})

	r.Run(":" + port)
}