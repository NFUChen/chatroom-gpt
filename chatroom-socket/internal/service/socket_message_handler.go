package service

import (
	"context"
	"errors"
	"strings"
)

func (service *ChatMessageService) ReceiveSocketMessage(ctx context.Context, user User, event string, message any) error {
	validEventTypes := service.GetValidEventTypes()
	if !service.IsValidEventType(event) {
		return errors.New("invalid event, please enter one of the following: " + strings.Join(validEventTypes, ","))
	}
	switch event {
	case string(EventSendRegularMessage):
		messageMap, ok := message.(map[string]any)
		if !ok {
			return errors.New("invalid message format")
		}
		content, ok := messageMap["content"].(string)
		if !ok {
			return errors.New("invalid message format, content key not found in message")
		}
		err := service.handleEventSendMessage(ctx, user, content)
		if err != nil {
			return err
		}
	}

	return nil
}

func (service *ChatMessageService) handleEventSendMessage(ctx context.Context, user User, content string) error {
	if _, err := service.SendMessageToRoomId(ctx, user.Id, content); err != nil {
		return err
	}
	return nil
}
