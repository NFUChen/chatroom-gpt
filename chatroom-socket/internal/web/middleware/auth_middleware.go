package middleware

import (
	"chatroom-socket/internal/repository"
	"chatroom-socket/internal/web"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const (
	AuthCookieName = "jwt"
	AuthKeyName    = "token"
)

type AuthMiddleWare struct {
	SecretKey       string
	WebSocketRoutes []string
}

func NewAuthMiddleWare(secretKey string, webSocketRoutes []string) *AuthMiddleWare {
	return &AuthMiddleWare{SecretKey: secretKey, WebSocketRoutes: webSocketRoutes}
}

func (middleware *AuthMiddleWare) parseJWT(secretKey string, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

func (middleware *AuthMiddleWare) DeleteCookie(c *gin.Context, cookieName string) {
	c.SetCookie(cookieName, "", -1, "/", "", false, true)
}

func (middleware *AuthMiddleWare) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		isWebSocketRoute := false
		url := c.Request.URL
		for _, route := range middleware.WebSocketRoutes {
			if strings.HasPrefix(url.Path, route) {
				isWebSocketRoute = true
			}
		}
		var tokenString string

		if isWebSocketRoute {
			authToken := c.GetHeader(AuthKeyName)
			if len(authToken) == 0 {
				authToken = c.Query(AuthKeyName)
			}
			isEmptyToken := len(authToken) == 0
			if isEmptyToken {
				c.JSON(http.StatusUnauthorized, gin.H{"detail": "Please pass token into Auth header"})
				c.Abort()
			}
			tokenString = authToken
		} else {
			authCookieToken, err := c.Cookie(AuthCookieName)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"detail": "Please login first"})
				c.Abort()
			}
			tokenString = authCookieToken
		}

		if c.IsAborted() {
			return
		}

		claims, err := middleware.parseJWT(middleware.SecretKey, tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid token"})
			middleware.DeleteCookie(c, AuthCookieName)
			c.Abort()
			return
		}
		// Store the JWT claims in the context
		userId := claims["id"].(float64)
		userName := claims["user_name"].(string)
		email := claims["email"].(string)
		role := claims["role"].(string)

		c.Set(web.UserKey, repository.User{
			ID:       int(userId),
			UserName: userName,
			Role:     role,
			Email:    email,
		},
		)
		// Continue to the next handler
		c.Next()
	}
}
