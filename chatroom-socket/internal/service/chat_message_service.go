package service

import (
	. "chatroom-socket/internal/repository"
	"context"
	"net/http"
	"slices"
)

type EventType string

const (
	EventSendRegularMessage       EventType = "event_send_regular_message"
	EventSendAssistantChatMessage EventType = "event_send_assistant_chat_message"
)

type ChatMessageService struct {
	RoomService           *RoomService
	httpClient            *http.Client
	chatMessageRepository IChatMessageRepository
}

func NewChatMessageService(roomService *RoomService, messageRepository IChatMessageRepository) *ChatMessageService {
	return &ChatMessageService{RoomService: roomService, httpClient: http.DefaultClient, chatMessageRepository: messageRepository}
}

func (service *ChatMessageService) GetAllMessagesByRoomId(roomId string, offset uint, limit uint) ([]*ChatMessage, error) {
	return service.chatMessageRepository.GetAllMessagesByRoomId(roomId, offset, limit)
}

func (service *ChatMessageService) GetValidEventTypes() []string {
	return []string{
		string(EventSendRegularMessage),
		string(EventSendAssistantChatMessage),
	}
}

func (service *ChatMessageService) IsValidEventType(event string) bool {
	return slices.Contains(service.GetValidEventTypes(), event)
}

func (service *ChatMessageService) SendMessageToRoomId(ctx context.Context, senderId int, content string) (*ChatMessage, error) {
	roomId, err := service.RoomService.GetUserLocation(senderId)
	if err != nil {
		return nil, err
	}

	room, err := service.RoomService.GetRoom(roomId)
	if err != nil {
		return nil, err
	}

	message, err := NewChatMessage(roomId, senderId, content)
	if err != nil {
		return nil, err
	}
	room.broadcastMessage(message)
	message, err = service.chatMessageRepository.SaveMessageToRoomId(ctx, message)
	if err != nil {
		return nil, err
	}
	return message, nil

}
