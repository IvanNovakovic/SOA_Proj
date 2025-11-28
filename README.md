# Tour Management Platform

Microservices platform for creating and experiencing guided tours with real-time GPS tracking.

## Services

- **Tour Service** (Go) - Tour and key point management
- **Stakeholders Service** (Go) - Authentication
- **Blog Service** (Go) - Blogs and comments
- **Follower Service** (Go) - Social graph (Neo4j)
- **Purchase Service** (Python) - Shopping cart
- **Gateway Service** (Go) - API gateway
- **Frontend Service** (Vue.js 3) - SPA with maps

**Stack:** Go | Python | Vue.js | MongoDB | Neo4j | Docker

## Setup

1. Clone repository:
```bash
git clone https://github.com/IvanNovakovic/SOA_Proj.git
cd SOA_Proj
```

2. Add API key to `frontend-service/.env`:
```
VITE_OPENROUTE_API_KEY=your_key
```

3. Run:
```bash
docker-compose up --build
```

4. Access:
- App: http://localhost:8087
- API: http://localhost:8080

## Database Seeding

Populate the databases with test data (users, follows, tours, blogs):

**Windows:**
```bash
cd scripts
seed.bat
```

**Linux/Mac:**
```bash
cd scripts
./seed.sh
```

This creates:
- 8 test users (guides, tourists, admin)
- 15 follow relationships
- 7 tours with various difficulties
- 6 blog posts with comments and likes

**Login credentials:** All users have password `password123`
- Guides: `john_guide`, `sarah_explorer`, `emma_guide`, `lisa_guide`
- Tourists: `mike_tourist`, `alex_wanderer`, `david_blogger`
- Admin: `anna_admin`

See [scripts/README.md](scripts/README.md) for details.

## Development

```bash
# Rebuild service
docker-compose up -d --build service-name

# View logs
docker logs service-name

# Stop
docker-compose down
```