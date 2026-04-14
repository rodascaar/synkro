#!/bin/bash

echo "=== Synkro TUI Test Script ==="
echo ""
echo "This script will test the TUI in non-interactive mode"
echo ""

cd /Users/home/Downloads/nichogram/synkro

echo "Step 1: Check database..."
if [ -f "memory.db" ]; then
    echo "✓ Database exists"
    echo ""
    echo "Step 2: List memories..."
    go run ./cmd/synkro/ list
    MEMORY_COUNT=$(go run ./cmd/synkro/ list 2>/dev/null | wc -l)
    echo ""
    echo "✓ Found $MEMORY_COUNT memories"
else
    echo "✗ Database not found, initializing..."
    go run ./cmd/synkro/ init
fi

echo ""
echo "Step 3: Test TUI compilation..."
go build -o synkro ./cmd/synkro/ 2>&1
if [ $? -eq 0 ]; then
    echo "✓ TUI compiled successfully"
    echo ""
    echo "Step 4: TUI Test Instructions..."
    echo ""
    echo "The TUI is ready for interactive testing."
    echo ""
    echo "To test manually, run:"
    echo "  ./synkro tui"
    echo ""
    echo "Expected behavior:"
    echo "  • Full-screen interface (AltScreen)"
    echo "  • 3 panels: sidebar (filters), content (memories), detail (selected)"
    echo "  • Use ↑/↓ or j/k to navigate"
    echo "  • Press / to search"
    echo "  • Press g to toggle graph view"
    echo "  • Press Ctrl+C to quit"
    echo ""
    echo "Troubleshooting:"
    echo "  • If TUI doesn't start: Resize terminal to min 120x40"
    echo "  • If navigation doesn't work: Press Esc to exit search mode"
    echo "  • If no memories: Check database has data"
    echo ""
    echo "For detailed testing guide, see TUI_TEST.md"
else
    echo "✗ Compilation failed"
    exit 1
fi

echo ""
echo "=== Test Complete ==="
echo ""
echo "Binary size: $(ls -lh synkro | awk '{print $5}')"
echo "Database size: $(ls -lh memory.db | awk '{print $5}')"
echo ""
