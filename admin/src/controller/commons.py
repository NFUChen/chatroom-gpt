
from typing import Any
from pydantic import BaseModel

class ResponseEntity(BaseModel):
    message: Any
    status: int
