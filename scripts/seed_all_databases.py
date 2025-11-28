#!/usr/bin/env python3
"""
Database Seeding Script
Wipes and populates all databases with test data including:
- Users with different roles
- Follow relationships (Neo4j)
- Blog posts with comments and likes
- Tours with key points
"""

import requests
import time
import sys
from datetime import datetime, timedelta
import random
from pymongo import MongoClient
from neo4j import GraphDatabase

# Service URLs
GATEWAY_URL = "http://localhost:8080"
STAKEHOLDERS_URL = "http://localhost:8084"
BLOG_URL = "http://localhost:8081"
TOUR_URL = "http://localhost:8083"

# Database connections
MONGO_URI = "mongodb://localhost:27017"
NEO4J_URI = "bolt://localhost:7687"
NEO4J_USER = "neo4j"
NEO4J_PASSWORD = "testtest123"

# Colors for output
class Colors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_step(message):
    print(f"{Colors.OKCYAN}➜ {message}{Colors.ENDC}")

def print_success(message):
    print(f"{Colors.OKGREEN}✓ {message}{Colors.ENDC}")

def print_error(message):
    print(f"{Colors.FAIL}✗ {message}{Colors.ENDC}")

def print_header(message):
    print(f"\n{Colors.HEADER}{Colors.BOLD}{'='*60}")
    print(f"  {message}")
    print(f"{'='*60}{Colors.ENDC}\n")

# Sample users data
USERS = [
    {
        "username": "john_guide",
        "password": "password123",
        "email": "john@example.com",
        "name": "John",
        "surname": "Guide",
        "roles": ["tourist", "guide"],
        "address": {"street": "123 Main St", "city": "New York", "country": "USA"},
        "biography": "Experienced tour guide specializing in historical tours",
        "motto": "Discover history, one step at a time"
    },
    {
        "username": "sarah_explorer",
        "password": "password123",
        "email": "sarah@example.com",
        "name": "Sarah",
        "surname": "Explorer",
        "roles": ["tourist", "guide"],
        "address": {"street": "456 Oak Ave", "city": "London", "country": "UK"},
        "biography": "Adventure guide and nature enthusiast",
        "motto": "Life is an adventure!"
    },
    {
        "username": "mike_tourist",
        "password": "password123",
        "email": "mike@example.com",
        "name": "Mike",
        "surname": "Tourist",
        "roles": ["tourist"],
        "address": {"street": "789 Pine Rd", "city": "Paris", "country": "France"},
        "biography": "Travel blogger and photography lover",
        "motto": "Capturing moments around the world"
    },
    {
        "username": "emma_guide",
        "password": "password123",
        "email": "emma@example.com",
        "name": "Emma",
        "surname": "Guide",
        "roles": ["tourist", "guide"],
        "address": {"street": "321 Elm St", "city": "Berlin", "country": "Germany"},
        "biography": "Cultural tour specialist with 10 years experience",
        "motto": "Every place has a story"
    },
    {
        "username": "alex_wanderer",
        "password": "password123",
        "email": "alex@example.com",
        "name": "Alex",
        "surname": "Wanderer",
        "roles": ["tourist"],
        "address": {"street": "654 Maple Dr", "city": "Tokyo", "country": "Japan"},
        "biography": "Solo traveler exploring the world",
        "motto": "Wander often, wonder always"
    },
    {
        "username": "lisa_guide",
        "password": "password123",
        "email": "lisa@example.com",
        "name": "Lisa",
        "surname": "Guide",
        "roles": ["tourist", "guide"],
        "address": {"street": "987 Cedar Ln", "city": "Sydney", "country": "Australia"},
        "biography": "Coastal and marine tour expert",
        "motto": "The ocean is calling"
    },
    {
        "username": "david_blogger",
        "password": "password123",
        "email": "david@example.com",
        "name": "David",
        "surname": "Blogger",
        "roles": ["tourist"],
        "address": {"street": "147 Birch St", "city": "Rome", "country": "Italy"},
        "biography": "Food and culture blogger",
        "motto": "Eat, travel, blog, repeat"
    },
    {
        "username": "anna_admin",
        "password": "password123",
        "email": "anna@example.com",
        "name": "Anna",
        "surname": "Admin",
        "roles": ["admin"],
        "address": {"street": "258 Spruce Ave", "city": "Toronto", "country": "Canada"},
        "biography": "Platform administrator",
        "motto": "Keeping the platform running smoothly"
    }
]

# Follow relationships (username pairs)
FOLLOW_RELATIONSHIPS = [
    ("mike_tourist", "john_guide"),
    ("mike_tourist", "sarah_explorer"),
    ("mike_tourist", "emma_guide"),
    ("alex_wanderer", "john_guide"),
    ("alex_wanderer", "lisa_guide"),
    ("david_blogger", "sarah_explorer"),
    ("david_blogger", "emma_guide"),
    ("david_blogger", "lisa_guide"),
    ("john_guide", "sarah_explorer"),
    ("john_guide", "emma_guide"),
    ("sarah_explorer", "lisa_guide"),
    ("emma_guide", "john_guide"),
    ("lisa_guide", "sarah_explorer"),
    ("alex_wanderer", "mike_tourist"),
    ("mike_tourist", "alex_wanderer"),
]

# Sample tours
TOURS = [
    {
        "author": "john_guide",
        "name": "Historic Downtown Walking Tour",
        "description": "Explore the rich history of downtown through iconic landmarks and hidden gems. Perfect for history enthusiasts!",
        "difficulty": "easy",
        "tags": ["history", "culture", "walking", "downtown"],
        "price": 25.00,
        "distance": 3.5,
        "durations": {"walking": 120, "biking": 45, "driving": 20}
    },
    {
        "author": "sarah_explorer",
        "name": "Mountain Peak Adventure",
        "description": "Challenging hike to the summit with breathtaking panoramic views. Suitable for experienced hikers.",
        "difficulty": "hard",
        "tags": ["hiking", "mountain", "adventure", "nature"],
        "price": 75.00,
        "distance": 12.5,
        "durations": {"walking": 360, "biking": 180, "driving": 45}
    },
    {
        "author": "emma_guide",
        "name": "Riverside Cycling Route",
        "description": "Scenic bike ride along the peaceful riverside path. Great for families and casual cyclists.",
        "difficulty": "medium",
        "tags": ["cycling", "river", "scenic", "family-friendly"],
        "price": 35.00,
        "distance": 15.0,
        "durations": {"walking": 240, "biking": 90, "driving": 30}
    },
    {
        "author": "john_guide",
        "name": "Museum District Tour",
        "description": "Visit five world-class museums in one afternoon. Includes skip-the-line access and expert commentary.",
        "difficulty": "easy",
        "tags": ["museums", "art", "culture", "indoor"],
        "price": 50.00,
        "distance": 2.0,
        "durations": {"walking": 180, "biking": 30, "driving": 15}
    },
    {
        "author": "lisa_guide",
        "name": "Coastal Sunset Walk",
        "description": "Relaxing evening stroll along the beautiful coastline. Watch the sun set over the ocean.",
        "difficulty": "easy",
        "tags": ["beach", "sunset", "coastal", "photography"],
        "price": 30.00,
        "distance": 4.0,
        "durations": {"walking": 90, "biking": 30, "driving": 15}
    },
    {
        "author": "sarah_explorer",
        "name": "Forest Trail Discovery",
        "description": "Immerse yourself in nature on this moderate forest trail. Learn about local flora and fauna.",
        "difficulty": "medium",
        "tags": ["forest", "nature", "wildlife", "education"],
        "price": 40.00,
        "distance": 8.0,
        "durations": {"walking": 180, "biking": 75, "driving": 25}
    },
    {
        "author": "emma_guide",
        "name": "Food Market Experience",
        "description": "Culinary journey through the city's best food markets. Taste local delicacies and meet vendors.",
        "difficulty": "easy",
        "tags": ["food", "market", "culinary", "local"],
        "price": 45.00,
        "distance": 2.5,
        "durations": {"walking": 150, "biking": 35, "driving": 20}
    }
]

# Sample blog posts
BLOG_POSTS = [
    {
        "author": "mike_tourist",
        "title": "My First Mountain Hike Experience",
        "description": "After months of preparation, I finally conquered my first mountain peak! The journey was challenging but incredibly rewarding. The views from the top were absolutely breathtaking. I learned so much about perseverance and pushing my limits. Can't wait for the next adventure!",
        "images": []
    },
    {
        "author": "david_blogger",
        "title": "Top 10 Hidden Food Gems in the City",
        "description": "As a food enthusiast, I've spent years exploring local eateries. Here are my favorite hidden gems that most tourists miss. From authentic street food to family-owned restaurants, these places offer unforgettable culinary experiences. Don't leave the city without trying them!",
        "images": []
    },
    {
        "author": "sarah_explorer",
        "title": "The Importance of Sustainable Tourism",
        "description": "As tour guides, we have a responsibility to protect the places we love. This post discusses practical ways to make your tours more sustainable and environmentally friendly. Small changes can make a big difference for future generations.",
        "images": []
    },
    {
        "author": "mike_tourist",
        "title": "Photography Tips for Travel Bloggers",
        "description": "Over the years, I've learned some valuable lessons about travel photography. From lighting to composition, here are my top tips for capturing stunning travel moments. These techniques have helped me grow my blog significantly.",
        "images": []
    },
    {
        "author": "john_guide",
        "title": "Why History Tours Matter in Modern Times",
        "description": "In our fast-paced digital world, taking time to understand history is more important than ever. Historical tours help us connect with our roots and learn from the past. This reflection explores why I'm passionate about sharing history with others.",
        "images": []
    },
    {
        "author": "david_blogger",
        "title": "Solo Travel: My Journey to Self-Discovery",
        "description": "Traveling alone was terrifying at first, but it became the most transformative experience of my life. This post shares my personal journey and tips for anyone considering solo travel. You'll discover more about yourself than you ever imagined.",
        "images": []
    }
]

# Global storage for created entities
created_users = {}
created_tours = {}
user_tokens = {}

def wait_for_services():
    """Wait for all services to be ready"""
    print_step("Waiting for services to be ready...")
    
    services = {
        "Stakeholders": f"{STAKEHOLDERS_URL}/health",
        "Blog": f"{BLOG_URL}/health",
        "Tour": f"{TOUR_URL}/health",
    }
    
    max_retries = 30
    for name, url in services.items():
        for i in range(max_retries):
            try:
                response = requests.get(url, timeout=2)
                if response.status_code == 200:
                    print_success(f"{name} service is ready")
                    break
            except:
                if i == max_retries - 1:
                    print_error(f"{name} service is not responding")
                    return False
                time.sleep(2)
    
    return True

def wipe_databases():
    """Wipe all databases"""
    print_header("WIPING DATABASES")
    
    try:
        # Wipe MongoDB collections
        print_step("Connecting to MongoDB...")
        mongo_client = MongoClient(MONGO_URI)
        
        # Drop stakeholders database
        print_step("Dropping stakeholders database...")
        mongo_client.drop_database("stakeholders")
        print_success("Stakeholders database dropped")
        
        # Drop blogs database
        print_step("Dropping blogs database...")
        mongo_client.drop_database("blogs")
        print_success("Blogs database dropped")
        
        # Drop tours database
        print_step("Dropping tours database...")
        mongo_client.drop_database("tours")
        print_success("Tours database dropped")
        
        mongo_client.close()
        
        # Wipe Neo4j
        print_step("Connecting to Neo4j...")
        neo4j_driver = GraphDatabase.driver(NEO4J_URI, auth=(NEO4J_USER, NEO4J_PASSWORD))
        
        with neo4j_driver.session() as session:
            print_step("Deleting all Neo4j nodes and relationships...")
            session.run("MATCH (n) DETACH DELETE n")
            print_success("Neo4j database cleared")
        
        neo4j_driver.close()
        
        print_success("All databases wiped successfully")
        time.sleep(2)
        
    except Exception as e:
        print_error(f"Error wiping databases: {e}")
        print_error("Make sure MongoDB and Neo4j are accessible")
        return False
    
    return True

def register_user(user_data):
    """Register a single user"""
    try:
        response = requests.post(
            f"{STAKEHOLDERS_URL}/api/users/register",
            json=user_data,
            timeout=5
        )
        
        if response.status_code in [200, 201]:
            data = response.json()
            user_id = data.get("id") or data.get("_id")
            print_success(f"Created user: {user_data['username']} (ID: {user_id})")
            return user_id
        else:
            print_error(f"Failed to create user {user_data['username']}: {response.text}")
            return None
    except Exception as e:
        print_error(f"Error creating user {user_data['username']}: {e}")
        return None

def login_user(username, password):
    """Login and get JWT token"""
    try:
        response = requests.post(
            f"{STAKEHOLDERS_URL}/api/users/login",
            json={"username": username, "password": password},
            timeout=5
        )
        
        if response.status_code == 200:
            data = response.json()
            token = data.get("token")
            user_id = data.get("userId")
            if token:
                print_success(f"Logged in: {username}")
                return token, user_id
        
        print_error(f"Failed to login {username}")
        return None, None
    except Exception as e:
        print_error(f"Error logging in {username}: {e}")
        return None, None

def create_follow_relationship(follower_token, follower_id, followee_id):
    """Create a follow relationship"""
    try:
        response = requests.post(
            f"{GATEWAY_URL}/follow",
            json={"followerId": follower_id, "followeeId": followee_id},
            headers={"Authorization": f"Bearer {follower_token}"},
            timeout=5
        )
        
        if response.status_code in [200, 201]:
            return True
        else:
            print_error(f"Failed to create follow: {response.text}")
            return False
    except Exception as e:
        print_error(f"Error creating follow: {e}")
        return False

def create_tour(author_token, author_id, tour_data):
    """Create a tour"""
    try:
        tour_payload = {
            "authorId": author_id,
            "name": tour_data["name"],
            "description": tour_data["description"],
            "difficulty": tour_data["difficulty"],
            "tags": tour_data["tags"],
            "status": "published",
            "price": tour_data["price"],
            "distance": tour_data["distance"],
            "durations": tour_data["durations"]
        }
        
        response = requests.post(
            f"{TOUR_URL}/tours",
            json=tour_payload,
            headers={"Authorization": f"Bearer {author_token}"},
            timeout=5
        )
        
        if response.status_code in [200, 201]:
            data = response.json()
            tour_id = data.get("id") or data.get("_id")
            print_success(f"Created tour: {tour_data['name']} (ID: {tour_id})")
            return tour_id
        else:
            print_error(f"Failed to create tour {tour_data['name']}: {response.text}")
            return None
    except Exception as e:
        print_error(f"Error creating tour {tour_data['name']}: {e}")
        return None

def create_blog_post(author_token, author_id, author_name, blog_data):
    """Create a blog post"""
    try:
        blog_payload = {
            "title": blog_data["title"],
            "description": blog_data["description"],
            "author_id": author_id,
            "author_name": author_name,
            "images": blog_data.get("images", [])
        }
        
        response = requests.post(
            f"{BLOG_URL}/blogs",
            json=blog_payload,
            headers={"Authorization": f"Bearer {author_token}"},
            timeout=5
        )
        
        if response.status_code in [200, 201]:
            data = response.json()
            blog_id = data.get("id") or data.get("_id")
            print_success(f"Created blog: {blog_data['title'][:40]}... (ID: {blog_id})")
            return blog_id
        else:
            print_error(f"Failed to create blog: {response.text}")
            return None
    except Exception as e:
        print_error(f"Error creating blog: {e}")
        return None

def create_blog_comment(commenter_token, blog_id, comment_text):
    """Create a comment on a blog post"""
    try:
        response = requests.post(
            f"{BLOG_URL}/blogs/{blog_id}/comments",
            json={"text": comment_text},
            headers={"Authorization": f"Bearer {commenter_token}"},
            timeout=5
        )
        
        if response.status_code in [200, 201]:
            return True
        return False
    except Exception as e:
        return False

def like_blog_post(liker_token, blog_id):
    """Like a blog post"""
    try:
        response = requests.post(
            f"{BLOG_URL}/blogs/{blog_id}/like",
            headers={"Authorization": f"Bearer {liker_token}"},
            timeout=5
        )
        
        if response.status_code in [200, 201]:
            return True
        return False
    except Exception as e:
        return False

def seed_users():
    """Create all users and login"""
    print_header("CREATING USERS")
    
    for user_data in USERS:
        user_id = register_user(user_data)
        if user_id:
            created_users[user_data['username']] = {
                'id': user_id,
                'data': user_data
            }
            
            # Login to get token
            token, _ = login_user(user_data['username'], user_data['password'])
            if token:
                user_tokens[user_data['username']] = token
    
    time.sleep(2)
    print_success(f"Created {len(created_users)} users")

def seed_follows():
    """Create follow relationships"""
    print_header("CREATING FOLLOW RELATIONSHIPS")
    
    success_count = 0
    for follower_username, followee_username in FOLLOW_RELATIONSHIPS:
        if follower_username in created_users and followee_username in created_users:
            follower_token = user_tokens.get(follower_username)
            follower_id = created_users[follower_username]['id']
            followee_id = created_users[followee_username]['id']
            
            if follower_token and create_follow_relationship(follower_token, follower_id, followee_id):
                success_count += 1
                print_success(f"{follower_username} → {followee_username}")
    
    print_success(f"Created {success_count} follow relationships")

def seed_tours():
    """Create tours"""
    print_header("CREATING TOURS")
    
    for tour_data in TOURS:
        author_username = tour_data['author']
        if author_username in created_users:
            author_token = user_tokens.get(author_username)
            author_id = created_users[author_username]['id']
            
            if author_token:
                tour_id = create_tour(author_token, author_id, tour_data)
                if tour_id:
                    created_tours[tour_id] = {
                        'data': tour_data,
                        'author': author_username
                    }
    
    print_success(f"Created {len(created_tours)} tours")

def seed_blogs():
    """Create blog posts with comments and likes"""
    print_header("CREATING BLOG POSTS")
    
    created_blogs = []
    
    for blog_data in BLOG_POSTS:
        author_username = blog_data['author']
        if author_username in created_users:
            author_token = user_tokens.get(author_username)
            author_id = created_users[author_username]['id']
            author_name = created_users[author_username]['data']['name']
            
            if author_token:
                blog_id = create_blog_post(author_token, author_id, author_name, blog_data)
                if blog_id:
                    created_blogs.append(blog_id)
    
    print_success(f"Created {len(created_blogs)} blog posts")
    
    # Add some comments and likes
    print_step("Adding comments and likes...")
    
    sample_comments = [
        "Great post! Very informative.",
        "I loved reading this, thanks for sharing!",
        "This inspired me to try this myself.",
        "Excellent tips, keep them coming!",
        "I had a similar experience, totally agree!",
        "Wonderful content, looking forward to more."
    ]
    
    comment_count = 0
    like_count = 0
    
    for blog_id in created_blogs:
        # Random users comment (2-4 comments per blog)
        num_comments = random.randint(2, 4)
        commenters = random.sample(list(user_tokens.keys()), min(num_comments, len(user_tokens)))
        
        for commenter_username in commenters:
            token = user_tokens[commenter_username]
            comment = random.choice(sample_comments)
            if create_blog_comment(token, blog_id, comment):
                comment_count += 1
        
        # Random users like (3-6 likes per blog)
        num_likes = random.randint(3, 6)
        likers = random.sample(list(user_tokens.keys()), min(num_likes, len(user_tokens)))
        
        for liker_username in likers:
            token = user_tokens[liker_username]
            if like_blog_post(token, blog_id):
                like_count += 1
    
    print_success(f"Added {comment_count} comments and {like_count} likes")

def main():
    print_header("DATABASE SEEDING SCRIPT")
    print("This script will wipe and populate all databases with test data\n")
    
    # Wait for services
    if not wait_for_services():
        print_error("Services are not ready. Please ensure all containers are running.")
        sys.exit(1)
    
    # Wipe databases (manual step noted)
    wipe_databases()
    
    # Seed data
    seed_users()
    seed_follows()
    seed_tours()
    seed_blogs()
    
    # Summary
    print_header("SEEDING COMPLETE")
    print_success(f"✓ {len(created_users)} users created")
    print_success(f"✓ {len(FOLLOW_RELATIONSHIPS)} follow relationships created")
    print_success(f"✓ {len(created_tours)} tours created")
    print_success(f"✓ {len(BLOG_POSTS)} blog posts created (with comments and likes)")
    
    print(f"\n{Colors.OKBLUE}{'='*60}{Colors.ENDC}")
    print(f"{Colors.OKBLUE}Sample login credentials:{Colors.ENDC}")
    print(f"{Colors.OKGREEN}  Username: john_guide{Colors.ENDC}")
    print(f"{Colors.OKGREEN}  Password: password123{Colors.ENDC}")
    print(f"\n{Colors.OKCYAN}Other test users:{Colors.ENDC}")
    print(f"  - sarah_explorer (guide)")
    print(f"  - emma_guide (guide)")
    print(f"  - lisa_guide (guide)")
    print(f"  - mike_tourist (tourist)")
    print(f"  - alex_wanderer (tourist)")
    print(f"  - david_blogger (tourist)")
    print(f"  - anna_admin (admin)")
    print(f"{Colors.OKBLUE}{'='*60}{Colors.ENDC}\n")

if __name__ == "__main__":
    main()
