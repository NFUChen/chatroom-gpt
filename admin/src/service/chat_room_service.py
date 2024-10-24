from uuid import UUID
from py_spring_core import Component

from src.repository.commons import ChatRoomRead
from src.repository.models import ChatRoom
from src.repository.repository import ChatRoomRepository

class ChatRoomService(Component):
    chat_room_repo: ChatRoomRepository
    def get_all_rooms(self) -> list[ChatRoom]:
        return self.chat_room_repo.find_all()
    
    def add_room(self, room_id: UUID, name: str, owner_id: int) -> ChatRoomRead:
        room = ChatRoom(id= room_id,name=name, owner_id= owner_id)
        return self.chat_room_repo.save_room(room).as_read()
    
    def update_room_name(self, room_id: UUID, name: str) -> ChatRoomRead:
        return self.chat_room_repo.update_room_name(room_id, name).as_read()
    