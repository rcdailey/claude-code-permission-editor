# Three-Column UX Design

## Proposed Implementation Plan

### Layout Structure

**Three-column layout with optimal column ordering:**
- **LOCAL (left)**: Primary workspace, largest column, gets focus first on TAB
- **REPO (center)**: Common target for team-shared tools
- **USER (right)**: Global defaults, typically fewer permissions

Rationale: Claude Code dumps all permissions to LOCAL by default, so users primarily work left-to-right, moving permissions out of LOCAL to REPO/USER.

```
╭─── LOCAL (8) ────╮  ╭─── REPO (4) ─────╮  ╭─── USER (2) ────╮
│ [ ] Bash         │  │ [ ] Grep         │  │ [ ] Edit        │
│ [ ] Glob         │  │ [ ] LS           │  │ [ ] Read        │
│ [ ] mcp__ide__*  │  │ [ ] Task         │  │                 │
│ [ ] Write        │  │ [ ] WebSearch    │  │                 │
│ ...              │  │                  │  │                 │
╰──────────────────╯  ╰──────────────────╯  ╰─────────────────╯
```

### Mental Model: Direct Manipulation WYSIWYG

- **Columns represent desired final state** of permission distribution
- Users arrange permissions where they want them to end up
- ENTER commits the visual arrangement to actual settings files
- No multi-select complexity - focus on single-item operations

### Navigation & Controls

**Navigation (non-destructive):**
- **Up/Down arrows**: Navigate within current column
- **Left/Right arrows**: Switch between columns
- **TAB**: Switch between panels (permissions ↔ duplicates)

**Actions (destructive but deliberate):**
- **1**: Move selected item to LOCAL column
- **2**: Move selected item to REPO column
- **3**: Move selected item to USER column

**Rationale:**
- Separates navigation from actions clearly
- Numbers 1-2-3 map naturally to left-to-right column order
- Left hand on numbers, right hand on arrows for ergonomics
- Single keystroke actions are fast and memorable

### Original Location Tracking

**Status bar approach** for selected item:
```
> Bash (originally USER → moving to REPO)
> Edit (no changes)
```

**Benefits:**
- Doesn't clutter column layout
- Progressive disclosure (only when item selected)
- Shows both original and target locations
- Discoverable through normal navigation

### Technical Implementation

**Built-in Component Usage:**
- **Viewport scrolling**: Use `bubbles/viewport` for column scrolling with built-in indicators
- **Text truncation**: Use `lipgloss.Style.MaxWidth()` for long permission names
- **No checkboxes**: Eliminate multi-select complexity, use simple highlighting

**Corner Cases Handled:**
- **Long permission names**: Truncate with ellipsis (`"mcp__github__create_iss..."`)
- **Empty columns**: Show placeholder `"(no permissions)"` or hide if terminal narrow
- **Narrow terminals**: Fallback to single-column when width < 120 chars
- **Duplicate permissions**: Visual distinction for items appearing in multiple columns
- **Navigation edges**: Left arrow from LOCAL stays put (no wrapping)
- **Status bar overflow**: Truncate permission name in status display
- **Terminal resize**: Dynamic column width recalculation via layout engine

### Keyboard Shortcuts Summary

```
Navigation:
  ↑↓    Navigate within column
  ←→    Switch between columns
  TAB   Switch panels (permissions ↔ duplicates)

Actions:
  1     Move to LOCAL
  2     Move to REPO
  3     Move to USER
  ENTER Commit changes to files
  ESC   Cancel/return
  /     Filter mode
```

---

## Context & Historical Information

### Original Problem Analysis

The single-column permission list with right-aligned level indicators created UX issues:
- **Eye scanning fatigue**: Users had to scan horizontally across wide terminals to connect permissions with levels
- **Level as metadata**: Level information felt secondary rather than primary organizational structure
- **Poor distribution visibility**: Difficult to understand permission spread across levels at a glance
- **Underutilized screen space**: Wide terminals weren't leveraged effectively

### Design Evolution Process

**Initial Approach Considered**: Action queue system with staging area
- Users would select permissions and queue moves
- Separate confirmation step to apply changes
- **Rejected** because it felt abstract and removed from direct manipulation

**Multi-select Elimination Decision**:
- Original design included checkboxes for bulk operations
- **Analysis**: Permission management is inherently careful, one-at-a-time work
- Bulk operations are rare and error-prone in configuration management
- **Decision**: Eliminate checkboxes, focus on single-item efficiency
- **Result**: Cleaner visual design, faster navigation, reduced complexity

**Keyboard Shortcut Evolution**:
- **Original**: U/R/L keys (first letters of User/Repo/Local)
- **Problem**: Awkward reach, doesn't follow TUI conventions
- **Solution**: 1/2/3 number keys mapping to column positions
- **Rationale**: Left-hand efficiency, natural left-to-right mapping

**Column Ordering Research**:
- **Initial assumption**: USER → REPO → LOCAL (logical hierarchy)
- **User insight**: LOCAL is the "dumping ground" where Claude Code puts everything
- **Workflow reality**: Users primarily move items OUT of LOCAL
- **Decision**: LOCAL → REPO → USER (workflow-optimized)
- **Additional benefit**: LOCAL gets focus first when tabbing to permissions panel

### Technical Research Findings

**Bubble Tea/Lipgloss Capabilities**:
- `bubbles/viewport`: Provides built-in scrolling with visual indicators, methods like `AtTop()`, `AtBottom()`, `ScrollPercent()`
- `lipgloss.Style.MaxWidth()`: Automatic text truncation with proper handling
- No need to build custom UI components - leverage existing mature solutions

**Layout Engine Integration**:
- Existing layout engine can support three-panel responsive design
- Dynamic column width calculation needed for terminal resize
- Frame overhead calculations already available

### UX Principles Applied

- **Direct manipulation** over abstract action queues
- **Spatial grouping** over flat lists with remote labels
- **WYSIWYG preview** over hidden state changes
- **Progressive disclosure** of details (status bar only when needed)
- **Leverage screen real estate** instead of forcing narrow layouts
- **Separation of concerns**: Navigation vs action operations
- **Ergonomic key placement**: Frequent operations on home row area

### Key Design Insights

**Multi-select vs Single-select**: Permission management is inherently careful work done one item at a time. Single-select eliminates complexity while matching the actual workflow.

**Column ordering optimization**: LOCAL → REPO → USER matches the primary workflow of moving permissions out of Claude Code's default dumping ground (LOCAL) to more appropriate levels.

**Status bar for origin tracking**: Progressive disclosure approach keeps the main interface clean while providing necessary context when items are selected.

**Navigation separation**: Distinguishing navigation (arrows) from actions (numbers) prevents accidental moves and follows ergonomic principles.
