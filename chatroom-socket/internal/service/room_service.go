package service

import (
	. "chatroom-socket/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"slices"
	"sync"
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
	UserLocation       map[int]string
	AllRooms           map[string]*Room
	RoomServiceLock    *sync.Mutex
	httpClient         *http.Client
	chatRoomRepository IChatRoomRepository
}

func (service *RoomService) addRoomToCache(read RoomRead) *Room {

	room := newRoom(read)
	if cachedRoom, exists := service.AllRooms[read.Id.String()]; exists {
		return cachedRoom
	}

	service.RoomServiceLock.Lock()
	defer service.RoomServiceLock.Unlock()

	service.AllRooms[read.Id.String()] = room

	return room
}

func NewRoomService(chatRoomRepository IChatRoomRepository) (*RoomService, error) {

	service := &RoomService{
		chatRoomRepository: chatRoomRepository,
		UserLocation:       make(map[int]string),
		AllRooms:           make(map[string]*Room),
		RoomServiceLock:    new(sync.Mutex),
		httpClient:         http.DefaultClient,
	}
	roomReads, err := service.GetAllRooms()
	if err != nil {
		return nil, err
	}

	for _, read := range roomReads {
		service.addRoomToCache(read)
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
	return service.chatRoomRepository.GetAllRooms()
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

func (service *RoomService) validateAddRomSchema(schema AddRoomSchema) error {
	validRoomTypes := []string{
		string(RoomTypePublic),
		string(RoomTypePrivate),
	}

	if len(schema.Name) == 0 {
		return errors.New("name is required")
	}
	if len(schema.RoomType) == 0 {
		return errors.New("room_type is required")
	}

	if !slices.Contains(validRoomTypes, string(schema.RoomType)) {
		return errors.New("room_type is not valid")
	}

	if schema.RoomType == RoomTypePrivate && len(schema.RoomPassword) == 0 {
		return errors.New("room_password is required for private room")
	}
	return nil
}

func (service *RoomService) CreateNewRoom(ctx context.Context, addRoomSchema AddRoomSchema) (*Room, error) {
	if err := service.validateAddRomSchema(addRoomSchema); err != nil {
		return nil, err
	}

	newRoomRead, err := service.chatRoomRepository.CreateNewRoom(ctx, &addRoomSchema)
	if err != nil {
		return nil, err
	}
	cachedRoom := service.addRoomToCache(*newRoomRead)
	return cachedRoom, nil
}
