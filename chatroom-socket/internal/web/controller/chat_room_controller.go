package controller

import (
	"chatroom-socket/internal/repository"
	"chatroom-socket/internal/service"
	"chatroom-socket/internal/web"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

type RoomController struct {
	Engine                 *gin.Engine
	Router                 *gin.RouterGroup
	RoomService            *service.RoomService
	SocketService          *service.SocketService
	RequestTimeoutDuration time.Duration
}

func NewRoomController(router *gin.RouterGroup, roomService *service.RoomService, socketService *service.SocketService, requestTimeoutSeconds int) *RoomController {
	return &RoomController{
		Router:                 router,
		RoomService:            roomService,
		SocketService:          socketService,
		RequestTimeoutDuration: time.Duration(requestTimeoutSeconds) * time.Second,
	}
}

func (controller *RoomController) RegisterRoutes() {

	controller.Router.GET("/all_rooms", controller.AllRooms)
	controller.Router.POST("/user_join_room", controller.UserJoinRoom)
	controller.Router.GET("/user_leave_room", controller.UserLeaveRoom)
	controller.Router.POST("/user_switch_room", controller.UserSwitchRoom)
	controller.Router.GET("/user_location", controller.UserLocation)
	controller.Router.POST("/chat_room", controller.CreateNewRoom)
	controller.Router.GET("/chat_room_settings/:room_id", controller.GetChatRoomSettings)
	controller.Router.DELETE("/chat_room/:room_id", controller.DeleteRoom)
}

func (controller *RoomController) UserLocation(c *gin.Context) {
	user, err := web.GetUserFromContext(c)
	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	roomId, err := controller.RoomService.GetUserLocation(user.ID)
	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":     user,
		"location": roomId,
	})
}

func (controller *RoomController) AllRooms(c *gin.Context) {
	rooms := controller.RoomService.GetAllRoomViews()
	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}

func (controller *RoomController) GetChatRoomSettings(c *gin.Context) {
	roomId := uuid.MustParse(c.Param("room_id"))
	if roomId == uuid.Nil {
		web.HandleBadRequest(c, fmt.Errorf("room_id is empty"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), controller.RequestTimeoutDuration)
	defer cancel()
	settings, err := controller.RoomService.GetChatRoomSettings(ctx, roomId)
	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
	})

}

func (controller *RoomController) UserJoinRoom(c *gin.Context) {
	var schema struct {
		RoomId uuid.UUID `json:"room_id"`
	}
	user, err := web.GetUserFromContext(c)
	if err != nil {
		return
	}
	if err := c.BindJSON(&schema); err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	socket, err := controller.SocketService.GetSocketByUserId(user.ID)

	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}

	if err := controller.RoomService.UserJoinRoom(schema.RoomId, *user, socket); err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"roomId":  schema.RoomId,
		"message": fmt.Sprintf("join room %s successfully", schema.RoomId),
	})
}

func (controller *RoomController) UserLeaveRoom(c *gin.Context) {
	user, err := web.GetUserFromContext(c)
	if err != nil {
		return
	}

	_, err = controller.RoomService.UserLeaveRoom(*user)

	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user leave room successfully",
	})
}

func (controller *RoomController) UserSwitchRoom(c *gin.Context) {
	user, err := web.GetUserFromContext(c)
	if err != nil {
		return
	}

	var schema struct {
		TargetRoomId uuid.UUID `json:"target_room_id"`
	}
	if err := c.BindJSON(&schema); err != nil {
		web.HandleBadRequest(c, err)
		return
	}

	if schema.TargetRoomId == uuid.Nil {
		web.HandleBadRequest(c, fmt.Errorf("target_room_id is empty"))
		return
	}

	if err := controller.RoomService.UserSwitchRoom(*user, schema.TargetRoomId); err != nil {
		web.HandleBadRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user switch room successfully",
	})

}

func (controller *RoomController) CreateNewRoom(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), controller.RequestTimeoutDuration)
	defer cancel()

	user, err := web.GetUserFromContext(c)
	if err != nil {
		return
	}
	var addRoomSchema repository.AddRoomSchema
	if err := c.BindJSON(&addRoomSchema); err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	addRoomSchema.OwnerId = user.ID
	log.Println(fmt.Sprintf("user: %v created room: %v", addRoomSchema.OwnerId, addRoomSchema.Name))
	room, err := controller.RoomService.CreateNewRoom(ctx, &addRoomSchema)
	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	log.Println(fmt.Sprintf("SocketRoom created: %v", room.Read.ID))
	c.JSON(http.StatusOK, gin.H{"message": room.Read})
}

func (controller *RoomController) DeleteRoom(c *gin.Context) {
	roomId := uuid.MustParse(c.Param("room_id"))
	if roomId == uuid.Nil {
		web.HandleBadRequest(c, fmt.Errorf("room_id is empty"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), controller.RequestTimeoutDuration)
	defer cancel()

	user, err := web.GetUserFromContext(c)
	if err != nil {
		web.HandleBadRequest(c, err)
	}
	if err := controller.RoomService.DeleteRoom(ctx, user.ID, roomId); err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "delete room successfully"})

}
