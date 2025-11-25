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
saga_state_col = purchase_db['saga_states']  # Track saga execution state

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
    user_id = current_user
    saga_id = str(uuid.uuid4())
    saga_start_time = datetime.utcnow()
    
    # Initialize saga state
    saga_state = {
        'saga_id': saga_id,
        'user_id': user_id,
        'status': 'STARTED',
        'current_step': 'INIT',
        'started_at': saga_start_time,
        'updated_at': saga_start_time,
        'steps_completed': [],
        'created_tokens': [],
        'created_purchases': [],
        'payment_amount': 0.0,
        'payment_processed': False,
        'error': None
    }
    saga_state_col.insert_one(saga_state)
    logging.info("SAGA[%s] Started checkout for user %s", saga_id, user_id)
    
    try:
        # Execute SAGA steps sequentially
        cart = validate_cart(saga_id, user_id)
        validate_no_duplicates(saga_id, user_id, cart)
        validate_tours_available(saga_id, cart)
        process_payment_step(saga_id, cart)
        created_purchases = create_tokens_and_purchases(saga_id, user_id, cart)
        clear_cart_step(saga_id, user_id)
        
        # Mark saga as completed
        saga_state_col.update_one(
            {'saga_id': saga_id},
            {'$set': {
                'status': 'COMPLETED',
                'current_step': 'DONE',
                'completed_at': datetime.utcnow(),
                'updated_at': datetime.utcnow()
            }}
        )
        logging.info("SAGA[%s] COMPLETED: Checkout successful for user %s", saga_id, user_id)
        
        purchased_safe = _stringify_objectids(created_purchases)
        return CheckoutResult(user_id=user_id, purchased=purchased_safe, failed=[])
        
    except HTTPException:
        # Business validation errors - no compensation needed
        saga_state_col.update_one(
            {'saga_id': saga_id},
            {'$set': {'status': 'FAILED', 'updated_at': datetime.utcnow()}}
        )
        raise
        
    except Exception as e:
        # Critical failure - execute compensation
        saga_state_col.update_one(
            {'saga_id': saga_id},
            {'$set': {'status': 'COMPENSATING', 'error': str(e), 'updated_at': datetime.utcnow()}}
        )
        compensate_saga(saga_id, user_id)
        raise HTTPException(
            status_code=500, 
            detail='A critical error occurred while finalizing the purchase. Your payment will be refunded.'
        )


def update_saga_step(saga_id: str, step_name: str, step_status: str, error: str = None, metadata: dict = None):
    """Update saga state with step execution details."""
    update_data = {
        'current_step': step_name,
        'updated_at': datetime.utcnow()
    }
    
    step_record = {
        'step': step_name,
        'status': step_status,
        'timestamp': datetime.utcnow()
    }
    
    if error:
        step_record['error'] = error
        update_data['error'] = error
    
    if metadata:
        step_record['metadata'] = metadata
    
    saga_state_col.update_one(
        {'saga_id': saga_id},
        {
            '$set': update_data,
            '$push': {'steps_completed': step_record}
        }
    )


def validate_cart(saga_id: str, user_id: str):
    """SAGA Step 1: Fetch and validate cart state."""
    update_saga_step(saga_id, 'FETCH_CART', 'STARTED')
    cart = cart_col.find_one({'user_id': user_id})
    if not cart or not cart.get('items'):
        update_saga_step(saga_id, 'FETCH_CART', 'FAILED', error='cart empty')
        logging.warning("SAGA[%s] Step 1 FAILED: Cart empty for user %s", saga_id, user_id)
        raise HTTPException(status_code=400, detail='cart empty')
    update_saga_step(saga_id, 'FETCH_CART', 'COMPLETED', metadata={'item_count': len(cart['items'])})
    logging.info("SAGA[%s] Step 1 COMPLETED: Fetched cart with %d items", saga_id, len(cart['items']))
    return cart


def validate_no_duplicates(saga_id: str, user_id: str, cart: dict):
    """SAGA Step 2: Validate user doesn't already own tours in cart."""
    update_saga_step(saga_id, 'VALIDATE_DUPLICATES', 'STARTED')
    tour_ids_in_cart = [item.get('tour_id') for item in cart['items']]
    existing_tokens = list(tokens_col.find({'user_id': user_id, 'tour_id': {'$in': tour_ids_in_cart}}))
    if existing_tokens:
        owned_tour_id = existing_tokens[0]['tour_id']
        tour_name = next((item.get('name', 'Unknown') for item in cart['items'] if item.get('tour_id') == owned_tour_id), 'Unknown')
        update_saga_step(saga_id, 'VALIDATE_DUPLICATES', 'FAILED', error=f'duplicate purchase: {tour_name}')
        logging.warning("SAGA[%s] Step 2 FAILED: User already owns tour %s", saga_id, tour_name)
        raise HTTPException(
            status_code=400, 
            detail=f"You have already purchased the tour: '{tour_name}'."
        )
    update_saga_step(saga_id, 'VALIDATE_DUPLICATES', 'COMPLETED')
    logging.info("SAGA[%s] Step 2 COMPLETED: No duplicate purchases found", saga_id)


def validate_tours_available(saga_id: str, cart: dict):
    """SAGA Step 3: Validate all tours are published and available."""
    update_saga_step(saga_id, 'VALIDATE_TOURS', 'STARTED')
    for item in cart['items']:
        tour_id = item.get('tour_id')
        tour = get_tour(tour_id)
        if not tour:
            update_saga_step(saga_id, 'VALIDATE_TOURS', 'FAILED', error=f'tour not found: {tour_id}')
            logging.warning("SAGA[%s] Step 3 FAILED: Tour %s not found", saga_id, tour_id)
            raise HTTPException(status_code=400, detail=f"Tour '{item.get('name', tour_id)}' not found.")
        status = tour.get('status') if isinstance(tour, dict) else tour.get('Status')
        if status == 'archived' or (status and status.lower() not in ['published', 'active']):
            update_saga_step(saga_id, 'VALIDATE_TOURS', 'FAILED', error=f'tour unavailable: {tour_id} (status: {status})')
            logging.warning("SAGA[%s] Step 3 FAILED: Tour %s is unavailable (status: %s)", saga_id, tour_id, status)
            raise HTTPException(
                status_code=400, 
                detail=f"Tour '{item.get('name', tour_id)}' is no longer available for purchase."
            )
    update_saga_step(saga_id, 'VALIDATE_TOURS', 'COMPLETED', metadata={'tours_validated': len(cart['items'])})
    logging.info("SAGA[%s] Step 3 COMPLETED: All %d tours validated", saga_id, len(cart['items']))


def process_payment_step(saga_id: str, cart: dict) -> float:
    """SAGA Step 4: Process payment and return total amount."""
    update_saga_step(saga_id, 'PROCESS_PAYMENT', 'STARTED')
    total_amount = sum(float(item.get('price', 0)) for item in cart['items'])
    saga_state_col.update_one(
        {'saga_id': saga_id},
        {'$set': {'payment_amount': total_amount}}
    )
    payment_successful = simulate_payment(total_amount)
    if not payment_successful:
        update_saga_step(saga_id, 'PROCESS_PAYMENT', 'FAILED', error='payment processing failed')
        logging.error("SAGA[%s] Step 4 FAILED: Payment of %.2f failed", saga_id, total_amount)
        raise HTTPException(status_code=402, detail='Payment processing failed.')
    saga_state_col.update_one(
        {'saga_id': saga_id},
        {'$set': {'payment_processed': True}}
    )
    update_saga_step(saga_id, 'PROCESS_PAYMENT', 'COMPLETED', metadata={'amount': total_amount})
    logging.info("SAGA[%s] Step 4 COMPLETED: Payment of %.2f processed", saga_id, total_amount)
    return total_amount


def create_tokens_and_purchases(saga_id: str, user_id: str, cart: dict):
    """SAGA Step 5: Create purchase tokens and records."""
    update_saga_step(saga_id, 'CREATE_TOKENS_PURCHASES', 'STARTED')
    created_tokens = []
    created_purchases = []
    
    for item in cart['items']:
        tour_id = item.get('tour_id')
        token = str(uuid.uuid4())
        tok_doc = {'token': token, 'user_id': user_id, 'tour_id': tour_id, 'created_at': datetime.utcnow()}
        tokens_col.insert_one(tok_doc)
        created_tokens.append(tok_doc)
        logging.debug("SAGA[%s] Created token %s for tour %s", saga_id, token, tour_id)
        
        purchase_doc = {'user_id': user_id, 'tour_id': tour_id, 'token': token, 'created_at': datetime.utcnow()}
        purchases_col.insert_one(purchase_doc)
        created_purchases.append(purchase_doc)
        logging.debug("SAGA[%s] Created purchase record for tour %s", saga_id, tour_id)
    
    # Save created resources in saga state for potential compensation
    saga_state_col.update_one(
        {'saga_id': saga_id},
        {'$set': {
            'created_tokens': [t['token'] for t in created_tokens],
            'created_purchases': [{'tour_id': p['tour_id'], 'token': p['token']} for p in created_purchases]
        }}
    )
    update_saga_step(saga_id, 'CREATE_TOKENS_PURCHASES', 'COMPLETED', metadata={'tokens_created': len(created_tokens)})
    logging.info("SAGA[%s] Step 5 COMPLETED: Created %d tokens and purchases", saga_id, len(created_tokens))
    return created_purchases


def clear_cart_step(saga_id: str, user_id: str):
    """SAGA Step 6: Clear user's cart after successful purchase."""
    update_saga_step(saga_id, 'CLEAR_CART', 'STARTED')
    cart_col.delete_one({'user_id': user_id})
    update_saga_step(saga_id, 'CLEAR_CART', 'COMPLETED')
    logging.info("SAGA[%s] Step 6 COMPLETED: Cart cleared for user %s", saga_id, user_id)


def compensate_saga(saga_id: str, user_id: str):
    """Execute compensation (rollback) for failed saga."""
    logging.error("SAGA[%s] CRITICAL FAILURE - starting compensation", saga_id)
    saga_state_col.update_one(
        {'saga_id': saga_id},
        {'$set': {
            'status': 'COMPENSATING',
            'current_step': 'ROLLBACK',
            'updated_at': datetime.utcnow()
        }}
    )
    
    # Retrieve saga state for compensation
    saga = saga_state_col.find_one({'saga_id': saga_id})
    created_token_ids = saga.get('created_tokens', [])
    created_purchase_data = saga.get('created_purchases', [])
    payment_processed = saga.get('payment_processed', False)
    total_amount = saga.get('payment_amount', 0.0)
    
    # Rollback tokens
    logging.info("SAGA[%s] COMPENSATION: Rolling back %d tokens", saga_id, len(created_token_ids))
    for token_id in created_token_ids:
        try:
            tokens_col.delete_one({'token': token_id})
            logging.debug("SAGA[%s] Deleted token %s", saga_id, token_id)
        except Exception as del_err:
            logging.error("SAGA[%s] Failed to delete token %s: %s", saga_id, token_id, del_err)
    
    # Rollback purchases
    logging.info("SAGA[%s] COMPENSATION: Rolling back %d purchases", saga_id, len(created_purchase_data))
    for purchase_info in created_purchase_data:
        try:
            purchases_col.delete_one({'token': purchase_info['token']})
            logging.debug("SAGA[%s] Deleted purchase for tour %s", saga_id, purchase_info.get('tour_id'))
        except Exception as del_err:
            logging.error("SAGA[%s] Failed to delete purchase: %s", saga_id, del_err)
    
    # Refund payment if it was processed
    if payment_processed:
        logging.warning("SAGA[%s] COMPENSATION: Refunding payment of %.2f", saga_id, total_amount)
        refund_successful = refund_payment(user_id, total_amount)
        if not refund_successful:
            logging.critical("SAGA[%s] CRITICAL: Payment refund failed for user %s amount %.2f", saga_id, user_id, total_amount)
            saga_state_col.update_one(
                {'saga_id': saga_id},
                {'$set': {'status': 'COMPENSATION_FAILED', 'updated_at': datetime.utcnow()}}
            )
            return False
        logging.info("SAGA[%s] COMPENSATION: Refund successful", saga_id)
    
    # Mark saga as compensated
    saga_state_col.update_one(
        {'saga_id': saga_id},
        {'$set': {
            'status': 'COMPENSATED',
            'updated_at': datetime.utcnow(),
            'compensated_at': datetime.utcnow()
        }}
    )
    logging.info("SAGA[%s] COMPENSATED: Rollback complete", saga_id)
    return True


def simulate_payment(amount: float) -> bool:
    """Simulate payment processing. Replace with real payment gateway (Stripe, PayPal, etc.)."""
    logging.info("Simulating payment of %.2f", amount)
    # In production: call payment gateway API here
    # For now, always succeed
    return True


def refund_payment(user_id: str, amount: float) -> bool:
    """Simulate payment refund. Replace with real payment gateway refund API."""
    logging.warning("Simulating payment refund of %.2f for user %s", amount, user_id)
    # In production: call payment gateway refund API here
    # For now, always succeed
    return True

@app.get('/tokens', response_model=List[TourPurchaseToken])
def get_tokens(current_user: str = Depends(get_current_user)):
    user_id = current_user
    docs = list(tokens_col.find({'user_id': user_id}))
    docs_safe = _stringify_objectids(docs)
    out = []
    for d in docs_safe:
        out.append(TourPurchaseToken(token=d['token'], user_id=d['user_id'], tour_id=d['tour_id'], created_at=d['created_at']))
    return out
