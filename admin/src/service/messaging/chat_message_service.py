from loguru import logger
from py_spring_core import Component
from src.repository.commons import ChatMessageRead
from src.repository.models import ChatMessage
from src.repository.repository import ChatMessageRepository

class ChatMessageService(Component):
    message_repo: ChatMessageRepository

    def save_message(self, message: ChatMessage) -> ChatMessageRead:
        message_read = self.message_repo.save(message).as_read()
        logger.info(f"[MESSAGE SAVING] Message: {message_read} saved.")
        return message_read