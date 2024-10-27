package controller

import (
	"chatroom-socket/internal/repository"
	"chatroom-socket/internal/service"
	"chatroom-socket/internal/web"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

type RoomController struct {
	Engine                 *gin.Engine
	RoomService            *service.RoomService
	SocketService          *service.SocketService
	RequestTimeoutDuration time.Duration
}

func NewRoomController(engine *gin.Engine, roomService *service.RoomService, socketService *service.SocketService, requestTimeoutSeconds int) *RoomController {
	return &RoomController{
		Engine:                 engine,
		RoomService:            roomService,
		SocketService:          socketService,
		RequestTimeoutDuration: time.Duration(requestTimeoutSeconds) * time.Second,
	}
}

func (controller *RoomController) RegisterRoutes() {
	controller.Engine.GET("/all_rooms", controller.AllRooms)
	controller.Engine.POST("/user_join_room", controller.UserJoinRoom)
	controller.Engine.GET("/user_leave_room", controller.UserLeaveRoom)
	controller.Engine.POST("/user_switch_room", controller.UserSwitchRoom)
}

func (controller *RoomController) AllRooms(c *gin.Context) {
	rooms, err := controller.RoomService.GetAllRooms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		c.Abort()
	}
	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}

func (controller *RoomController) UserJoinRoom(c *gin.Context) {
	var schema struct {
		RoomId string
	}
	user, err := web.GetUserFromContext(c)
	if err != nil {
		return
	}
	if err := c.BindJSON(&schema); err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	socket, err := controller.SocketService.GetSocketByUserId(user.Id)

	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}

	if err := controller.RoomService.UserJoinRoom(schema.RoomId, *user, socket); err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("join room %s successfully", schema.RoomId),
	})
}

func (controller *RoomController) UserLeaveRoom(c *gin.Context) {
	user, err := web.GetUserFromContext(c)
	if err != nil {
		return
	}

	socketUser, err := controller.RoomService.UserLeaveRoom(*user)

	defer func() {
		if socketUser != nil && socketUser.Socket != nil {
			if closeErr := socketUser.Socket.Close(); closeErr != nil {
				log.Println(closeErr)
			}
		}
	}()

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
		TargetRoomId string `json:"target_room_id"`
	}
	if err := c.BindJSON(&schema); err != nil {
		web.HandleBadRequest(c, err)
		return
	}

	if len(schema.TargetRoomId) == 0 {
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

func (controller *RoomController) AddNewRoom(c *gin.Context) {
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
	addRoomSchema.OwnerId = user.Id
	room, err := controller.RoomService.CreateNewRoom(ctx, addRoomSchema)
	if err != nil {
		web.HandleBadRequest(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": room,
	})
}
