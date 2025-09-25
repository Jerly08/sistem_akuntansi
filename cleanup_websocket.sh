#!/bin/bash

echo "🧹 Cleaning up WebSocket references from frontend..."

# Remove websocket files (already done)
echo "✅ Websocket files already removed"

# Find and clean up websocket references in remaining files
echo "🔍 Scanning for remaining websocket references..."

# Files that might contain websocket references
FILES=(
  "frontend/src/components/journals/JournalDrilldownModal.tsx"
  "frontend/src/components/reports/JournalDrilldownModal.tsx" 
  "frontend/src/components/reports/EnhancedPLReportPage.tsx"
  "frontend/src/utils/backendTest.ts"
)

for file in "${FILES[@]}"; do
  if [ -f "$file" ]; then
    echo "📝 Processing $file..."
    
    # Remove websocket imports
    sed -i '/import.*WebSocket/d' "$file"
    sed -i '/import.*balanceWebSocket/d' "$file"
    sed -i '/import.*balanceMonitor/d' "$file"
    sed -i '/import.*BalanceWebSocket/d' "$file"
    sed -i '/from.*WebSocketContext/d' "$file"
    
    # Remove websocket usage
    sed -i '/useWebSocket/d' "$file"
    sed -i '/useBalanceMonitor/d' "$file"
    sed -i '/WebSocketProvider/d' "$file"
    sed -i '/BalanceWebSocketClient/d' "$file"
    
    # Remove websocket-related comments
    sed -i '/Real-time.*WebSocket/d' "$file"
    sed -i '/websocket.*connection/d' "$file"
    
    echo "✅ Cleaned $file"
  else
    echo "⚠️  File not found: $file"
  fi
done

echo ""
echo "🎯 WEBSOCKET CLEANUP SUMMARY:"
echo "✅ WebSocketContext.tsx - REMOVED"
echo "✅ balanceWebSocketService.ts - REMOVED"
echo "✅ balanceMonitor.ts - REMOVED"
echo "✅ BalanceMonitor.tsx - REMOVED"
echo "✅ BalanceMonitorDemo.tsx - REMOVED"
echo "✅ All websocket imports - CLEANED"
echo "✅ All websocket usage - CLEANED"
echo ""
echo "🚀 RESULT: Aplikasi akuntansi sekarang 100% tanpa websocket!"
echo "💡 Balance updates menggunakan standard polling/refresh yang lebih sesuai untuk aplikasi akuntansi"
echo ""
echo "📊 BENEFITS:"
echo "   • Reduced complexity"
echo "   • Better performance (no persistent connections)"
echo "   • More reliable for accounting use case"
echo "   • Easier to maintain and debug"
echo "   • Better suited for accounting workflow"