#!/bin/bash

echo "=== Synkro TUI - Complete Testing Script ==="
echo ""

cd /Users/home/Downloads/nichogram/synkro

echo "Step 1: Build TUI..."
go build -o synkro ./cmd/synkro/ 2>&1
if [ $? -ne 0 ]; then
    echo "✗ Build failed"
    exit 1
fi
echo "✓ Build successful"
echo ""

echo "Step 2: Check database..."
if [ -f "memory.db" ]; then
    echo "✓ Database exists"
    echo ""
    echo "Step 3: List all memories..."
    ./synkro list
    COUNT=$(./synkro list 2>/dev/null | wc -l)
    echo ""
    echo "✓ Found $COUNT memories"
else
    echo "✗ Database not found, initializing..."
    ./synkro init
fi

echo ""
echo "Step 4: Test search from command line..."
./synkro search "TUI"
echo ""

echo "Step 5: Launch TUI..."
echo ""
echo "=== TUI TESTING GUIDE ==="
echo ""
echo "🔍 TEST 1: SEARCH"
echo "  1. Press '/' to open search"
echo "  2. Type 'TUI' - should filter to 'Test TUI Navigation'"
echo "  3. Type more letters - continues filtering"
echo "  4. Press 'Esc' - closes search"
echo ""
echo "📋 TEST 2: DETAILS VIEW"
echo "  1. Navigate to any memory with ↑/↓"
echo "  2. Right panel shows:"
echo "     - Type (decision, task, note)"
echo "     - Full ID"
echo "     - Source, Status"
echo "     - Tags (if any)"
echo "     - Complete content"
echo ""
echo "🔗 TEST 3: RELATIONS"
echo "  1. Navigate to 'Nota: Relaciones en Synkro'"
echo "  2. Press 'g' - shows relations"
echo "  3. Read explanation of what relations mean"
echo "  4. Press 'g' again or 'Esc' to hide"
echo ""
echo "🎯 TEST 4: NAVIGATION"
echo "  - Use ↑/↓ or j/k to navigate"
echo "  - Press '/' to search"
echo "  - Press 'g' to toggle graph"
echo "  - Press 'Ctrl+C' to quit"
echo ""
echo "⚠️  TROUBLESHOOTING"
echo "  - If search finds nothing: Check spelling, try shorter terms"
echo "  - If details not showing: Select a memory first"
echo "  - If relations empty: Most memories have no relations yet"
echo "  - If TUI doesn't start: Resize terminal to min 120x40"
echo ""
echo "Press Enter to launch TUI..."
read

./synkro tui

echo ""
echo "=== TEST COMPLETE ==="
echo ""
echo "Thank you for testing Synkro TUI!"
echo ""
