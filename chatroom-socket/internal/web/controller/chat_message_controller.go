package controller

import (
	"chatroom-socket/internal/repository"
	"chatroom-socket/internal/service"
	"chatroom-socket/internal/web"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"time"
)

type ChatMessageController struct {
	Router                 *gin.RouterGroup
	RoomService            *service.RoomService
	ChatMessageService     *service.ChatMessageService
	RequestTimeoutDuration time.Duration
}

func NewChatMessageController(router *gin.RouterGroup, roomService *service.RoomService, chatMessageService *service.ChatMessageService, requestTimeoutSeconds int) *ChatMessageController {
	return &ChatMessageController{
		Router:                 router,
		RoomService:            roomService,
		ChatMessageService:     chatMessageService,
		RequestTimeoutDuration: time.Duration(requestTimeoutSeconds) * time.Second,
	}
}

func (controller *ChatMessageController) RegisterRoutes() {
	controller.Router.GET("/message/:room_id", controller.GetChatMessagesByRoomId)
	controller.Router.POST("/send_chat_message", controller.SendMessageToRoomId)
}

func (controller *ChatMessageController) GetChatMessagesByRoomId(c *gin.Context) {
	roomId := uuid.MustParse(c.Param("room_id"))
	if roomId == uuid.Nil {
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
	var messageSchema repository.ChatMessage
	if err := c.BindJSON(&messageSchema); err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	user, err := web.GetUserFromContext(c)
	if err != nil {
		return
	}

	chatMessage, err := repository.NewChatMessage(messageSchema.RoomId, user.ID, messageSchema.Content)
	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}

	response, err := controller.ChatMessageService.SendMessageToRoomId(ctx, user.ID, chatMessage.Content)
	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}
