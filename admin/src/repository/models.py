from datetime import datetime
from enum import Enum
from typing import Optional
from uuid import UUID

from py_spring_model import PySpringModel
from pydantic import field_validator
from sqlalchemy import Column, DateTime, func
from sqlmodel import Field

from src.repository.commons import ChatMessageRead, ChatRoomRead

class RoomType(str, Enum):
    Private = "private"
    Public = "public"

class ChatRoomSetting(PySpringModel, table=True):
    __tablename__: str = "chat_room_settings"
    id: int = Field(primary_key=True)
    room_id: UUID | None = Field(default=None, foreign_key="chat_room.id")
    assistant_rule: str
    room_type: RoomType
    password: str | None = Field(default=None)

    @field_validator("password", mode= "before")
    @classmethod
    def validate_password(cls, value: str) -> str:
        if cls.room_type == RoomType.Private and value is None:
            raise ValueError("Password is required for private rooms")
        return value

    
    


class ChatRoom(PySpringModel, table=True):
    __tablename__: str = "chat_room"
    id: UUID = Field(primary_key=True)
    name: str = Field(unique= True)
    owner_id: int | None = Field(default=None, foreign_key="app_user.id")

    created_at: datetime = Field(default= None,
        sa_column=Column(
            DateTime(timezone=True), server_default=func.now(), nullable=True
        )
    )
    updated_at: Optional[datetime] = Field(default= None,
        sa_column=Column(DateTime(timezone=True), onupdate=func.now(), nullable=True)
    )

    def as_read(self) -> ChatRoomRead:
        return ChatRoomRead(
            id= self.id,
            name= self.name,
            created_at= self.created_at,
            updated_at= self.updated_at
        )


class ChatMessage(PySpringModel, table=True):
    __tablename__: str = "chat_message"
    id: UUID = Field(primary_key=True)
    content: str
    room_id: UUID | None = Field(default=None, foreign_key="chat_room.id")
    sender_id: int | None = Field(default=None, foreign_key="app_user.id")

    created_at: datetime = Field(
        default= None,
        sa_column=Column(
            DateTime(timezone=True), server_default=func.now(), nullable=True
        )
    )
    updated_at: Optional[datetime] = Field(
        default= None,
        sa_column=Column(DateTime(timezone=True), onupdate=func.now(), nullable=True)
    )
    def as_read(self) -> ChatMessageRead:
        return ChatMessageRead(
            id= self.id,
            content= self.content,
            room_id= self.room_id,
            sender_id= self.sender_id,
            created_at= self.created_at,
            updated_at= self.updated_at
        )
