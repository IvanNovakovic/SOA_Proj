#!/usr/bin/env bash
# Seed Neo4j with sample users and follow relationships for development/demo
# Usage: ./scripts/seed_neo4j.sh [NEO_USER] [NEO_PASS] [CONTAINER_NAME]

set -euo pipefail

NEO_USER=${1:-neo4j}
NEO_PASS=${2:-test}
CONTAINER=${3:-neo4j}

echo "Seeding Neo4j in container '$CONTAINER' with user '$NEO_USER'..."

CY1="CREATE (u1:User {id:'user1', name:'Alice'});"
CY2="CREATE (u2:User {id:'user2', name:'Bob'});"
CY3="CREATE (u3:User {id:'user3', name:'Carol'});"
# follow relations: user1->user2, user2->user3, user1->user3
REL1="MATCH (a:User {id:'user1'}), (b:User {id:'user2'}) MERGE (a)-[:FOLLOWS]->(b);"
REL2="MATCH (a:User {id:'user2'}), (b:User {id:'user3'}) MERGE (a)-[:FOLLOWS]->(b);"
REL3="MATCH (a:User {id:'user1'}), (b:User {id:'user3'}) MERGE (a)-[:FOLLOWS]->(b);"

echo "Creating users..."
docker exec -i "$CONTAINER" cypher-shell -u "$NEO_USER" -p "$NEO_PASS" -f - <<'CYPHER'
$CY1
$CY2
$CY3
CYPHER

echo "Creating follow relationships..."
docker exec -i "$CONTAINER" cypher-shell -u "$NEO_USER" -p "$NEO_PASS" -f - <<'CYPHER'
$REL1
$REL2
$REL3
CYPHER

echo "Seed complete. Example follows: user1 follows user2 and user3; user2 follows user3."
