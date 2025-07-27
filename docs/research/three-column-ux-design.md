# Three-Column UX Design

## Final Implementation Plan

### Multi-Screen Workflow Architecture

**State-Driven Progressive Disclosure**: The application uses distinct screens for different tasks, eliminating TAB navigation complexity and providing focused, single-purpose interfaces.

**Two Main Screens:**
1. **Duplicates Resolution Screen** (when duplicates exist)
2. **Permission Organization Screen** (when no duplicates remain)

**Incremental Save Workflow:**
- Each screen transition triggers its own save operation
- Users review smaller chunks of changes instead of batching everything
- Clear checkpoints prevent data loss and provide natural stopping points

**Screen Continuity Design:**
- Header/Footer maintain consistent positioning across screens
- Only middle content area changes between screens
- Same status bar location for contextual information
- Consistent visual anchors for spatial orientation

### Layout Structure

**Three-column layout with optimal column ordering:**
- **LOCAL (left)**: Primary workspace, largest column, primary focus
- **REPO (center)**: Common target for team-shared tools
- **USER (right)**: Global defaults, typically fewer permissions

Rationale: Claude Code dumps all permissions to LOCAL by default, so users primarily work left-to-right, moving permissions out of LOCAL to REPO/USER.

```
╭─── LOCAL (8) ────╮  ╭─── REPO (4) ─────╮  ╭─── USER (2) ────╮
│   Bash           │  │   Grep           │  │   Edit          │
│   Glob           │  │   LS             │  │   Read          │
│   mcp__ide__*    │  │   Task           │  │                 │
│   Write          │  │   WebSearch      │  │                 │
│   ...            │  │                  │  │                 │
╰──────────────────╯  ╰──────────────────╯  ╰─────────────────╯
```

### Mental Model: Direct Manipulation WYSIWYG

- **Columns represent files** that will receive changes (LOCAL/REPO/USER settings)
- Users arrange permissions where they want them to end up
- ENTER commits the visual arrangement to actual settings files
- No multi-select complexity - focus on single-item operations
- Each screen handles one specific task (resolve conflicts OR organize permissions)

### Screen-Specific Navigation & Controls

**Duplicates Resolution Screen:**
- **Up/Down arrows**: Navigate between duplicate conflicts
- **1/2/3 keys**: Resolve conflict (keep permission in LOCAL/REPO/USER)
- **ENTER**: Save resolutions & advance to organization screen
- **ESC**: Cancel and exit

**Permission Organization Screen:**
- **Up/Down arrows**: Navigate within current column
- **Left/Right arrows**: Switch between columns
- **1/2/3 keys**: Move selected permission to LOCAL/REPO/USER column
- **ENTER**: Save moves & exit application
- **ESC**: Cancel and exit

**Rationale:**
- Separates navigation from actions clearly
- Numbers 1-2-3 map naturally to left-to-right column order
- Left hand on numbers, right hand on arrows for ergonomics
- Single keystroke actions are fast and memorable
- No TAB navigation needed - each screen has single focus area

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

### Duplicates Resolution Design

**Table-Based Interface**: Uses `bubbles/table` component for optimal UX with built-in row highlighting and navigation.

**Layout Structure:**
```
╭─── RESOLVE DUPLICATES (3 remaining) ────────────────────────╮
│ Step 1: Resolve Duplicates                                  │
├──────────────────────────────────────────────────────────────┤
│ LOCAL            REPO             USER                      │
├──────────────────────────────────────────────────────────────┤
│ Bash             Bash                                        │
│                                                              │
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│░Edit                             Edit                      ░│ ← Selected
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░│
│                                                              │
│ WebSearch        WebSearch       WebSearch                   │
├──────────────────────────────────────────────────────────────┤
│ ENTER: Save resolutions & continue                          │
╰──────────────────────────────────────────────────────────────╯
```

**Real-Time Visual Feedback**: When user selects resolution (1/2/3), non-selected instances immediately show:
- Faded text color (transparency effect)
- Strikethrough text
- Selected choice remains normal styling

**Benefits:**
- Full-width row highlighting (built into table component)
- Spatial consistency with three-column organization screen
- Clear visual feedback of resolution choices
- Standard table navigation patterns users expect

### Success Confirmation Modals

**Implementation**: Use `lipgloss.Place()` to center modal over existing view with background dimming.

**Modal Content Example:**
```
┌─────────────────────────────────────┐
│ ✓ 3 duplicates resolved             │
│                                     │
│ • Bash → kept in REPO               │
│ • Edit → kept in LOCAL              │
│ • WebSearch → kept in USER          │
│                                     │
│ [ENTER] Continue to organize        │
└─────────────────────────────────────┘
```

**Usage Pattern:**
- Success confirmation only (not for errors)
- Brief audit summary of what was accomplished
- Clear next step indication
- Modal auto-dismisses on ENTER to continue workflow

### Terminal Width Handling

**Graceful Degradation**: Enforce minimum width constraint rather than responsive layout changes.

**Approach:**
- Set minimum terminal width (e.g., 120 characters)
- If terminal is smaller, maintain layout but add horizontal scrolling if needed
- No complex responsive behavior or single-column fallbacks
- Simple and predictable behavior across terminal sizes

**Rationale:**
- Avoids complex layout switching logic
- Maintains consistent spatial relationships
- Users understand minimum requirements clearly

### Error Handling Philosophy

**Prevention Over Recovery:**
- Preemptive validation to avoid error states
- Design interface to prevent invalid actions
- Clear visual cues for valid/invalid states

**Error Display Strategy:**
- Errors appear inline within panel borders where action occurred
- Modal errors only for non-deterministic failures (file I/O, etc.)
- Failed operations keep user on current screen with error context
- No progression to next screen until current screen operations succeed

### Corner Cases Handled

- **Long permission names**: Truncate with ellipsis (`"mcp__github__create_iss..."`)
- **Empty columns**: Show placeholder `"(no permissions)"`
- **Navigation edges**: Left arrow from LOCAL stays put (no wrapping)
- **Status bar overflow**: Truncate permission name in status display
- **Terminal resize**: Dynamic column width recalculation via layout engine

### Visual Design & Color Scheme

**Dark Theme Only**: Application uses dark theme exclusively. All adaptive color configurations should be simplified to use only dark theme values.

**Column Focus Indication**: Uses existing application color scheme for consistency.

**Focused Column Border**:
- Uses `activePanelStyle` with cyan color (`#38BDF8`)
- Indicates which column currently has focus for navigation and actions

**Non-Focused Column Borders**:
- Uses `panelStyle` with gray color (`Color("8")`)
- Maintains visual hierarchy without distraction

**Item Selection Within Columns**:
- Standard single-item highlighting per column (no full-row spanning)
- Each column operates independently with its own selection state
- Navigation: Left/Right switches column focus, Up/Down navigates within focused column

**Code Cleanup Required**: Remove all `lipgloss.AdaptiveColor` usage and light theme references from styles.go and related files. Simplify to dark theme values only.

### Status Bar Design & Placement

**Research-Based Design**: Following established TUI patterns from successful applications (vim, tmux, htop, git TUIs).

**Layout Structure** (bottom to top):
```
├─ Main Content ─────┤ (permissions/duplicates panels)
├─ Status Bar ───────┤ ← NEW: contextual selection info
└─ Footer ───────────┘ ← EXISTING: action hints
```

**Visual Treatment:**
- **Status Bar**: Inverted colors (light background, dark text) for visual hierarchy
- **Footer**: Current dark styling (dark background, light text)
- **Clear separation**: Different background colors distinguish information types

**Content Strategy:**
- **Status Bar**: Selection context + summary counts
  - Permissions screen: `"Edit (originally USER → in LOCAL)     [LOCAL 8 | REPO 4 | USER 2]"`
  - Duplicates screen: `"Edit conflict: LOCAL vs USER (choose 1/3)     [3 conflicts remaining]"`
- **Footer**: Action hints
  - `"ENTER: Save & continue     ESC: Cancel"`

**Lifecycle**: Status bar always reserves fixed space (no layout shifting) but content changes based on selection state:
- **With selection**: Show contextual information as above
- **No selection**: Show general state (`"Step 1: Resolve Duplicates"` or `"Ready to organize permissions"`)

**Implementation Notes:**
- Use established TUI contrast ratios for accessibility (4.5:1 minimum)
- Follow bottom-positioning convention from successful terminal applications
- Integrate status and action information in shared bottom area with visual distinction

### Implementation Details

**Modal Behavior:**
- Success modals require explicit ENTER keypress to dismiss (no auto-dismissal)
- User maintains control over workflow progression
- Modal remains until user acknowledges and chooses to continue

**Terminal Size Constraints:**
- Minimum terminal width determined by column layout requirements plus padding
- Three-column layout with borders and spacing defines natural minimum size
- No artificial size restrictions beyond what layout geometry requires

**Empty State Handling:**
- Empty columns show no placeholder text (completely empty)
- Clean visual presentation without unnecessary UI clutter
- Column headers remain visible for spatial orientation

**File I/O Operations:**
- Synchronous operations acceptable for brief pauses (target <500ms)
- No async loading states or progress indicators needed initially
- Simple pause in UI interaction during save operations

**Error Recovery:**
- Failed save operations keep modal visible until dismissed
- User returns to current screen (duplicates or permissions) after modal dismissal
- Natural retry workflow - user can re-attempt save operation
- No forced screen transitions on error conditions

**Same-Level Duplicate Handling:**
- Auto-resolve duplicates within same level (e.g., "Foo" appearing twice in LOCAL)
- No user interaction required - automatic cleanup during initial data load
- Include brief mention in success modal: `"✓ 3 conflicts resolved, 2 duplicate entries cleaned up"`
- Maintains data consistency without adding UI complexity

**Application Startup Flow:**
- Load all three level JSON files (USER/REPO/LOCAL)
- Run duplicate detection across all levels
- Auto-resolve same-level duplicates silently
- Determine starting screen based on remaining cross-level duplicates
- Show duplicates screen if conflicts exist, otherwise show permissions screen

**Screen Transition Logic:**
- No automatic screen transitions - all progression is user-driven
- Two possible workflows:
  - **With duplicates**: Duplicates Screen → User submits → Permissions Screen → User submits → Exit
  - **No duplicates**: Permissions Screen → User submits → Exit
- Each submit operation saves current screen state before progressing

### Final Keyboard Shortcuts

**Duplicates Resolution Screen:**
```
Navigation:
  ↑↓    Navigate between conflicts

Actions:
  1     Keep in LOCAL
  2     Keep in REPO
  3     Keep in USER
  ENTER Save & continue to organize
  ESC   Cancel/exit
```

**Permission Organization Screen:**
```
Navigation:
  ↑↓    Navigate within column
  ←→    Switch between columns

Actions:
  1     Move to LOCAL
  2     Move to REPO
  3     Move to USER
  ENTER Save & exit
  ESC   Cancel/exit
  /     Filter mode (future)
```

---

## Technical Implementation Analysis

### Current Architecture Assessment

**Data Structures (types/model.go):**
- `Model` struct contains all application state with mutex for thread safety
- `Permission` struct tracks current level, pending moves, and selection state
- `Duplicate` struct handles cross-level conflicts with keep level resolution
- Action queue system already exists for batching operations

**Settings Management (settings.go):**
- Robust loading system with chezmoi integration and git root detection
- `detectDuplicates()` function only handles cross-level duplicates currently
- `consolidatePermissions()` creates unified view, giving priority to first occurrence
- Save operations preserve non-permission JSON fields in settings files

**Current UI System (ui.go):**
- Two-panel layout: permissions (list) + duplicates (table)
- TAB navigation between panels with `ActivePanel` state
- Confirmation dialog system using `ConfirmMode` boolean
- Layout engine integration for responsive positioning

**Action System (actions.go):**
- Action queue with move and duplicate resolution operations
- Batch execution with atomic save operations to all three files
- `removePermission()` removes ALL occurrences (perfect for same-level cleanup)

### Implementation Plan for Three-Column Design

**1. Same-Level Duplicate Auto-Resolution:**
```go
// New function in settings.go
func autoResolveSameLevelDuplicates(level *types.SettingsLevel) int {
    seen := make(map[string]bool)
    cleaned := []string{}
    duplicatesRemoved := 0

    for _, perm := range level.Permissions {
        if !seen[perm] {
            seen[perm] = true
            cleaned = append(cleaned, perm)
        } else {
            duplicatesRemoved++
        }
    }

    level.Permissions = cleaned
    return duplicatesRemoved
}
```

**2. Screen State Management:**
```go
// Add to types/model.go
const (
    ScreenDuplicates = iota
    ScreenOrganization
)

type Model struct {
    // ... existing fields
    CurrentScreen int
    CleanupStats  struct {
        DuplicatesResolved int
        SameLevelCleaned   int
    }
}
```

**3. Three-Column Permissions Layout:**
- Replace `list.Model` with three separate viewports or custom column components
- Implement column focus tracking and left/right navigation
- Use existing `activePanelStyle` for focused column borders

**4. Table-Based Duplicates Screen:**
- Leverage existing `DuplicatesTable` (bubbles/table)
- Populate with three-column data: `[]table.Row{{"Edit", "Edit", ""}}`
- Implement visual feedback with strikethrough styling

**5. Modal Success Confirmations:**
- Use `lipgloss.Place()` for centering modals over content
- Replace current `ConfirmMode` with modal state management
- Include cleanup statistics in modal content

**6. Status Bar Implementation:**
- Add new layout component between main content and footer
- Use inverted styling (light background, dark text)
- Show selection context and permission counts

### Key Technical Insights

**Existing Strengths to Leverage:**
- Layout engine already supports multiple components with positioning
- Action queue system is perfect for incremental saves
- Mutex-protected state enables safe concurrent operations
- Table component provides built-in full-row highlighting

**Required Changes:**
- Modify `detectDuplicates()` to run auto-cleanup first
- Implement screen transition logic in main update loop
- Replace single list component with three-column layout
- Add status bar as new layout component
- Remove TAB navigation in favor of screen-based progression

**Minimal Breaking Changes:**
- Current data structures are compatible with new design
- Existing save operations can be reused for incremental workflow
- Modal system can replace confirmation dialog with minimal changes

This analysis confirms the design is technically feasible with the existing architecture.

---

## Design Evolution & Research Summary

### Final Design Decisions (January 2025)

**Multi-Screen Workflow Decision**: After analyzing workflow dependencies, chose state-driven progressive disclosure over simultaneous panels.

**Key Insight**: Duplicates must be resolved before permissions can be organized, but showing both contexts simultaneously creates functional and cognitive conflicts:
- Same permission appearing in multiple columns breaks mental model
- Users cannot organize ambiguous states
- Simultaneous commit operations prevent sequential workflow review

**Solution**: Separate screens with incremental saves and clear progression:
1. **Duplicates Screen** → Save resolutions → **Organization Screen** → Save moves → Exit
2. Each screen has single purpose and clear commit boundaries
3. Modal confirmations provide checkpoint feedback between screens

### UX Research Validation

**TUI Conflict Resolution Patterns** (Research conducted January 2025):
- Git mergetool patterns confirm multi-column layouts as gold standard
- Full-width row highlighting solves spanning selection challenge
- Table-based interfaces provide expected navigation patterns
- Spatial consistency across screens improves user mental models

**Key Research Findings:**
- Users develop strong spatial memory in text interfaces requiring consistent element positioning
- Spatial relationships more important when visual differentiation is constrained
- Progressive disclosure reduces cognitive load in complex workflows
- Modal confirmations appropriate for batch operations in TUI context

**Bubble Tea/Lipgloss Technical Validation:**
- `bubbles/table` provides optimal row highlighting and navigation
- `lipgloss.Place()` enables proper modal implementation with background dimming
- Existing layout engine supports three-column responsive design
- Built-in viewport scrolling eliminates need for custom components

### Rejected Alternatives

**TAB Navigation Between Panels**: Eliminated in favor of single-purpose screens
- **Problem**: Added complexity without workflow benefit
- **Solution**: State-driven screen transitions match task boundaries

**Action Queue with Staging**: Rejected for duplicates workflow
- **Problem**: Abstract and removed from direct manipulation
- **Solution**: WYSIWYG visual feedback with immediate state changes

**Single-Column Duplicates List**: Rejected in favor of three-column table
- **Problem**: Broke spatial consistency with organization screen
- **Solution**: Table-based three-column layout maintains file-centric mental model

**Responsive Layout Switching**: Rejected for minimum width constraint
- **Problem**: Complex layout logic and inconsistent spatial relationships
- **Solution**: Graceful degradation with horizontal scrolling if needed

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
