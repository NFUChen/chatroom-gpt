from uuid import UUID
from fastapi import status
from py_spring_core import RestController

from src.repository.models import ChatMessage
from src.service.chat_message_service import ChatMessageService
from src.controller.commons import ResponseEntity
from src.service.chat_room_service import ChatRoomService
from src.controller.schema import ChatRoomSchema

class ChatRoomController(RestController):
    class Config:
        prefix: str = "/api/internal"

    chat_room_service: ChatRoomService
    chat_message_service: ChatMessageService
    def register_routes(self) -> None:
        
        @self.router.post("/chat_room")
        def add_chatroom(room: ChatRoomSchema) -> ResponseEntity:
            saved_room = self.chat_room_service.add_room(
                room_id= room.id, 
                name= room.name, 
                owner_id= room.owner_id
            )
            return ResponseEntity(status= status.HTTP_201_CREATED, message= saved_room)
        
        @self.router.put("/chat_room/{room_id}")
        def update_room_name(room_id: UUID, name: str) -> ResponseEntity:
            updated_room = self.chat_room_service.update_room_name(room_id= room_id, name= name)
            return ResponseEntity(status= status.HTTP_200_OK, message=updated_room)
        
        @self.router.post("/send_chat_message")
        def save_message(message: ChatMessage) -> ResponseEntity:
            saved_message = self.chat_message_service.save_message(message)
            return ResponseEntity(status= status.HTTP_201_CREATED, message= saved_message)
