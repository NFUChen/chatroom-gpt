
from uuid import UUID

from loguru import logger
from py_spring_model import CrudRepository, PySpringModel
from sqlalchemy import text

from src.errors import RoomNotFoundError
from src.repository.models import ChatMessage, ChatRoom



class ChatRoomRepository(CrudRepository[UUID, ChatRoom]):
    def is_room_exists(self, room_name: str) -> bool:
        _, optional_room = self._find_by_query({"name": room_name})
        return optional_room is not None
    
    def save_room(self, room: ChatRoom) -> ChatRoom:
        with PySpringModel.create_managed_session() as session:
            session.add(room)
        optional_room = self.find_by_id(room.id, session)
        if optional_room is None:
            raise RoomNotFoundError(f"Room not found: {room.id}")
        
        session.add(optional_room)

        return optional_room
    
    def update_room_name(self, room_id: UUID, name: str) -> ChatRoom:
        with PySpringModel.create_managed_session() as session:
            optional_room = self.find_by_id(room_id, session)
            if optional_room is None:
                raise RoomNotFoundError(f"Room not found: {room_id}")
            optional_room.name = name
            session.add(optional_room)        
            return optional_room


    def post_construct(self) -> None:
        table_name = ChatRoom.__tablename__
        logger.info("[INDEX FOR CHATROOM] Creating chatroom index...")
        with PySpringModel.create_managed_session() as session:
            session.exec(
                text(f"CREATE INDEX IF NOT EXISTS {table_name} ON chat_room (owner_id)")  # type: ignore
            )


class ChatMessageRepository(CrudRepository[UUID, ChatMessage]):
    def post_construct(self) -> None:
        table_name = ChatMessage.__tablename__
        logger.info("[INDEX FOR CHATROOM] Creating chatroom index...")
        with PySpringModel.create_managed_session() as session:
            session.exec(
                text(
                    f"CREATE INDEX IF NOT EXISTS {table_name} ON chat_message (room_id)"
                )  # type: ignore
            )
            session.exec(
                text(
                    f"CREATE INDEX IF NOT EXISTS {table_name} ON chat_message (sender_id)"
                )  # type: ignore
            )
