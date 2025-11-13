#!/bin/bash
# Force Clean and Rebuild Script

echo "============================================"
echo "🔧 FORCE CLEAN & REBUILD"
echo "============================================"

# 1. Kill ALL Go processes
echo ""
echo "1️⃣  Killing ALL Go processes..."
pkill -9 -f "go" 2>/dev/null || true
pkill -9 -f "main" 2>/dev/null || true
pkill -9 -f "accounting" 2>/dev/null || true
sleep 2

# 2. Clean Go cache
echo ""
echo "2️⃣  Cleaning Go build cache..."
go clean -cache
go clean -modcache
go clean -testcache

# 3. Remove binary files
echo ""
echo "3️⃣  Removing old binaries..."
rm -f main 2>/dev/null || true
rm -f *.exe 2>/dev/null || true
rm -rf tmp 2>/dev/null || true

# 4. Force git reset to remote
echo ""
echo "4️⃣  Force resetting to remote version..."
git fetch origin main
git reset --hard origin/main

# 5. Verify critical file
echo ""
echo "5️⃣  Verifying critical file update..."
if grep -q "absBalance := netBalance.Abs()" services/unified_period_closing_service.go; then
    echo "   ✅ Code is updated correctly"
else
    echo "   ❌ Code is NOT updated!"
    echo "   Manually checking file..."
    head -n 150 services/unified_period_closing_service.go | tail -n 20
fi

# 6. Re-download dependencies
echo ""
echo "6️⃣  Re-downloading dependencies..."
go mod download
go mod tidy

# 7. Fix database
echo ""
echo "7️⃣  Running database fix..."
go run cmd/fix_period_closing_comprehensive.go

# 8. Build fresh
echo ""
echo "8️⃣  Building fresh binary..."
go build -a -o main main.go

echo ""
echo "============================================"
echo "✅ Clean rebuild completed!"
echo "============================================"
echo ""
echo "Now run: ./main"
echo "Or: go run main.go"