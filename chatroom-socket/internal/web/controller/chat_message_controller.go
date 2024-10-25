package controller

import (
	"chatroom-socket/internal/service"
	"chatroom-socket/internal/web"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type ChatMessageController struct {
	Engine                 *gin.Engine
	RoomService            *service.RoomService
	ChatMessageService     *service.ChatMessageService
	RequestTimeoutDuration time.Duration
}

func NewChatMessageController(engine *gin.Engine, roomService *service.RoomService, chatMessageService *service.ChatMessageService, requestTimeoutSeconds int) *ChatMessageController {
	return &ChatMessageController{
		Engine:                 engine,
		RoomService:            roomService,
		ChatMessageService:     chatMessageService,
		RequestTimeoutDuration: time.Duration(requestTimeoutSeconds) * time.Second,
	}
}

func (controller *ChatMessageController) RegisterRoutes() {
	controller.Engine.GET("/message/:room_id", controller.GetChatMessagesByRoomId)
	controller.Engine.POST("/send_chat_message", controller.SendMessageToRoomId)
}

func (controller *ChatMessageController) GetChatMessagesByRoomId(c *gin.Context) {
	roomId := c.Param("room_id")
	if len(roomId) == 0 {
		web.HandleBadRequest(c, errors.New("room_id is required"))
		return
	}

	limit, err := strconv.ParseUint(c.Query("message_limit"), 10, 64)
	if err != nil || limit == 0 {
		web.HandleBadRequest(c, errors.New("message limit must be provided, and can't be 0"))
		return
	}
	messageOffset, err := strconv.ParseUint(c.Query("message_offset"), 10, 64)
	if err != nil {
		web.HandleBadRequest(c, errors.New("message offset must be provided"))
		return
	}

	messages, err := controller.ChatMessageService.GetAllMessagesByRoomId(roomId, uint(messageOffset), uint(limit))
	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (controller *ChatMessageController) SendMessageToRoomId(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), controller.RequestTimeoutDuration)
	defer cancel()
	var messageSchema service.ChatMessage
	if err := c.BindJSON(&messageSchema); err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	user := web.GetUserFromContext(c)

	chatMessage, err := service.NewChatMessage(messageSchema.RoomId, messageSchema.SenderId, messageSchema.Content)
	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}

	response, err := controller.ChatMessageService.SendMessageToRoomId(ctx, user.Id, chatMessage.Content)
	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}
