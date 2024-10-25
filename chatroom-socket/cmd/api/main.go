package main

import (
	"chatroom-socket/internal/server"
	"chatroom-socket/internal/service"
	"chatroom-socket/internal/web/controller"
	"chatroom-socket/internal/web/middleware"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	user                  string
	password              string
	databaseName          string
	host                  string
	sqlPort               string
	jwtSecret             string
	port                  int
	requestTimeoutSeconds int
)

func init() {
	user = os.Getenv("SQL_USER")
	password = os.Getenv("SQL_PASSWORD")
	databaseName = os.Getenv("SQL_DATABASE")
	host = os.Getenv("SQL_HOST")
	sqlPort = os.Getenv("SQL_PORT")
	jwtSecret = os.Getenv("JWT_SECRET")
	requestTimeoutSeconds, _ = strconv.Atoi(os.Getenv("REQUEST_TIMEOUT_SECONDS"))
	port, _ = strconv.Atoi(os.Getenv("PORT"))
}

func gracefulShutdown(apiServer *http.Server) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")
}

func NewServerEngine() *gin.Engine {
	engine := gin.New()

	return engine

}

func main() {
	serverEngine := NewServerEngine()
	websocketRoutes := []string{
		"/ws",
	}

	authMiddleWare := middleware.NewAuthMiddleWare(jwtSecret, websocketRoutes)
	loggingMiddleware := middleware.NewLoggingMiddleware()
	serverEngine.Use(authMiddleWare.Handler())
	serverEngine.Use(loggingMiddleware.Handler())

	sqlConnectionUrl := fmt.Sprintf("user=%s port=%s password=%s dbname=%s host=%s sslmode=disable", user, sqlPort, password, databaseName, host)

	sqlxEngine, err := sqlx.Connect("postgres", sqlConnectionUrl)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Database connected...")

	roomService, err := service.NewRoomService(sqlxEngine)
	if err != nil {
		log.Fatalln(err)
	}
	socketService := service.NewSocketService()
	chatMessageService := service.NewChatMessageService(sqlxEngine, roomService)
	assistantService, err := service.NewAssistantService(sqlxEngine, chatMessageService)
	if err != nil {
		log.Fatalln(err)
	}

	controllers := []server.Controller{
		controller.NewRoomController(serverEngine, roomService, socketService),
		controller.NewSocketController(serverEngine, socketService, roomService, chatMessageService, requestTimeoutSeconds),
		controller.NewChatMessageController(serverEngine, roomService, chatMessageService, requestTimeoutSeconds),
		controller.NewAssistantController(serverEngine, assistantService),
	}

	_server := server.NewServer(serverEngine, port, controllers)

	go gracefulShutdown(_server)

	err = _server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(fmt.Sprintf("http server error: %s", err))
	}
}
