package repository

import (
	"bytes"
	"chatroom-socket/internal"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type AddRoomSchema struct {
	Id           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	OwnerId      int       `json:"owner_id"`
	RoomType     RoomType  `json:"room_type"`
	RoomPassword string    `json:"room_password"`
	SettingsId   uuid.UUID `json:"settings_id"`
}

type IChatRoomRepository interface {
	GetAllRooms() ([]Room, error)
	CreateNewRoom(ctx context.Context, addRoomSchema *AddRoomSchema) (*Room, error)
	DeleteRoom(ctx context.Context, roomId uuid.UUID) error
	GetRoomSettings(ctx context.Context, roomId uuid.UUID) (*ChatRoomSettings, error)
	GetRoomById(ctx context.Context, roomId uuid.UUID) (*Room, error)
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

func (repository *ChatRoomRepository) GetRoomById(ctx context.Context, roomId uuid.UUID) (*Room, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		sql := `SELECT cr.*,
				   crs.room_type
			FROM   chat_room cr
				   JOIN chat_room_settings crs
					 ON cr.id = crs.room_id
			WHERE  is_deleted = false AND crs.room_id = $1`
		var roomRead Room
		if err := repository.Engine.Get(&roomRead, sql, roomId); err != nil {
			return nil, err
		}
		return &roomRead, nil
	}
}

func (repository *ChatRoomRepository) GetAllRooms() ([]Room, error) {
	sql := `SELECT cr.*,
				   crs.room_type
			FROM   chat_room cr
				   JOIN chat_room_settings crs
					 ON cr.id = crs.room_id
			WHERE  is_deleted = false`
	var rooms []Room
	if err := repository.Engine.Select(&rooms, sql); err != nil {
		return rooms, err
	}
	return rooms, nil
}

func (repository *ChatRoomRepository) CreateNewRoom(ctx context.Context, addRoomSchema *AddRoomSchema) (*Room, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		body, err := json.Marshal(addRoomSchema)
		if err != nil {
			return nil, err
		}

		request, err := http.NewRequest(http.MethodPost, internal.ChatRoomApi, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		response, err := repository.httpClient.Do(request)
		if err != nil {
			return nil, err
		}

		responseBody, err := internal.HandleResponse[Room](response)
		if err != nil {
			return nil, err
		}
		newRoomRead := responseBody.Message
		return &newRoomRead, nil

	}
}

func (repository *ChatRoomRepository) DeleteRoom(ctx context.Context, roomId uuid.UUID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		sql := "UPDATE chat_room SET is_deleted = TRUE WHERE id = $1"
		_, err := repository.Engine.Exec(sql, roomId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repository *ChatRoomRepository) GetRoomSettings(ctx context.Context, roomId uuid.UUID) (*ChatRoomSettings, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		sql := "SELECT * FROM chat_room_settings WHERE room_id = $1"
		var roomSettings ChatRoomSettings
		if err := repository.Engine.Get(&roomSettings, sql, roomId); err != nil {
			return nil, err
		}
		return &roomSettings, nil
	}
}
