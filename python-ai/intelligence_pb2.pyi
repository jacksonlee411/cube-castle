from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class InterpretRequest(_message.Message):
    __slots__ = ("user_text", "session_id")
    USER_TEXT_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    user_text: str
    session_id: str
    def __init__(self, user_text: _Optional[str] = ..., session_id: _Optional[str] = ...) -> None: ...

class InterpretResponse(_message.Message):
    __slots__ = ("intent", "structured_data_json")
    INTENT_FIELD_NUMBER: _ClassVar[int]
    STRUCTURED_DATA_JSON_FIELD_NUMBER: _ClassVar[int]
    intent: str
    structured_data_json: str
    def __init__(self, intent: _Optional[str] = ..., structured_data_json: _Optional[str] = ...) -> None: ...
