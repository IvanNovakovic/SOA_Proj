# Database Seeding Scripts

This directory contains scripts to wipe and populate the databases with test data.

## Overview

The seeding script creates:
- **8 Users** with different roles (guides, tourists, admin)
- **15 Follow relationships** to demonstrate social features
- **7 Tours** created by various guides
- **6 Blog posts** with comments and likes

## Prerequisites

- Docker containers must be running (`docker-compose up -d`)
- Python 3.7+ installed
- MongoDB accessible on `localhost:27017`
- Neo4j accessible on `localhost:7687`

## Usage

### Windows
```bash
cd scripts
seed.bat
```

### Linux/Mac
```bash
cd scripts
chmod +x seed.sh
./seed.sh
```

### Direct Python
```bash
# Install dependencies
pip install -r scripts/requirements.txt

# Run seeding script
python scripts/seed_all_databases.py
```

## Test Users

All users have the password: `password123`

### Guides (can create tours)
- **john_guide** - Experienced tour guide specializing in historical tours
- **sarah_explorer** - Adventure guide and nature enthusiast
- **emma_guide** - Cultural tour specialist with 10 years experience
- **lisa_guide** - Coastal and marine tour expert

### Tourists
- **mike_tourist** - Travel blogger and photography lover (follows 3 guides)
- **alex_wanderer** - Solo traveler exploring the world (follows 2 guides)
- **david_blogger** - Food and culture blogger (follows 3 guides)

### Admin
- **anna_admin** - Platform administrator

## Sample Data Created

### Tours
1. Historic Downtown Walking Tour (john_guide) - Easy, $25
2. Mountain Peak Adventure (sarah_explorer) - Hard, $75
3. Riverside Cycling Route (emma_guide) - Medium, $35
4. Museum District Tour (john_guide) - Easy, $50
5. Coastal Sunset Walk (lisa_guide) - Easy, $30
6. Forest Trail Discovery (sarah_explorer) - Medium, $40
7. Food Market Experience (emma_guide) - Easy, $45

### Follow Network
- Tourists follow guides they're interested in
- Guides follow each other for collaboration
- Creates a realistic social graph for recommendations

### Blog Posts
- Mix of travel experiences, tips, and reflections
- Each post has 2-4 comments from different users
- Each post has 3-6 likes

## What Gets Wiped

The script will completely wipe:
- All MongoDB databases: `stakeholders`, `blogs`, `tours`
- All Neo4j nodes and relationships

**⚠️ Warning**: This will delete ALL existing data in these databases.

## Troubleshooting

### Services not responding
Make sure all containers are running:
```bash
docker-compose ps
docker-compose logs [service-name]
```

### Connection errors
Check that ports are accessible:
- MongoDB: `localhost:27017`
- Neo4j: `localhost:7687`
- Services: Check `docker-compose.yml` for port mappings

### Python dependencies issues
```bash
# Create fresh virtual environment
python -m venv venv
source venv/bin/activate  # or venv\Scripts\activate on Windows
pip install -r scripts/requirements.txt
```

## Customization

Edit `seed_all_databases.py` to modify:
- `USERS` - Add/remove users
- `FOLLOW_RELATIONSHIPS` - Change who follows whom
- `TOURS` - Add/remove tours
- `BLOG_POSTS` - Add/remove blog content

## Script Output

The script provides colored output showing:
- ✓ Green checkmarks for successful operations
- ✗ Red X marks for failures
- ➜ Cyan arrows for progress steps
- Detailed information about created entities

## Testing the Data

After seeding, you can:

1. **Login as a user**:
   ```bash
   curl -X POST http://localhost:8084/api/users/login \
     -H "Content-Type: application/json" \
     -d '{"username":"john_guide","password":"password123"}'
   ```

2. **View tours**:
   ```bash
   curl http://localhost:8083/tours
   ```

3. **View blog posts**:
   ```bash
   curl http://localhost:8081/blogs
   ```

4. **Check follow relationships**:
   ```bash
   # Get token from login, then:
   curl http://localhost:8080/followers/USER_ID \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

## Notes

- All operations are idempotent - you can run the script multiple times
- The script waits for services to be ready before starting
- Failed operations are logged but won't stop the script
- Some randomization is used for comments and likes
