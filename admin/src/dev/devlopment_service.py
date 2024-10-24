from typing import Optional
from uuid import UUID, uuid4
from py_spring_core import Component

from src.repository.models import ChatMessage, ChatRoom
from src.repository.repository import ChatMessageRepository, ChatRoomRepository

class DevelopmentService(Component):
    chat_room_repo: ChatRoomRepository
    chat_message_repo: ChatMessageRepository

    def _create_room(self, room_id: UUID, name: str, owner_id: int) -> ChatRoom:
        return ChatRoom(id= room_id, name= name, owner_id=owner_id)
    

    def create_message(self, room_id: Optional[UUID], sender_id: int, content: str) -> ChatMessage:
        return ChatMessage(
            content= content,
            id= uuid4(),
            room_id= room_id,
            sender_id= sender_id
        )

    def post_construct(self) -> None:
        room_id = uuid4()
        room = self._create_room(room_id,"Test Room", 1)
        if not self.chat_room_repo.is_room_exists(room.name):
            self.chat_room_repo.save_room(room)
            messages = [
            self.create_message(room_id, 1, "Hello")
            ]
            self.chat_message_repo.save_all(messages)
        