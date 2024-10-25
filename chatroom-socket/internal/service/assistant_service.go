package service

import (
	"github.com/jmoiron/sqlx"
)

const (
	AssistantName = "assistant"
)

type AssistantMessage struct {
	Message     ChatMessage
	IsFinalWord bool `json:"is_final_word"`
}

type AssistantService struct {
	Engine             *sqlx.DB
	ChatMessageService *ChatMessageService
	AssistantUser      *User
}

func NewAssistantService(engine *sqlx.DB, chatMessageService *ChatMessageService) (*AssistantService, error) {
	service := &AssistantService{
		ChatMessageService: chatMessageService,
		Engine:             engine,
	}
	assistantUser, err := service.GetAssistantUser()
	if err != nil {
		return nil, err
	}
	service.AssistantUser = assistantUser
	return service, nil
}

func (service *AssistantService) GetAssistantUser() (*User, error) {
	sql := "SELECT * FROM app_user WHERE user_name = $1"
	var user User
	if err := service.Engine.Get(&user, sql, AssistantName); err != nil {
		return nil, err
	}
	return &user, nil // Return the pointer to the user
}
