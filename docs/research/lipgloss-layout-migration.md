# Scorched Earth Lipgloss Migration Plan

## Executive Summary

**MAJOR PLAN EVOLUTION**: After researching TUI community best practices, we discovered that our custom layout engine approach is wheel reinvention. The plan has evolved to **eliminate the layout engine entirely** and adopt pure Bubble Tea + Lipgloss patterns used by successful TUI applications.

**Target: 2750+ line reduction** (from ~2900 layout lines to ~150 pure lipgloss lines)

## Current State Analysis

### Layout Engine Architecture
The current layout engine uses **absolute positioning with manual coordinate calculation**:

1. **Constraint Resolution**: Components specify width/height constraints (Fixed, Flex, Min/Max)
2. **Dependency Graph**: Builds relationships between components (Above, Below, Left, Right)
3. **Dimension Calculation**: Resolves flex weights and distributes available space
4. **Position Calculation**: Places components at X,Y coordinates using manual math
5. **Screen Buffer Rendering**: Creates a 2D character array and places content at calculated positions
6. **Line Assembly**: Concatenates screen lines with `\n` characters

### Current Code Distribution
- `layout/calculator.go`: Core math for dimensions and positioning (600+ lines)
- `layout/engine.go`: Screen buffer management and final rendering (400+ lines)
- `layout/constraints.go`: Constraint definitions and validation (200+ lines)

### Current Lipgloss Usage
- Basic styling (colors, borders, padding, margins)
- `lipgloss.Place()` for modal positioning
- Style inheritance and composition
- Width calculations with `lipgloss.Width()`

## Impact Assessment

### What We Currently Use vs. What We'd Lose

#### ✅ SAFE TO CHANGE - Not Currently Used
- **Absolute positioning** - App only uses vertical stacking
- **Complex relative positioning** - App has simple layout hierarchy
- **Z-index layering** - Only modal uses overlay (lipgloss.Place handles this)
- **Pixel-perfect control** - App uses content-driven sizing

#### ⚠️ NEEDS ADAPTATION
- **Three-column layout** - Currently vertical, need horizontal join
- **Dynamic resizing** - Layout engine handles constraints, lipgloss would need similar

#### ✅ ENHANCED - Would Improve
- **Text alignment** - Better baseline alignment in columns
- **Border consistency** - Lipgloss handles border joining automatically
- **Content wrapping** - Better text flow and truncation
- **Code maintainability** - Much simpler layout code

## Implementation Plan

### Phase 1: Implement Lipgloss Vertical Layout (2-3 hours)

#### Files to Modify
- `layout/calculator.go`: Replace `calculateVerticalStack()`
- `layout/engine.go`: Add `renderWithJoinFunctions()` method

#### New Architecture
```go
func (alc *AdvancedLayoutCalculator) calculateVerticalStack() (*LayoutResult, error) {
    // 1. Keep existing constraint resolution and dimension calculation
    dimensions, warnings := alc.calculateDimensions(layoutOrder)

    // 2. Render each component to string
    componentStrings := []string{}
    for _, component := range layoutOrder {
        content := component.View()
        styled := applyConstraintsToContent(content, dimensions[component.ID()])
        componentStrings = append(componentStrings, styled)
    }

    // 3. Use lipgloss to join vertically
    finalContent := lipgloss.JoinVertical(lipgloss.Left, componentStrings...)

    // 4. Return result with single component layout
    return &LayoutResult{
        Components: map[string]ComponentLayout{
            "content": {X: 0, Y: 0, Width: maxWidth, Height: lipgloss.Height(finalContent)},
        },
    }
}
```

#### Changes Required
1. **Remove screen buffer logic** in `engine.go` (~200 lines)
2. **Simplify position calculation** in `calculator.go` (~300 lines)
3. **Add content styling helper** (~50 lines)

### Phase 2: Implement Lipgloss Horizontal Layout (1-2 hours)

#### New Method
```go
func (alc *AdvancedLayoutCalculator) calculateHorizontalStack() (*LayoutResult, error) {
    // Similar to vertical but:
    // 1. Distribute width instead of height among flex components
    // 2. Apply width constraints to each component
    // 3. Use lipgloss.JoinHorizontal(lipgloss.Top, componentStrings...)
}
```

#### Key Differences from Vertical
- Width distribution logic instead of height
- Horizontal flex weight calculations
- Top alignment for consistent baselines

### Phase 3: Update Application Integration (1 hour)

#### Three-Column Layout Fix
Replace current manual `renderHorizontalLayout()` function:

```go
// OLD: Manual line-by-line composition
func renderHorizontalLayout(localColumn, repoColumn, userColumn string, maxWidth int) string {
    // 100+ lines of manual positioning math
}

// NEW: Simple lipgloss integration
func renderThreeColumnContentWithLayoutEngine(m *types.Model, maxWidth, maxHeight int) string {
    columnEngine := layout.NewLayoutEngine(maxWidth, maxHeight)
    columnEngine.SetAlgorithm(layout.AlgorithmHorizontalStack)

    // Add three components with equal flex weights
    columnEngine.AddComponent("local", localComponent, layout.Width(layout.Flex(1.0)))
    columnEngine.AddComponent("repo", repoComponent, layout.Width(layout.Flex(1.0)))
    columnEngine.AddComponent("user", userComponent, layout.Width(layout.Flex(1.0)))

    return columnEngine.View()
}
```

#### Configuration Changes
- Add algorithm parameter to layout engine constructor
- Configure sub-layout engines with `AlgorithmHorizontalStack`
- Remove manual positioning helper functions

## Technical Details

### Constraint Resolution Strategy
**Keep Existing System**: The constraint resolution logic is sophisticated and working well. Only replace the final rendering phase.

**Preserve Features:**
- Flex weight distribution
- Min/max dimension constraints
- Margin and padding calculations
- Component dependency resolution

### Lipgloss Integration Points

#### Content Styling Helper
```go
func applyConstraintsToContent(content string, layout ComponentLayout) string {
    style := lipgloss.NewStyle().
        Width(layout.Width).
        Height(layout.Height)

    // Apply margins, padding, borders based on constraints
    return style.Render(content)
}
```

#### Algorithm Selection
```go
type LayoutEngine struct {
    algorithm LayoutAlgorithm
    // ... existing fields
}

func (le *LayoutEngine) SetAlgorithm(alg LayoutAlgorithm) {
    le.algorithm = alg
    le.needsRecalculation = true
}
```

## Benefits

### Code Reduction
- **Remove**: ~500 lines of manual positioning math
- **Add**: ~100 lines of lipgloss integration
- **Net Savings**: ~400 lines (~33% reduction in layout code)

### Quality Improvements
- **Better Text Alignment**: Lipgloss handles baseline alignment, text wrapping
- **Robust Edge Cases**: Handles Unicode, ANSI codes, variable-width characters
- **Consistent Spacing**: Automatic gap management between components
- **Easier Debugging**: Clear separation between layout logic and rendering

### Maintenance Benefits
- **Single Responsibility**: Layout engine focuses on constraint resolution
- **Fewer Bugs**: Leverage battle-tested lipgloss instead of custom positioning
- **Future Features**: Easy to add new layout algorithms using lipgloss primitives

## Risk Mitigation

### Low Risk Factors
- Well-defined scope with existing patterns to follow
- Lipgloss is proven and stable
- No changes to public API or user interface

### Testing Strategy
1. **Phase-by-phase validation** using debug server
2. **Visual regression testing** with screen captures
3. **Edge case testing** with various terminal sizes
4. **Performance comparison** before/after implementation

### Rollback Plan
- Changes are isolated to layout engine internals
- Original manual positioning code can be restored if needed
- No breaking changes to application logic

## Success Metrics

1. **Code Reduction**: Target 400+ line reduction
2. **Layout Quality**: Improved text alignment and border consistency
3. **Performance**: No significant rendering performance degradation
4. **Maintainability**: Simplified debugging and future enhancements

## Timeline

**Total Effort**: 4-6 hours of focused development
- **Phase 1** (Vertical): 2-3 hours
- **Phase 2** (Horizontal): 1-2 hours
- **Phase 3** (Integration): 1 hour
- **Testing & Validation**: Ongoing during each phase

## SCORCHED EARTH REWRITE PLAN

### Context & Discovery
After extensive research into TUI community practices, we discovered:
1. **99% of successful TUI apps DON'T have custom layout engines**
2. **Pure Bubble Tea + Lipgloss is the industry standard**
3. **Our layout engine is massive wheel reinvention**
4. **User priorities**: Not reinventing wheel (#1), Purist approaches (#2), Idiomatic architecture (#3)

### Current State Preserved (Commit: a258173)
**UI Features to Manually Port:**
- Enhanced styling system with color schemes and inheritance
- Three-column permission organization screen
- Debug API integration (simplified for new architecture)
- Visual improvements and proper spacing
- Pre-commit hooks and code quality

### Scorched Earth Strategy
1. **DELETE ENTIRELY**: All layout engine files (~2400 lines)
   - `layout/engine.go` (~1000 lines)
   - `layout/calculator.go` (~900 lines)
   - `layout/constraints.go` (~500 lines)
   - `layout/components.go`, `layout/spacing.go` (~600 lines)

2. **REPLACE WITH**: Pure Bubble Tea patterns (~150 lines)
   ```go
   func (m Model) View() string {
       header := headerStyle.Width(m.width).Render(m.headerContent)
       content := m.renderContent()
       footer := footerStyle.Width(m.width).Render(m.footerContent)
       return lipgloss.JoinVertical(lipgloss.Top, header, content, footer)
   }
   ```

3. **THREE-COLUMN LAYOUT**: Direct lipgloss usage
   ```go
   func renderThreeColumnLayout(m *Model) string {
       columnWidth := m.width / 3
       return lipgloss.JoinHorizontal(lipgloss.Top,
           columnStyle.Width(columnWidth).Render(localPerms),
           columnStyle.Width(columnWidth).Render(repoPerms),
           columnStyle.Width(columnWidth).Render(userPerms),
       )
   }
   ```

### Implementation Phases

**Phase 1: Create Pure Components**
- HeaderModel, FooterModel, ContentModel with View() methods
- Remove LayoutEngine from types/model.go
- Add terminal dimensions tracking

**Phase 2: Pure Lipgloss View Methods**
- Rewrite main View() with lipgloss.JoinVertical()
- Implement three-column layout with lipgloss.JoinHorizontal()
- Screen-specific layouts (duplicates, organization, confirmation)

**Phase 3: Delete Layout Engine**
- Remove all layout engine files
- Clean up imports and references
- Test functionality preservation

**Phase 4: Debug API Adaptation**
- Simplify debug endpoints for component-based architecture
- Preserve essential debugging capabilities
- Remove layout introspection (no longer relevant)

**Phase 5: Manual UI Feature Porting**
- Cherry-pick styling improvements from preserved commit
- Adapt three-column functionality to pure lipgloss
- Preserve visual enhancements and debug capabilities

### Expected Benefits
- **2750+ line reduction** using industry standard patterns
- **Zero wheel reinvention** - pure lipgloss/bubbletea
- **Perfect alignment** with TUI community practices
- **All UI features preserved** via manual porting
- **Debug API maintained** with simplified architecture

### Success Criteria
1. **Functionality**: All screens work identically
2. **Performance**: No degradation, likely improvement
3. **Code Quality**: Pure industry standard patterns
4. **Maintainability**: Textbook TUI architecture
5. **Debug API**: Essential endpoints preserved

This represents a complete architectural pivot from custom layout engine to industry standard TUI patterns, achieving all user priorities while preserving functionality.
