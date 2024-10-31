package service

import (
	. "chatroom-socket/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"slices"
	"sync"
)

type SocketUser struct {
	User   User
	Socket *websocket.Conn
}

func (user *SocketUser) SendMessage(socketMessage *SocketMessage) error {
	if err := user.Socket.WriteJSON(socketMessage); err != nil {
		return err
	}
	return nil
}

type SocketRoom struct {
	Read           *Room
	Users          map[int]*SocketUser
	MessageChannel chan *SocketMessage
	RoomContext    context.Context
	NumberOfPeople uint
}

func (service *RoomService) GetAllRoomViews() []RoomView {
	var roomViews []RoomView
	for _, room := range service.AllRooms {
		roomViews = append(roomViews, RoomView{
			Id:             room.Read.ID,
			NumberOfPeople: room.NumberOfPeople,
			RoomName:       room.Read.Name,
			RoomType:       room.Read.RoomType,
		})
	}
	return roomViews
}

func (room *SocketRoom) GetSocketUser(userId int) (*SocketUser, error) {
	user, ok := room.Users[userId]
	if !ok {
		return nil, errors.New("socket not found")
	}
	return user, nil
}

func newRoom(read Room) *SocketRoom {
	ctx := context.Background()
	room := &SocketRoom{
		Read:           &read,
		Users:          make(map[int]*SocketUser),
		MessageChannel: make(chan *SocketMessage),
		RoomContext:    ctx,
	}
	log.Println(fmt.Sprintf("Listen message for broadcasting, room: %v", room.Read.ID))
	go room.ListenMessage(ctx)
	return room
}

func (room *SocketRoom) AllUsersLeave() {
	notifyUserFunc := func(wg *sync.WaitGroup, user *SocketUser) {
		defer wg.Done()
		if err := user.Socket.WriteMessage(websocket.TextMessage, []byte("System calling for all user leave")); err != nil {
			log.Println(err)
		}
		if err := user.Socket.Close(); err != nil {
			log.Println(err)
		}
	}

	wg := new(sync.WaitGroup)
	numberOfUsers := len(room.Users)
	wg.Add(numberOfUsers)
	for _, user := range room.Users {
		go notifyUserFunc(wg, user)
	}

	wg.Wait()

	clear(room.Users)
	close(room.MessageChannel)
	room.NumberOfPeople = 0
	room.RoomContext.Done()
}

func (room *SocketRoom) BroadCastMessage(message *SocketMessage) {
	room.MessageChannel <- message
}

func (room *SocketRoom) ListenMessage(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			return
		default:
			message := <-room.MessageChannel
			room.broadcastMessage(message)
		}
	}
}

func (room *SocketRoom) StopMessageListening() {
	log.Println("Sending Done to room context")
	room.RoomContext.Done()
}

func (room *SocketRoom) broadcastMessage(message *SocketMessage) {
	for _, user := range room.Users {
		err := user.Socket.WriteJSON(message)
		if err != nil {
			continue
		}
	}
}

func (room *SocketRoom) UserJoin(socket *websocket.Conn, user User) error {
	if _, ok := room.Users[user.ID]; ok {
		return errors.New("user is already joined")
	}
	socketUser := SocketUser{
		User:   user,
		Socket: socket,
	}

	room.Users[user.ID] = &socketUser

	room.broadcastMessage(
		NewSocketMessage(EventUserJoinRoom, fmt.Sprintf("User: %v joined room", user.UserName)),
	)

	room.NumberOfPeople++
	return nil
}

func (room *SocketRoom) UserLeave(user User) (*SocketUser, error) {
	if _, ok := room.Users[user.ID]; !ok {
		return nil, errors.New("user does not join any room")
	}
	socketUser := room.Users[user.ID]
	delete(room.Users, user.ID)
	room.broadcastMessage(
		NewSocketMessage(EventUserJoinRoom, fmt.Sprintf("User: %v left room", user.UserName)),
	)
	room.NumberOfPeople--
	return socketUser, nil
}

type RoomService struct {
	UserLocation       map[int]uuid.UUID
	AllRooms           map[uuid.UUID]*SocketRoom
	RoomServiceLock    *sync.Mutex
	httpClient         *http.Client
	chatRoomRepository IChatRoomRepository
}

func (service *RoomService) addRoomToCache(read Room) *SocketRoom {

	if cachedRoom, exists := service.AllRooms[read.ID]; exists {
		return cachedRoom
	}

	room := newRoom(read)

	service.RoomServiceLock.Lock()
	defer service.RoomServiceLock.Unlock()
	log.Println(fmt.Sprintf("Adding room cache: %v", read.ID))
	service.AllRooms[read.ID] = room

	go func() {
		err := service.UpdateRoom(context.Background(), read.ID)
		if err != nil {
			return
		}
	}()

	return room
}

func (service *RoomService) removeRoomFromCache(read Room) {
	service.RoomServiceLock.Lock()
	defer service.RoomServiceLock.Unlock()
	log.Println(fmt.Sprintf("Removing room from cache: %v", read.ID))
	delete(service.AllRooms, read.ID)
}

func NewRoomService(chatRoomRepository IChatRoomRepository) (*RoomService, error) {

	service := &RoomService{
		chatRoomRepository: chatRoomRepository,
		UserLocation:       make(map[int]uuid.UUID),
		AllRooms:           make(map[uuid.UUID]*SocketRoom),
		RoomServiceLock:    new(sync.Mutex),
		httpClient:         http.DefaultClient,
	}
	rooms, err := service.chatRoomRepository.GetAllRooms()
	if err != nil {
		return nil, err
	}

	for _, read := range rooms {
		service.addRoomToCache(read)
	}

	return service, nil
}

func (service *RoomService) GetRoom(roomId uuid.UUID) (*SocketRoom, error) {
	if room, ok := service.AllRooms[roomId]; ok {
		return room, nil
	}
	return nil, errors.New("room not found")
}

func (service *RoomService) GetUserLocation(userId int) (uuid.UUID, error) {

	roomId, ok := service.UserLocation[userId]
	if !ok {
		return uuid.Nil, errors.New(fmt.Sprintf("user: %v did not join any room", userId))
	}
	return roomId, nil

}

func (service *RoomService) GetAllRooms() []Room {
	var rooms []Room

	for _, room := range service.AllRooms {
		rooms = append(rooms, *room.Read)
	}

	return rooms
}

func (service *RoomService) GetChatRoomSettings(ctx context.Context, roomId uuid.UUID) (*ChatRoomSettings, error) {
	return service.chatRoomRepository.GetRoomSettings(ctx, roomId)
}

func (service *RoomService) UserJoinRoom(roomId uuid.UUID, user User, socket *websocket.Conn) error {
	service.RoomServiceLock.Lock()
	func() {
		fmt.Println("Unlocking the service lock")
		service.RoomServiceLock.Unlock()
	}()

	if err := service.UnsafeUserJoinRoom(roomId, user, socket); err != nil {
		return err
	}
	return nil
}

func (service *RoomService) UnsafeUserJoinRoom(roomId uuid.UUID, user User, socket *websocket.Conn) error {
	if joinedRoomId, ok := service.UserLocation[user.ID]; ok {
		log.Println(fmt.Sprintf("User %v already joined room %v", user.UserName, joinedRoomId))
		if _, err := service.UnsafeUserLeaveRoom(user); err != nil {
			return err
		}
	}

	room, err := service.GetRoom(roomId)
	if err != nil {
		return fmt.Errorf("room %s does not exist", roomId)
	}

	if err := room.UserJoin(socket, user); err != nil {
		return err
	}
	service.UserLocation[user.ID] = roomId
	log.Println(fmt.Sprintf("user %v left room: %v", user.UserName, roomId))

	return nil
}

func (service *RoomService) UnsafeUserLeaveRoom(user User) (*SocketUser, error) {
	userLocationRoomId, ok := service.UserLocation[user.ID]
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

	delete(service.UserLocation, user.ID)
	log.Println(fmt.Sprintf("user %v left room: %v", user.UserName, userLocationRoomId))
	return socketUser, nil
}

func (service *RoomService) UserLeaveRoom(user User) (*SocketUser, error) {
	service.RoomServiceLock.Lock()
	defer func() {
		fmt.Println("Unlocking the service lock")
		service.RoomServiceLock.Unlock()
	}()
	socketUser, err := service.UnsafeUserLeaveRoom(user)
	if err != nil {
		return nil, err
	}
	return socketUser, nil
}

func (service *RoomService) UserSwitchRoom(user User, targetRoomId uuid.UUID) error {
	log.Println(fmt.Sprintf("User: %v switch room to %v", user.UserName, targetRoomId))
	_, err := service.GetRoom(targetRoomId)
	if err != nil {
		return err
	}

	roomId, err := service.GetUserLocation(user.ID)
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

func (service *RoomService) validateAddRomSchema(schema *AddRoomSchema) error {
	validRoomTypes := []string{
		string(RoomTypePublic),
		string(RoomTypePrivate),
	}

	if schema.Id == uuid.Nil {
		return errors.New(fmt.Sprintf("SocketRoom id %v is invalid", schema.Id))
	}
	if schema.SettingsId == uuid.Nil {
		return errors.New(fmt.Sprintf("SocketRoom settings id %v is invalid", schema.Id))
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

func (service *RoomService) CreateNewRoom(ctx context.Context, addRoomSchema *AddRoomSchema) (*SocketRoom, error) {
	if addRoomSchema.Id == uuid.Nil {
		addRoomSchema.Id = uuid.New()
	}

	if addRoomSchema.SettingsId == uuid.Nil {
		addRoomSchema.SettingsId = uuid.New()
	}
	if err := service.validateAddRomSchema(addRoomSchema); err != nil {
		return nil, err
	}

	_newRoom, err := service.chatRoomRepository.CreateNewRoom(ctx, addRoomSchema)
	if err != nil {
		return nil, err
	}
	cachedRoom := service.addRoomToCache(*_newRoom)
	return cachedRoom, nil
}

func (service *RoomService) DeleteRoom(ctx context.Context, userId int, roomId uuid.UUID) error {
	room, err := service.GetRoom(roomId)
	if err != nil {
		return err
	}

	if room.Read.OwnerId != userId {
		return errors.New(fmt.Sprintf("room %v does not own user %v", roomId, userId))
	}
	room.StopMessageListening()
	err = service.chatRoomRepository.DeleteRoom(ctx, roomId)
	if err != nil {
		return err
	}
	room.AllUsersLeave()
	service.removeRoomFromCache(*room.Read)
	return nil
}

func (service *RoomService) UpdateRoom(ctx context.Context, roomId uuid.UUID) error {
	log.Println(fmt.Sprintf("Update room %v", roomId))
	room, err := service.chatRoomRepository.GetRoomById(ctx, roomId)
	if err != nil {
		return err
	}
	socketRoom, ok := service.AllRooms[room.ID]
	if !ok {
		return errors.New(fmt.Sprintf("room %v does not exist", room.ID))
	}

	socketRoom.Read = room
	return nil

}
