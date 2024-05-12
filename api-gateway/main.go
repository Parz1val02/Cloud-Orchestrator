package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	APIGatewayURL  = "http://localhost:8080"
	AuthserviceURL = "http://localhost:8081"
	APIKey         = "your-api-key"
)

func main() {
	proxy1, err := createReverseProxy(AuthserviceURL)
	if err != nil {
		log.Fatalf("Error creating reverse proxy for authservice: %v", err)
	}

	r := gin.Default()

	r.Use(authenticate(APIKey))
	r.Use(rateLimit())

	r.POST("/login", loginHandler)
	r.Any("/authservice/*path", proxy1)

	log.Printf("API Gateway listening on %s", APIGatewayURL)
	log.Fatal(r.Run(":8080"))
}

func createReverseProxy(targetURL string) (func(*gin.Context), error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(c *gin.Context) {
		log.Printf("Request received for %s", c.Request.URL.Path)

		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/authservice")

		proxy.ServeHTTP(c.Writer, c.Request)
	}, nil
}

func authenticate(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/login" {
			fmt.Println("entrando /login")
			c.Next()
			return
		}
		key := c.GetHeader("X-API-Key")
		if key != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

func rateLimit() gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(1*time.Minute), 5)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}
		c.Next()
	}
}
