package repository

import (
	"github.com/google/uuid"
	"time"
)

type ChatMessageType string

type RoomType string

const (
	AssistantMessageType ChatMessageType = "assistant" // Message from the assistant.
	HumanMessageType     ChatMessageType = "human"     // Message from the human user.
)

const (
	RoomTypePublic  RoomType = "public"
	RoomTypePrivate RoomType = "private"
)

// User represents a user of the system with necessary details and metadata.
type User struct {
	ID         int    `db:"id" json:"id"`                   // Primary key, auto-incremented integer.
	UserName   string `db:"user_name" json:"user_name"`     // Unique username for each user, required.
	Password   string `db:"password" json:"-"`              // User's password (excluded from JSON).
	Role       string `db:"role" json:"role"`               // User role (e.g., "admin", "user"), required.
	Email      string `db:"email" json:"email"`             // Unique user email, required.
	IsVerified bool   `db:"is_verified" json:"is_verified"` // Verification status, defaults to false.
}

type ChatRoomSettings struct {
	ID            uuid.UUID `json:"id"`
	RoomID        uuid.UUID `json:"room_id"` // Foreign key referencing the Room ID.
	AssistantRule string    `json:"assistant_rule"`
	RoomType      RoomType  `json:"room_type"`
	Password      *string   `json:"-"`
}

type Room struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	Name      string           `json:"name" db:"name"`
	OwnerId   int              `json:"owner_id" db:"owner_id"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time       `json:"updated_at" db:"updated_at"`
	IsDeleted bool             `json:"is_deleted" db:"is_deleted"`
	RoomType  RoomType         `json:"room_type" db:"room_type"`
	Settings  ChatRoomSettings `json:"settings" db:"-"` // Defines one-to-one relationship
}

// ChatMessage represents a message within a chat room.
type ChatMessage struct {
	ID          string          `db:"id" json:"id"`                     // Message ID as primary key.
	Content     string          `db:"content" json:"content"`           // Message content, required.
	RoomId      uuid.UUID       `db:"room_id" json:"room_id"`           // Foreign key referencing the Room ID.
	SenderId    int             `db:"sender_id" json:"sender_id"`       // Foreign key referencing the User ID.
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`     // Timestamp of message creation.
	UpdatedAt   *time.Time      `db:"updated_at" json:"updated_at"`     // Timestamp of the last update (nullable).
	MessageType ChatMessageType `db:"message_type" json:"message_type"` // Type of message (e.g., assistant, human).
	IsCommitted bool            `db:"-" json:"is_committed"`            // Message commit status, defaults to false (excluded from database).
}
