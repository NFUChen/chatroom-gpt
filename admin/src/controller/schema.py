from typing import Optional
from uuid import UUID

from pydantic import BaseModel


class ChatRoomSchema(BaseModel):
    id: UUID
    name: str
    owner_id: int


class ChatMessageSchema(BaseModel):
    id: UUID
    sender_id: int