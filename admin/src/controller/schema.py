from typing import Optional
from uuid import UUID

from pydantic import BaseModel

from src.repository.models import RoomType


class ChatRoomSchema(BaseModel):
    id: UUID
    name: str
    room_type: RoomType
    owner_id: int


class ChatMessageSchema(BaseModel):
    id: UUID
    sender_id: int