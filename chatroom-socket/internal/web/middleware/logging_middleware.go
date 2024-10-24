package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (writer bodyLogWriter) Write(b []byte) (int, error) {
	writer.body.Write(b)
	return writer.ResponseWriter.Write(b)
}

type LoggingMiddleware struct{}

func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

func (middleware *LoggingMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		bodyLogWriter := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyLogWriter

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		log.Printf("| %3d | %13v | %15s | %s | %s | %s",
			statusCode,
			duration,
			c.ClientIP(),
			c.Request.Method,
			c.Request.RequestURI,
			bodyLogWriter.body.String(),
		)
	}
}
