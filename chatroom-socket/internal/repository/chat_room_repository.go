package repository

import (
	"bytes"
	"chatroom-socket/internal"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"net/http"
	"time"
)

type RoomType string

const (
	RoomTypePublic  RoomType = "public"
	RoomTypePrivate RoomType = "private"
)

type AddRoomSchema struct {
	Id           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	OwnerId      int       `json:"owner_id"`
	RoomType     RoomType  `json:"room_type"`
	RoomPassword string    `json:"room_password"`
}

type RoomRead struct {
	Id        uuid.UUID  `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	OwnerId   int        `db:"owner_id" json:"owner_id"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
}

type IChatRoomRepository interface {
	GetAllRooms() ([]RoomRead, error)
	CreateNewRoom(ctx context.Context, addRoomSchema *AddRoomSchema) (*RoomRead, error)
}

type ChatRoomRepository struct {
	Engine     *sqlx.DB
	httpClient *http.Client
}

func NewChatRoomRepository(engine *sqlx.DB) *ChatRoomRepository {
	return &ChatRoomRepository{
		Engine:     engine,
		httpClient: http.DefaultClient,
	}
}

func (repository *ChatRoomRepository) GetAllRooms() ([]RoomRead, error) {
	sql := `SELECT * FROM chat_room`
	var rooms []RoomRead
	if err := repository.Engine.Select(&rooms, sql); err != nil {
		return rooms, err
	}
	return rooms, nil
}

func (repository *ChatRoomRepository) CreateNewRoom(context context.Context, addRoomSchema *AddRoomSchema) (*RoomRead, error) {
	select {
	case <-context.Done():
		return nil, context.Err()
	default:
		body, err := json.Marshal(addRoomSchema)
		if err != nil {
			return nil, err
		}

		request, err := http.NewRequest(http.MethodPost, internal.PostAddNewRoom, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		response, err := repository.httpClient.Do(request)
		if err != nil {
			return nil, err
		}

		responseBody, err := internal.HandleResponse[RoomRead](response)
		if err != nil {
			return nil, err
		}
		newRoomRead := responseBody.Message
		return &newRoomRead, nil

	}

}
