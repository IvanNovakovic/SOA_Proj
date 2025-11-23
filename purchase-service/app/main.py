import os
import uuid
from fastapi import FastAPI, HTTPException, Depends
from fastapi.middleware.cors import CORSMiddleware
from typing import List
from pymongo import MongoClient
from pymongo.errors import PyMongoError
from .models import OrderItem, ShoppingCart, CheckoutResult, TourPurchaseToken
from datetime import datetime
from bson import ObjectId
import requests
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
import jwt
import logging

# Basic logging for debug (local dev)
logging.basicConfig(level=logging.INFO)

MONGO_URI = os.getenv('MONGO_URI', 'mongodb://mongo:27017')
TOUR_DB = os.getenv('TOUR_DB', 'tours')
PURCHASE_DB = os.getenv('PURCHASE_DB', 'purchases')
TOUR_SERVICE_BASE = os.getenv('TOUR_SERVICE_BASE', 'http://tour-service:8083')

app = FastAPI(title="purchase-service")
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Mongo setup
client = MongoClient(MONGO_URI)
purchase_db = client[PURCHASE_DB]
cart_col = purchase_db['carts']
tokens_col = purchase_db['tokens']
purchases_col = purchase_db['purchases']

# JWT auth
security = HTTPBearer()


def get_current_user(credentials: HTTPAuthorizationCredentials = Depends(security)) -> str:
    """Verify JWT bearer token and return user id string.

    Looks for `sub`, `user_id`, or `id` in token payload.
    Environment variables:
      - JWT_SECRET (default 'secret')
      - JWT_ALGORITHM (default 'HS256')
    """
    token = credentials.credentials
    secret = os.getenv('JWT_SECRET', 'dev-secret-change-me')
    alg = os.getenv('JWT_ALGORITHM', 'HS256')
    aud = os.getenv('JWT_AUDIENCE')
    iss = os.getenv('JWT_ISSUER')
    try:
        # inspect header first and ensure algorithm matches expected
        try:
            unhdr = jwt.get_unverified_header(token)
            logging.info("JWT header: %s", unhdr)
        except Exception as he:
            logging.exception("Failed to parse JWT header: %s", he)
            raise HTTPException(status_code=401, detail='invalid token')
        if unhdr.get('alg') != alg:
            logging.warning("JWT alg mismatch: header=%s expected=%s", unhdr.get('alg'), alg)
            raise HTTPException(status_code=401, detail='invalid token algorithm')
        # more tolerant audience handling:
        # - if JWT_AUDIENCE is set, prefer to validate it against token 'aud'
        # - tokens issued by Go services may contain aud: [""] when env var is empty
        #   so accept that as a wildcard for backward compatibility
        unverified = jwt.decode(token, options={"verify_signature": False})
        token_aud = unverified.get('aud')
        aud_env = aud if aud and str(aud).strip() != '' else None
        if aud_env:
            # normalize token_aud to list
            tlist = token_aud if isinstance(token_aud, list) else ([token_aud] if token_aud is not None else [])
            # treat a single empty-string aud as 'no audience' (back-compat)
            if tlist == [""]:
                # accept and only verify signature (skip audience check)
                payload = jwt.decode(token, secret, algorithms=[alg], options={"verify_aud": False})
            elif aud_env in tlist:
                payload = jwt.decode(token, secret, algorithms=[alg], audience=aud_env, issuer=iss if iss else None)
            else:
                logging.warning("JWT audience mismatch; token aud=%s expected=%s", tlist, aud_env)
                raise HTTPException(status_code=401, detail='invalid token audience')
        else:
            # no expected audience configured — verify signature only
            payload = jwt.decode(token, secret, algorithms=[alg], options={"verify_aud": False})
        user_id = payload.get('sub') or payload.get('user_id') or payload.get('id') or payload.get('uid')
        if not user_id:
            raise HTTPException(status_code=401, detail='invalid token payload')
        return str(user_id)
    except jwt.ExpiredSignatureError:
        logging.exception("JWT expired for token: %s", token)
        raise HTTPException(status_code=401, detail='token expired')
    except jwt.PyJWTError as e:
        logging.exception("JWT decode failed (%s): %s", type(e).__name__, str(e))
        raise HTTPException(status_code=401, detail='invalid token')
    except Exception as e:
        logging.exception("Unexpected error while decoding JWT (%s): %s", type(e).__name__, str(e))
        raise HTTPException(status_code=401, detail='invalid token')

# Helper functions
def recalc_total(items: List[dict]) -> float:
    return sum(float(i.get('price', 0)) for i in items)


def _stringify_objectids(value):
    """Recursively convert ObjectId instances to strings in dicts/lists."""
    if isinstance(value, ObjectId):
        return str(value)
    if isinstance(value, dict):
        return {k: _stringify_objectids(v) for k, v in value.items()}
    if isinstance(value, list):
        return [_stringify_objectids(v) for v in value]
    return value

def get_tour(tour_id: str):
    # Read tour directly from tour-service database to check archived status
    # Fallback: try calling gateway/tour-service HTTP if exists
    try:
        # Read from tour mongo directly
        tour_db = client[TOUR_DB]
        tours = tour_db['tours']
        from bson import ObjectId
        obj = ObjectId(tour_id)
        t = tours.find_one({'_id': obj})
        return t
    except Exception:
        # try HTTP lookup via tour-service (if endpoint provided)
        try:
            r = requests.get(f"{TOUR_SERVICE_BASE}/tours/{tour_id}", timeout=3)
            if r.status_code == 200:
                return r.json()
        except Exception:
            return None
    return None

@app.post('/cart/items', response_model=ShoppingCart)
def add_item(item: OrderItem, current_user: str = Depends(get_current_user)):
    # Add item to cart and recalc total (user from token)
    try:
        user_id = current_user
        cart = cart_col.find_one({'user_id': user_id})
        it = item.dict()
        it['id'] = str(uuid.uuid4())
        if cart:
            cart['items'].append(it)
            cart['total'] = recalc_total(cart['items'])
            cart_col.update_one({'user_id': user_id}, {'$set': {'items': cart['items'], 'total': cart['total']}})
        else:
            cart = {'user_id': user_id, 'items': [it], 'total': recalc_total([it])}
            cart_col.insert_one(cart)
        cart_safe = _stringify_objectids(cart)
        return ShoppingCart(**cart_safe)
    except PyMongoError as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.delete('/cart/items/{item_id}', response_model=ShoppingCart)
def remove_item(item_id: str, current_user: str = Depends(get_current_user)):
    user_id = current_user
    cart = cart_col.find_one({'user_id': user_id})
    if not cart:
        raise HTTPException(status_code=404, detail='cart not found')
    items = [i for i in cart.get('items', []) if i.get('id') != item_id]
    cart['items'] = items
    cart['total'] = recalc_total(items)
    cart_col.update_one({'user_id': user_id}, {'$set': {'items': items, 'total': cart['total']}})
    cart_safe = _stringify_objectids(cart)
    return ShoppingCart(**cart_safe)

@app.get('/cart', response_model=ShoppingCart)
def get_cart(current_user: str = Depends(get_current_user)):
    user_id = current_user
    cart = cart_col.find_one({'user_id': user_id})
    if not cart:
        return ShoppingCart(user_id=user_id, items=[], total=0.0)
    cart_safe = _stringify_objectids(cart)
    return ShoppingCart(**cart_safe)

@app.post('/cart/checkout', response_model=CheckoutResult)
def checkout(current_user: str = Depends(get_current_user)):
    # Saga orchestrator: for each item validate and create token & purchase record, on error compensate
    user_id = current_user
    cart = cart_col.find_one({'user_id': user_id})
    if not cart or not cart.get('items'):
        raise HTTPException(status_code=400, detail='cart empty')
    created_tokens = []
    created_purchases = []
    failed = []
    try:
        for item in cart['items']:
            tour_id = item.get('tour_id')
            # check tour exists and is not archived
            tour = get_tour(tour_id)
            if not tour:
                failed.append({'item': item, 'reason': 'tour_not_found'})
                break
            status = tour.get('status') if isinstance(tour, dict) else tour.get('Status')
            if status == 'archived':
                failed.append({'item': item, 'reason': 'tour_archived'})
                break
            # create token
            token = str(uuid.uuid4())
            tok_doc = {'token': token, 'user_id': user_id, 'tour_id': tour_id, 'created_at': datetime.utcnow()}
            tokens_col.insert_one(tok_doc)
            created_tokens.append(tok_doc)
            # create purchase record
            purchase_doc = {'user_id': user_id, 'tour_id': tour_id, 'token': token, 'created_at': datetime.utcnow()}
            purchases_col.insert_one(purchase_doc)
            created_purchases.append(purchase_doc)
        if failed:
            # compensation: remove created tokens and purchases
            for t in created_tokens:
                tokens_col.delete_one({'token': t['token']})
            for p in created_purchases:
                purchases_col.delete_one({'token': p['token']})
            failed_safe = _stringify_objectids(failed)
            return CheckoutResult(user_id=user_id, purchased=[], failed=failed_safe)
        # success: clear cart
        cart_col.delete_one({'user_id': user_id})
        purchased_safe = _stringify_objectids(created_purchases)
        return CheckoutResult(user_id=user_id, purchased=purchased_safe, failed=[])
    except Exception as e:
        # compensation
        for t in created_tokens:
            tokens_col.delete_one({'token': t['token']})
        for p in created_purchases:
            purchases_col.delete_one({'token': p['token']})
        raise HTTPException(status_code=500, detail=str(e))

@app.get('/tokens', response_model=List[TourPurchaseToken])
def get_tokens(current_user: str = Depends(get_current_user)):
    user_id = current_user
    docs = list(tokens_col.find({'user_id': user_id}))
    docs_safe = _stringify_objectids(docs)
    out = []
    for d in docs_safe:
        out.append(TourPurchaseToken(token=d['token'], user_id=d['user_id'], tour_id=d['tour_id'], created_at=d['created_at']))
    return out
