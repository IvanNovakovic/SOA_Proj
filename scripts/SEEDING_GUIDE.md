# Database Seeding Quick Reference

## 🚀 Quick Start

```bash
# Make sure services are running
docker-compose up -d

# Run seeding (Windows)
scripts\seed.bat

# Run seeding (Linux/Mac)
./scripts/seed.sh
```

## 👥 Test Users (password: `password123`)

| Username | Role | Description |
|----------|------|-------------|
| john_guide | Guide | Historical tours specialist |
| sarah_explorer | Guide | Adventure & nature |
| emma_guide | Guide | Cultural tours (10yr exp) |
| lisa_guide | Guide | Coastal & marine tours |
| mike_tourist | Tourist | Travel blogger, follows 3 guides |
| alex_wanderer | Tourist | Solo traveler, follows 2 guides |
| david_blogger | Tourist | Food blogger, follows 3 guides |
| anna_admin | Admin | Platform administrator |

## 🗺️ Sample Tours Created

| Tour Name | Author | Difficulty | Price | Distance |
|-----------|--------|------------|-------|----------|
| Historic Downtown Walking Tour | john_guide | Easy | $25 | 3.5 km |
| Mountain Peak Adventure | sarah_explorer | Hard | $75 | 12.5 km |
| Riverside Cycling Route | emma_guide | Medium | $35 | 15 km |
| Museum District Tour | john_guide | Easy | $50 | 2 km |
| Coastal Sunset Walk | lisa_guide | Easy | $30 | 4 km |
| Forest Trail Discovery | sarah_explorer | Medium | $40 | 8 km |
| Food Market Experience | emma_guide | Easy | $45 | 2.5 km |

## 📝 Sample Blog Posts

| Title | Author | Engagement |
|-------|--------|------------|
| My First Mountain Hike Experience | mike_tourist | 2-4 comments, 3-6 likes |
| Top 10 Hidden Food Gems | david_blogger | 2-4 comments, 3-6 likes |
| Sustainable Tourism Importance | sarah_explorer | 2-4 comments, 3-6 likes |
| Photography Tips for Travel | mike_tourist | 2-4 comments, 3-6 likes |
| Why History Tours Matter | john_guide | 2-4 comments, 3-6 likes |
| Solo Travel: Self-Discovery | david_blogger | 2-4 comments, 3-6 likes |

## 🔗 Follow Network

**Tourists follow Guides:**
- mike_tourist → john_guide, sarah_explorer, emma_guide
- alex_wanderer → john_guide, lisa_guide
- david_blogger → sarah_explorer, emma_guide, lisa_guide

**Guides collaborate:**
- john_guide ↔ sarah_explorer, emma_guide
- sarah_explorer → lisa_guide
- emma_guide → john_guide
- lisa_guide → sarah_explorer

**Tourists connect:**
- alex_wanderer ↔ mike_tourist

## 🧪 Testing After Seeding

### 1. Login as a user
```bash
curl -X POST http://localhost:8084/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john_guide","password":"password123"}'
```

### 2. View all tours
```bash
curl http://localhost:8083/tours
```

### 3. Get tours by author
```bash
curl http://localhost:8080/tours/author/USER_ID
```

### 4. View blog posts
```bash
curl http://localhost:8081/blogs
```

### 5. Check followers (requires auth token)
```bash
curl http://localhost:8080/followers/USER_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 6. Get follow recommendations
The follow network enables the recommendation system to suggest:
- Guides that friends follow
- Popular guides in the network
- Guides with similar interests

## 📊 Data Summary

- **Users:** 8 (4 guides, 3 tourists, 1 admin)
- **Follow relationships:** 15
- **Tours:** 7 (across different guides)
- **Blog posts:** 6 (with 12-24 comments, 18-36 likes total)
- **Databases wiped:** MongoDB (stakeholders, blogs, tours) + Neo4j

## ⚠️ Important Notes

1. **All existing data will be deleted** when running the script
2. Services must be running before seeding
3. MongoDB must be accessible on `localhost:27017`
4. Neo4j must be accessible on `localhost:7687`
5. All users have the same password for testing: `password123`

## 🔧 Customization

Edit `scripts/seed_all_databases.py` to:
- Add/modify users in the `USERS` list
- Change follow relationships in `FOLLOW_RELATIONSHIPS`
- Add/modify tours in the `TOURS` list
- Add/modify blog posts in the `BLOG_POSTS` list

## 📚 Use Cases Demonstrated

✅ **User Registration & Authentication**
✅ **Social Following Network**
✅ **Follow Recommendations Algorithm**
✅ **Tour Creation by Guides**
✅ **Tour Discovery**
✅ **Blog Posts with Engagement**
✅ **Comments System**
✅ **Likes System**
✅ **User Roles (Guide, Tourist, Admin)**
✅ **Multi-database Architecture**

## 🐛 Troubleshooting

**"Services not ready"**
- Check: `docker-compose ps`
- Wait 30 seconds after starting containers

**"Connection refused"**
- MongoDB: Check port 27017 is accessible
- Neo4j: Check port 7687 is accessible
- Services: Check `docker-compose logs [service]`

**"Module not found"**
- Run: `pip install -r scripts/requirements.txt`

**Script hangs**
- Increase timeout in the script
- Check service logs for errors
- Verify services are healthy
