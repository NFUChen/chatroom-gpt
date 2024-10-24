

from datetime import datetime
from typing import Optional
from uuid import UUID
from pydantic import BaseModel


class ChatRoomRead(BaseModel):
    id: UUID
    name: str
    created_at: datetime
    updated_at: Optional[datetime]


class ChatMessageRead(BaseModel):
    id: UUID
    content: str
    room_id: Optional[UUID]
    sender_id: Optional[int]
    created_at: datetime
    updated_at: Optional[datetime]