package service

import (
	. "chatroom-socket/internal/repository"
	"github.com/google/uuid"
)

type RoomView struct {
	Id             uuid.UUID `json:"id"`
	NumberOfPeople uint      `json:"number_of_people"`
	RoomName       string    `json:"room_name"`
	RoomType       RoomType  `json:"room_type"`
}
