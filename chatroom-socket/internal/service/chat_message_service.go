package service

import (
	"bytes"
	"chatroom-socket/internal"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"time"
)

const (
	BaseURL        = "http://0.0.0.0:8085/api/internal"
	SaveMessageApi = BaseURL + "/send_chat_message"
)

type ChatMessage struct {
	Id          string     `db:"id" json:"id"`
	Content     string     `db:"content" json:"content"`
	RoomId      string     `db:"room_id" json:"room_id"`
	SenderId    int        `db:"sender_id" json:"sender_id"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at"`
	IsCommitted bool       `json:"is_committed" db:"-"` // Exclude from db mapping

}

type ChatMessageService struct {
	Engine      *sqlx.DB
	RoomService *RoomService
	httpClient  *http.Client
}

func NewChatMessageService(engine *sqlx.DB, roomService *RoomService) *ChatMessageService {
	return &ChatMessageService{Engine: engine, RoomService: roomService, httpClient: http.DefaultClient}
}

func (service *ChatMessageService) GetAllMessagesByRoomId(roomId string, offset uint, limit uint) ([]*ChatMessage, error) {
	sql := "SELECT * FROM chat_message WHERE room_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	log.Println(sql, roomId, limit, offset)
	var chatMessages []*ChatMessage
	err := service.Engine.Select(&chatMessages, sql, roomId, limit, offset)
	if err != nil {
		return nil, err
	}

	for _, chatMessage := range chatMessages {
		chatMessage.IsCommitted = true
	}

	return chatMessages, nil
}

func (service *ChatMessageService) SendMessageToRoomId(ctx context.Context, senderId int, content string) (*ChatMessage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		roomId, err := service.RoomService.GetUserLocation(senderId)
		if err != nil {
			return nil, err
		}

		room, err := service.RoomService.GetRoom(roomId)
		if err != nil {
			return nil, err
		}

		schema := ChatMessage{
			Id:          uuid.New().String(),
			RoomId:      roomId,
			SenderId:    senderId,
			Content:     content,
			CreatedAt:   time.Now().UTC(),
			IsCommitted: false,
		}

		room.broadcastMessage(schema)

		body, err := json.Marshal(schema)
		if err != nil {
			return nil, err
		}
		request, err := http.NewRequest(http.MethodPost, SaveMessageApi, bytes.NewReader(body))

		resp, err := service.httpClient.Do(request)
		if err != nil {
			return nil, err
		}
		responseBody, err := internal.HandleResponse[ChatMessage](resp)
		if err != nil {
			return nil, err
		}
		responseBody.Message.IsCommitted = true
		room.broadcastMessage(responseBody.Message)
		return &responseBody.Message, nil
	}

}
