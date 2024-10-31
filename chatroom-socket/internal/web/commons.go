package web

import (
	"chatroom-socket/internal/repository"
	"errors"
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

func GetUserFromContext(c *gin.Context) (*repository.User, error) {
	userInterface, ok := c.Get(UserKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"detail": "user not found in context",
		})
		c.Abort()
		return nil, errors.New("user not found in context")
	}
	user, ok := userInterface.(repository.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"detail": "Invalid jwt payload",
		})
		c.Abort()
		return nil, errors.New("Invalid jwt payload")
	}
	return &user, nil
}
