package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	dbMap = make(map[string]string)

	mutex sync.RWMutex
)

func generateShortURL() string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	short := make([]byte, 6)
	for i := range short {
		short[i] = charset[rand.Intn(len(charset))]
	}
	return string(short)
}

func generateUniqueShortURL() string {
	var shortURL string
	mutex.RLock()
	defer mutex.RUnlock()

	for {
		shortURL = generateShortURL()

		if _, exists := dbMap[shortURL]; !exists {
			break
		}
	}

	return shortURL
}

func redirectURL(c *gin.Context) {
	shortURL := c.Param("short_url")

	mutex.RLock()
	originalURL, exists := dbMap[shortURL]
	mutex.RUnlock()
	fmt.Println(originalURL)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// Redirect to the original URL
	http.Redirect(c.Writer, c.Request, originalURL, http.StatusFound)
}

// Short URL Function
func shortenURL(c *gin.Context) {

	var reqBody struct {
		URL string `json:"url"`
	}

	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	shortURL := generateUniqueShortURL()
	mutex.Lock()
	defer mutex.Unlock()
	dbMap[shortURL] = reqBody.URL
	log.Printf("Shortened URL: http://localhost:8080/%s", shortURL)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"short_url": fmt.Sprintf("http://localhost:8080/%s", shortURL),
	})

}

func printUrls(c *gin.Context) {
	mutex.RLock()
	defer mutex.RUnlock()

	urls := make([]string, 0, len(dbMap))
	for shortURL, originalURL := range dbMap {
		urls = append(urls, fmt.Sprintf("http://localhost:8080/%s", shortURL))
		urls = append(urls, originalURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"urls":    urls,
	})

}

func main() {
	router := gin.Default()
	fmt.Println("Hi ")
	// To Check if the server is up or not
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	router.POST("/shorten", shortenURL)

	router.GET("/:short_url", redirectURL)

	router.GET("/all_urls", printUrls)

	router.Run(":8085")
}
