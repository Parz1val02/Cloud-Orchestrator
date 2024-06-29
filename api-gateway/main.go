package main

import (
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
	APIGatewayURL      = "http://localhost:4444"
	TemplateserviceURL = "http://localhost:5000"
	SliceserviceURL    = "http://localhost:9999"
)

func main() {
	proxy2, err2 := createReverseProxy(TemplateserviceURL)
	if err2 != nil {
		log.Fatalf("Error creating reverse proxy for templateservice: %v", err2)
	}

	proxySliceManager, errSliceManager := createReverseProxySliceManager(SliceserviceURL)
	if errSliceManager != nil {
		log.Fatalf("Error creating reverse proxy for sliceservice: %v", errSliceManager)
	}

	r := gin.Default()
	r.Use(authenticate())
	r.Use(rateLimit())
	r.POST("/login", loginHandler)
	// r.POST("/logout", logoutHandler)
	r.Any("/templateservice/*path", proxy2)
	r.Any("/sliceservice/*path", proxySliceManager)

	log.Printf("API Gateway listening on %s", APIGatewayURL)
	log.Fatal(r.Run(":4444"))
}

func createReverseProxy(targetURL string) (func(*gin.Context), error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(c *gin.Context) {
		log.Printf("Request received for %s", c.Request.URL.Path)

		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/templateservice")
		if user, exists := c.Get("user"); exists {
			userInfo := user.(UserResponse)
			c.Request.Header.Set("X-User-ID", userInfo.ID)
			c.Request.Header.Set("X-User-Username", userInfo.Username)
			c.Request.Header.Set("X-User-Role", userInfo.Role)
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}, nil
}

func createReverseProxySliceManager(targetURL string) (func(*gin.Context), error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(c *gin.Context) {
		log.Printf("Request received for %s", c.Request.URL.Path)

		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/sliceservice")
		if user, exists := c.Get("user"); exists {
			userInfo := user.(UserResponse)
			c.Request.Header.Set("X-User-ID", userInfo.ID)
			c.Request.Header.Set("X-User-Username", userInfo.Username)
			c.Request.Header.Set("X-User-Role", userInfo.Role)
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}, nil
}

func authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/login" || c.Request.URL.Path == "/logout" {
			c.Next()
			return
		}
		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Token not present"})
			return
		}

		// Validate jwt
		user, err := validateToken(key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func rateLimit() gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(1*time.Second), 5)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}
		c.Next()
	}
}
