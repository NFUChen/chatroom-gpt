package web

import (
	"chatroom-socket/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	UserKey = "user"
)

func HandleBadRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
	c.Abort()
}

func GetUserFromContext(c *gin.Context) service.User {
	userInterface, ok := c.Get(UserKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"detail": "user not found in context",
		})
		c.Abort()
	}
	user, ok := userInterface.(service.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"detail": "Invalid jwt payload",
		})
		c.Abort()
	}
	return user
}
