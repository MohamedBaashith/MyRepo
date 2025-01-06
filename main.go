package main

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// CORS middleware configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Route for authentication microservice
	router.Any("/auth/*path", func(c *gin.Context) {
		proxyRequest(c, "http://auth-server:5050")
	})

	// Route for CRUD microservice
	router.Any("/crud/*path", func(c *gin.Context) {
		proxyRequest(c, "http://crud-server:6000")
	})

	// Route for event capturing microservice
	router.Any("/event/*path", func(c *gin.Context) {
		proxyRequest(c, "http://event-server:6050")
	})

	// Checking for the Health of the MainServer
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Route to Search microservice
	router.Any("/search/*path", func(c *gin.Context) {
		proxyRequest(c, "http://search-server:7000")
	})

	router.Run(":5000")
}

func proxyRequest(c *gin.Context, target string) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	targetURL := target + c.Param("path")
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	for k, v := range c.Request.Header {
		if k != "Connection" && k != "Transfer-Encoding" {
			req.Header[k] = v
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reach microservice"})
		return
	}
	defer resp.Body.Close()

	c.Status(resp.StatusCode)
	for k, v := range resp.Header {
		if k == "Access-Control-Allow-Origin" || k == "Authorization" {
			c.Header(k, v[0])
		}
	}

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to forward response"})
	}
}

