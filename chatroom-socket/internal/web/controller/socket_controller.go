package controller

import (
	"chatroom-socket/internal/service"
	"chatroom-socket/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by default
		return true
	},
}

type SocketController struct {
	Engine        *gin.Engine
	SocketService *service.SocketService
}

// RegisterRoutes registers the WebSocket routes
func (controller *SocketController) RegisterRoutes() {
	// Register the WebSocket handler route
	controller.Engine.GET("/ws", controller.WebSocketHandler)
	controller.Engine.POST("/send_notification", controller.SendNotification)
}

func NewSocketController(engine *gin.Engine, socketService *service.SocketService) *SocketController {
	return &SocketController{
		Engine:        engine,
		SocketService: socketService,
	}
}

func (controller *SocketController) SendNotification(c *gin.Context) {
	var message any
	if err := c.BindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		c.Abort()
	}
	controller.SocketService.SendNotification(message)
}

// WebSocketHandler handles WebSocket requests from clients
func (controller *SocketController) WebSocketHandler(c *gin.Context) {
	// Upgrade the HTTP request to a WebSocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	userInterface, ok := c.Get(web.UserKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"detail": "user not found in context",
		})
		c.Abort()
	}
	user, ok := userInterface.(service.User)
	if ok {
		controller.SocketService.AddSocket(conn, user.Id)
		defer controller.SocketService.RemoveSocket(user.Id)
	}
	// Handle WebSocket messages
	for {
		// Read message from client
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		log.Printf("Received message: %s", message)

		// Echo message back to client
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}
}
