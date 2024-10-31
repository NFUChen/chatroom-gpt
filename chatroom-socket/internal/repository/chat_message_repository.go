package repository

import (
	"bytes"
	"chatroom-socket/internal"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"time"
)

type IChatMessageRepository interface {
	GetAllMessagesByRoomId(roomId uuid.UUID, offset uint, limit uint) ([]*ChatMessage, error)
	SaveMessageToRoomId(ctx context.Context, message *ChatMessage) (*ChatMessage, error)
}

type ChatMessageRepository struct {
	httpClient *http.Client
	Engine     *sqlx.DB
}

func NewChatMessageRepository(engine *sqlx.DB) *ChatMessageRepository {
	return &ChatMessageRepository{
		Engine:     engine,
		httpClient: http.DefaultClient,
	}
}

func NewChatMessage(roomId uuid.UUID, senderId int, content string) (*ChatMessage, error) {
	message := &ChatMessage{
		ID:          uuid.New().String(),
		RoomId:      roomId,
		SenderId:    senderId,
		Content:     content,
		CreatedAt:   time.Now().UTC(),
		IsCommitted: false,
	}
	if len(message.Content) == 0 {
		return nil, errors.New("message content is required")
	}
	return message, nil
}

func (repository *ChatMessageRepository) GetAllMessagesByRoomId(roomId uuid.UUID, offset uint, limit uint) ([]*ChatMessage, error) {
	sql := "SELECT * FROM chat_message WHERE room_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	log.Println(sql, roomId, limit, offset)
	var chatMessages []*ChatMessage
	err := repository.Engine.Select(&chatMessages, sql, roomId, limit, offset)
	if err != nil {
		return nil, err
	}

	for _, chatMessage := range chatMessages {
		chatMessage.IsCommitted = true
	}

	return chatMessages, nil
}

func (repository *ChatMessageRepository) SaveMessageToRoomId(ctx context.Context, chatMessage *ChatMessage) (*ChatMessage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		body, err := json.Marshal(chatMessage)
		if err != nil {
			return nil, err
		}
		request, err := http.NewRequest(http.MethodPost, internal.PostSaveMessageApi, bytes.NewReader(body))

		resp, err := repository.httpClient.Do(request)
		if err != nil {
			return nil, err
		}
		responseBody, err := internal.HandleResponse[ChatMessage](resp)
		if err != nil {
			return nil, err
		}
		responseBody.Message.IsCommitted = true

		return &responseBody.Message, nil
	}

}
