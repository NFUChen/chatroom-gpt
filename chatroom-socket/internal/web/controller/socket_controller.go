package controller

import (
	"chatroom-socket/internal/service"
	"chatroom-socket/internal/web"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
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
	Engine                 *gin.Engine
	SocketService          *service.SocketService
	RoomService            *service.RoomService
	ChatMessageService     *service.ChatMessageService
	RequestTimeoutDuration time.Duration
}

// RegisterRoutes registers the WebSocket routes
func (controller *SocketController) RegisterRoutes() {
	// Register the WebSocket handler route
	controller.Engine.GET("/ws", controller.WebSocketHandler)
	controller.Engine.POST("/send_notification", controller.SendNotification)
}

func NewSocketController(engine *gin.Engine, socketService *service.SocketService, roomService *service.RoomService, chatMessageService *service.ChatMessageService, requestTimeoutSeconds int) *SocketController {
	return &SocketController{
		Engine:                 engine,
		SocketService:          socketService,
		RoomService:            roomService,
		ChatMessageService:     chatMessageService,
		RequestTimeoutDuration: time.Duration(requestTimeoutSeconds) * time.Second,
	}
}

func (controller *SocketController) SendNotification(c *gin.Context) {
	var message any
	if err := c.BindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		c.Abort()
		return
	}
	controller.SocketService.SendNotification(message)
}

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
		return
	}
	user, ok := userInterface.(service.User)
	if ok {
		controller.SocketService.AddSocket(conn, user.Id)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Welcome back %s", user.UserName))); err != nil {
			log.Println("Failed to send message to user:", err)
		}
		defer func() {
			if err := controller.SocketService.RemoveSocket(user.Id); err != nil {
				return
			}
			if _, err = controller.RoomService.UserLeaveRoom(user); err != nil {
				return
			}
		}()
	}
	for {
		var socketMessage web.SocketMessage
		err := conn.ReadJSON(&socketMessage)
		if len(socketMessage.Event) == 0 {
			err = fmt.Errorf("event key not foudn in socket message")
		}
		if err != nil {
			message := fmt.Sprintf("Error reading message: %v", err)
			log.Println(message)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				log.Println(err.Error())
			}
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), controller.RequestTimeoutDuration)
		defer cancel()
		if err := controller.ChatMessageService.ReceiveSocketMessage(ctx, user, socketMessage.Event, socketMessage.Message); err != nil {
			message := err.Error()
			if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				log.Println(message)
			}
		}
		log.Printf("Received message: %s, from user:  %v", socketMessage, user.UserName)
	}
}
