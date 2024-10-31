package service

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type EventType string

const (
	EventSendRegularMessage       EventType = "event_send_regular_message"
	EventSendAssistantChatMessage EventType = "event_send_assistant_chat_message"
	EventRoomSendMessage          EventType = "event_room_send_message"
	EventUserJoinRoom             EventType = "event_user_join_room"
	EventUserLeftRoom             EventType = "event_user_left_room"
	EventNotification             EventType = "event_notification"
	EventGreeting                 EventType = "event_greeting"
)

type SocketMessage struct {
	Event   EventType `json:"event"`
	Content string    `json:"content"`
}

func NewSocketMessage(event EventType, content string) *SocketMessage {
	return &SocketMessage{
		Event:   event,
		Content: content,
	}
}

type SocketService struct {
	UserSockets map[int]*websocket.Conn
	ServiceLock *sync.Mutex
}

func NewSocketService() *SocketService {
	return &SocketService{
		UserSockets: make(map[int]*websocket.Conn),
		ServiceLock: new(sync.Mutex),
	}
}

func (service *SocketService) AddSocket(socket *websocket.Conn, userId int) {
	service.ServiceLock.Lock()
	defer service.ServiceLock.Unlock()

	service.UserSockets[userId] = socket

}

func (service *SocketService) SendNotification(message *SocketMessage) {
	for userId, socket := range service.UserSockets {
		err := socket.WriteJSON(message)
		if err != nil {
			log.Printf("socket write json error for user %d: %v", userId, err)
			continue
		}
	}
}

func (service *SocketService) GetSocketByUserId(userId int) (*websocket.Conn, error) {
	socket, ok := service.UserSockets[userId]
	if !ok {
		return nil, errors.New("user not found in all sockets")
	}
	return socket, nil
}

func (service *SocketService) RemoveSocket(userId int) error {
	service.ServiceLock.Lock()
	defer service.ServiceLock.Unlock()

	socket, ok := service.UserSockets[userId]
	if !ok {
		return fmt.Errorf("user %d not exist", userId)
	}
	delete(service.UserSockets, userId)
	err := socket.Close()
	if err != nil {
		return fmt.Errorf("socket close error: %s", err.Error())
	}
	return nil
}
