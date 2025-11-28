@echo off
setlocal EnableDelayedExpansion

echo ========================================
echo   Database Seeding Script
echo ========================================
echo.

REM Check if Python is installed
python --version >nul 2>&1
if errorlevel 1 (
    echo Error: Python is not installed
    exit /b 1
)

REM Check if services are running
echo Checking if Docker services are running...
docker ps | findstr "stakeholders-service" >nul
if errorlevel 1 (
    echo Error: Services are not running. Please start with 'docker-compose up -d'
    exit /b 1
)
echo OK - Services are running
echo.

REM Create virtual environment if it doesn't exist
if not exist "venv" (
    echo Creating virtual environment...
    python -m venv venv
    echo OK - Virtual environment created
    echo.
)

REM Activate virtual environment and install dependencies
call venv\Scripts\activate.bat

echo Installing dependencies...
pip install -q -r scripts\requirements.txt
echo OK - Dependencies installed
echo.

REM Run the seeding script
echo Starting database seeding...
echo.
python scripts\seed_all_databases.py

REM Deactivate virtual environment
call deactivate 2>nul

echo.
echo Done!
pause
