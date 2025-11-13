#!/bin/bash
# Restart backend with latest code from GitHub

echo "============================================"
echo "RESTART BACKEND WITH LATEST CODE"
echo "============================================"

# 1. Stop backend if running
echo ""
echo "1️⃣  Stopping backend server..."
pkill -f "go run main.go" 2>/dev/null || pkill -f "main" 2>/dev/null || echo "   Backend not running"
sleep 2

# 2. Pull latest code
echo ""
echo "2️⃣  Pulling latest code from GitHub..."
cd "$(dirname "$0")/.."
git pull origin main

if [ $? -ne 0 ]; then
    echo "   ❌ Failed to pull latest code"
    echo "   Please check your git configuration"
    exit 1
fi

# 3. Run database fix script
echo ""
echo "3️⃣  Running database cleanup/fix..."
cd backend
go run cmd/verify_and_fix_pc.go

# 4. Start backend
echo ""
echo "4️⃣  Starting backend server..."
echo "   Press Ctrl+C to stop"
echo ""
go run main.go
