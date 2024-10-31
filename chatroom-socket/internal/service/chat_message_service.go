package service

import (
	. "chatroom-socket/internal/repository"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"slices"
)

type ChatMessageService struct {
	RoomService           *RoomService
	httpClient            *http.Client
	chatMessageRepository IChatMessageRepository
}

func NewChatMessageService(roomService *RoomService, messageRepository IChatMessageRepository) *ChatMessageService {
	return &ChatMessageService{RoomService: roomService, httpClient: http.DefaultClient, chatMessageRepository: messageRepository}
}

func (service *ChatMessageService) GetAllMessagesByRoomId(roomId uuid.UUID, offset uint, limit uint) ([]*ChatMessage, error) {
	return service.chatMessageRepository.GetAllMessagesByRoomId(roomId, offset, limit)
}

func (service *ChatMessageService) GetValidEventTypes() []string {
	return []string{
		string(EventSendRegularMessage),
		string(EventSendAssistantChatMessage),
	}
}

func (service *ChatMessageService) IsValidEventType(event EventType) bool {
	return slices.Contains(service.GetValidEventTypes(), string(event))
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

	body, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	room.broadcastMessage(NewSocketMessage(EventRoomSendMessage, string(body)))
	message, err = service.chatMessageRepository.SaveMessageToRoomId(ctx, message)
	if err != nil {
		return nil, err
	}
	return message, nil

}
