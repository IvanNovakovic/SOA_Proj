from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime
from bson import ObjectId

class PyObjectId(ObjectId):
    @classmethod
    def __get_validators__(cls):
        yield cls.validate

    @classmethod
    def validate(cls, v):
        if not ObjectId.is_valid(v):
            raise ValueError('Invalid objectid')
        return ObjectId(v)

class OrderItem(BaseModel):
    id: Optional[str] = None
    tour_id: str
    name: str
    price: float

class ShoppingCart(BaseModel):
    user_id: str
    items: List[OrderItem] = []
    total: float = 0.0

class CheckoutResult(BaseModel):
    user_id: str
    purchased: List[dict]
    failed: List[dict]

class TourPurchaseToken(BaseModel):
    token: str
    user_id: str
    tour_id: str
    created_at: datetime = Field(default_factory=datetime.utcnow)
