# Synkro TUI - Professional Terminal UI

Synkro's TUI uses Bubble Tea and Lipgloss for a professional, modern terminal experience.

## Features

- **AltScreen**: Full-screen application feel, no scrolling
- **3-Panel Layout**: Sidebar (filters), Content (memories), Detail (selected)
- **Keyboard Navigation**: Vi-style (j/k) and arrow keys
- **Real-time Search**: Type '/' to filter memories instantly
- **Empty States**: Elegant messages when database is empty
- **Graph Visualization**: View memory relationships with 'g'
- **Help System**: Always-visible keyboard shortcuts

## Keyboard Shortcuts

### Navigation
- `↑` or `k` - Move up
- `↓` or `j` - Move down
- `←` or `h` - Focus sidebar
- `→` or `l` - Focus list
- `Enter` - Select/View details
- `Esc` - Back / Quit

### Actions
- `/` or `s` - Search memories
- `a` - Add new memory (coming soon)
- `g` - Toggle graph view
- `Ctrl+C` - Quit

## Layout

```
┌──────────────┐ ┌──────────────────────────────────────────────────────────┐
│   SYNKRO     │ │  MEMORIES (12)                    [ Search: _________ ]  │
├──────────────┤ ├──────────────────────────────────────────────────────────┤
│ > All        │ │ ● [Decision] Refactor TUI with Lipgloss         10:45 AM │
│   Decisions  │ │ ○ [Tech]     Switch to int8 embeddings          Yesterday│
│   Tasks      │ │ ○ [Idea]     SoccerInsight PY API               02 Feb   │
│   Notes      │ │                                                          │
│   Archive    │ │                                                          │
└──────────────┘ └──────────────────────────────────────────────────────────┤
                  │  DETAILS                                                 │
                  │  Type: Decision                                          │
                  │  ID: mem-20260410104512-1234                            │
                  │  Source: user                                           │
                  │                                                          │
                  │  Improving interface using Charm ecosystem...            │
                  │                                                          │
                  │  RELATIONS                                              │
                  │  -> related_to Switch to embeddings (1.0)               │
 ┌────────────────┴──────────────────────────────────────────────────────────┤
 │  ↑/k move • ↓/j move • / search • g graph • a add • Ctrl+C quit      │
 └───────────────────────────────────────────────────────────────────────────┘
```

## Color Scheme

- **Cyan (#00e5cc)**: Selected items, focus
- **White (#ffffff)**: Normal text
- **Gray (#666666)**: Secondary text, empty states
- **Gray (#444444)**: Borders

## Views

### Sidebar (Left)
Filter memories by type:
- All: Show all memories
- Decisions: Filter by type=decision
- Tasks: Filter by type=task
- Notes: Filter by type=note
- Archive: Filter by status=archived

### Content (Center)
Main memory list with:
- Memory count
- Active filter indicator
- Scrollable list with selection
- Search box when searching

### Detail (Right)
Selected memory details:
- Type, ID, Source
- Full content (word-wrapped)
- Relations (when graph enabled)
- Related memories with strength

## Empty States

When database is empty:
```
📥 Press 'a' to add your first memory
🔍 Press '/' to search
```

When no memories match filter:
```
No memories found matching your filter
```

## Performance

- **Instant updates**: No lag on navigation
- **Smart loading**: Loads 100 memories at once
- **Efficient rendering**: Only updates changed sections
- **Memory efficient**: Minimal RAM usage

## Customization

Change colors in `internal/tui/model.go`:
```go
headerStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#00e5cc")).  // Change this color
    Bold(true).
    Padding(0, 1)
```

Adjust panel widths:
```go
sidebarStyle = lipgloss.NewStyle().Width(20)  // Change width
contentStyle = lipgloss.NewStyle().Width(60)  // Change width
detailStyle = lipgloss.NewStyle().Width(40)   // Change width
```

## Troubleshooting

**Terminal too small**: Resize terminal to minimum 120x40
**Colors wrong**: Ensure terminal supports 256 colors
**AltScreen issues**: Some terminals may need `TERM=xterm-256color`

## Tips

1. **Vi users**: Use j/k for faster navigation
2. **Large databases**: Use search '/' to filter quickly
3. **Related memories**: Press 'g' to see graph
4. **Focus sidebar**: Press left/h to change filters
5. **Quick quit**: Ctrl+C always works
