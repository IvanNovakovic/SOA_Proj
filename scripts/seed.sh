#!/bin/bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}  Database Seeding Script${NC}"
echo -e "${CYAN}========================================${NC}\n"

# Check if Python3 is installed
if ! command -v python &> /dev/null; then
    echo -e "${RED}Error: Python3 is not installed${NC}"
    exit 1
fi

# Check if services are running
echo -e "${YELLOW}Checking if Docker services are running...${NC}"
if ! docker ps | grep -q "stakeholders-service"; then
    echo -e "${RED}Error: Services are not running. Please start with 'docker-compose up -d'${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Services are running${NC}\n"

# Install Python dependencies if needed
if [ ! -d "venv" ]; then
    echo -e "${YELLOW}Creating virtual environment...${NC}"
    python3 -m venv venv
    echo -e "${GREEN}✓ Virtual environment created${NC}\n"
fi

# Activate virtual environment
source venv/bin/activate 2>/dev/null || source venv/Scripts/activate 2>/dev/null

# Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"
pip install -q -r requirements.txt
echo -e "${GREEN}✓ Dependencies installed${NC}\n"

# Run the seeding script
echo -e "${CYAN}Starting database seeding...${NC}\n"
python seed_all_databases.py

# Deactivate virtual environment
deactivate 2>/dev/null

echo -e "\n${GREEN}Done!${NC}"
