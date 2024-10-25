package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"log"
	"sync"
	"time"
)

type User struct {
	Id       int    `db:"id" json:"id"`
	UserName string `db:"user_name" json:"userName"`
	Password string `db:"password" json:"-"`
	Role     string `db:"role" json:"role"`
	Email    string `db:"email" json:"email"`
}

type SocketUser struct {
	User   User
	Socket *websocket.Conn
}

type RoomRead struct {
	Id        uuid.UUID  `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	OwnerId   int        `db:"owner_id" json:"ownerId"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at" json:"updatedAt"`
}

type Room struct {
	Read           *RoomRead
	Users          map[int]*SocketUser
	MessageChannel chan any
}

func (room *Room) GetSocketUser(userId int) (*SocketUser, error) {
	user, ok := room.Users[userId]
	if !ok {
		return nil, errors.New("socket not found")
	}
	return user, nil
}

func newRoom(read RoomRead) *Room {
	room := &Room{
		Read:           &read,
		Users:          make(map[int]*SocketUser),
		MessageChannel: make(chan any),
	}
	log.Println(fmt.Sprintf("Listen message for broadcasting, room: %v", room.Read.Id))
	go room.ListenMessage()
	return room
}

func (room *Room) BroadCastMessage(message any) {
	room.MessageChannel <- message
}

func (room *Room) ListenMessage() {
	for {
		message := <-room.MessageChannel
		room.broadcastMessage(message)
	}
}

func (room *Room) broadcastMessage(message any) {
	for _, user := range room.Users {
		err := user.Socket.WriteJSON(message)
		if err != nil {
			continue
		}
	}
}

func (room *Room) UserJoin(socket *websocket.Conn, user User) error {
	if _, ok := room.Users[user.Id]; ok {
		return errors.New("user is already joined")
	}
	socketUser := SocketUser{
		User:   user,
		Socket: socket,
	}

	room.Users[user.Id] = &socketUser

	room.broadcastMessage(map[string]string{
		"message": fmt.Sprintf("User: %v joined room", user.UserName),
	})
	return nil
}

func (room *Room) UserLeave(user User) (*SocketUser, error) {
	if _, ok := room.Users[user.Id]; !ok {
		return nil, errors.New("user is not joined any room")
	}
	socketUser := room.Users[user.Id]
	delete(room.Users, user.Id)

	room.broadcastMessage(map[string]string{
		"message": fmt.Sprintf("User: %v left room", user.UserName),
	})
	return socketUser, nil
}

type RoomService struct {
	Engine          *sqlx.DB
	UserLocation    map[int]string
	AllRooms        map[string]*Room
	RoomServiceLock *sync.Mutex
}

func NewRoomService(engin *sqlx.DB) (*RoomService, error) {

	service := &RoomService{
		Engine:          engin,
		UserLocation:    make(map[int]string),
		AllRooms:        make(map[string]*Room),
		RoomServiceLock: new(sync.Mutex),
	}
	roomReads, err := service.GetAllRooms()
	if err != nil {
		return nil, err
	}

	for _, read := range roomReads {
		room := newRoom(read)
		// this is only init on program startup, no need to worry about thread safety
		service.AllRooms[read.Id.String()] = room
	}

	return service, nil
}

func (service *RoomService) GetRoom(roomId string) (*Room, error) {
	if room, ok := service.AllRooms[roomId]; ok {
		return room, nil
	}
	return nil, errors.New("room not found")
}

func (service *RoomService) GetUserLocation(userId int) (string, error) {

	roomId, ok := service.UserLocation[userId]
	if !ok {
		return "", errors.New(fmt.Sprintf("user: %v did not join any room", userId))
	}
	return roomId, nil

}

func (service *RoomService) GetAllRooms() ([]RoomRead, error) {
	sql := `SELECT * FROM chat_room`
	var rooms []RoomRead
	if err := service.Engine.Select(&rooms, sql); err != nil {
		return rooms, err
	}
	return rooms, nil
}

func (service *RoomService) UserJoinRoom(roomId string, user User, socket *websocket.Conn) error {
	service.RoomServiceLock.Lock()
	defer func() {
		fmt.Println("Unlocking the service lock")
		service.RoomServiceLock.Unlock()
	}()

	if joinedRoomId, ok := service.UserLocation[user.Id]; ok {
		return errors.New(fmt.Sprintf("User %v already joined room %v", user.UserName, joinedRoomId))
	}

	room, err := service.GetRoom(roomId)
	if err != nil {
		return fmt.Errorf("room %s does not exist", roomId)
	}

	if err := room.UserJoin(socket, user); err != nil {
		return err
	}

	service.UserLocation[user.Id] = roomId

	return nil
}

func (service *RoomService) UserLeaveRoom(user User) (*SocketUser, error) {
	service.RoomServiceLock.Lock()
	defer func() {
		fmt.Println("Unlocking the service lock")
		service.RoomServiceLock.Unlock()
	}()

	userLocationRoomId, ok := service.UserLocation[user.Id]
	if !ok {
		return nil, errors.New("user is not joined any room")
	}

	room, err := service.GetRoom(userLocationRoomId)
	if err != nil {
		return nil, fmt.Errorf("room %s does not exist", userLocationRoomId)
	}
	socketUser, err := room.UserLeave(user)
	if err != nil {
		return nil, err
	}

	delete(service.UserLocation, user.Id)
	return socketUser, nil
}

func (service *RoomService) UserSwitchRoom(user User, targetRoomId string) error {
	_, err := service.GetRoom(targetRoomId)
	if err != nil {
		return err
	}

	roomId, err := service.GetUserLocation(user.Id)
	if err != nil {
		return fmt.Errorf("user does not join any room, use join room instead")
	}

	if roomId == targetRoomId {
		return errors.New("target room is already joined")
	}

	socketUser, err := service.UserLeaveRoom(user)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if socketUser == nil {
		return errors.New("socket user is not found, consider establish a new socket connection")
	}
	err = service.UserJoinRoom(targetRoomId, socketUser.User, socketUser.Socket)
	if err != nil {
		return err
	}
	return nil

}
