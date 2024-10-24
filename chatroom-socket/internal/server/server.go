package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int
}

type Controller interface {
	RegisterRoutes()
}

func NewServer(engine *gin.Engine, port int, controllers []Controller) *http.Server {
	newServer := &Server{
		port: port,
	}
	for _, controller := range controllers {
		if ctl, ok := controller.(Controller); ok {
			ctl.RegisterRoutes()
		}

	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      engine,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
