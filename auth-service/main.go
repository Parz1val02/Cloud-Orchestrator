package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})
	r.POST("/logout", func(c *gin.Context) {
		var logoutReq struct {
			Username string `json:"username" binding:required`
		}

		if err := c.ShouldBindJSON(&logoutReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "User " + logoutReq.Username + " logged out",
		})
	})

	r.Run(":8081")
}
