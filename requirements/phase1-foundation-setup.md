# Phase 1 Foundation Setup Requirements - NTX Portfolio Management TUI

## Problem Statement

Create a clean, minimal foundation for the NTX (NEPSE Power Terminal) project that establishes the basic project structure, TUI framework, configuration system, and visual theme foundation without over-engineering or excessive complexity.

**Key Context**: The previous implementation was removed due to over-engineering concerns. This Phase 1 focuses on simplicity, clarity, and a solid foundation for future phases.

## Solution Overview

Build a minimal but complete foundation that includes:

1. **Go Project Structure**: Standard Go layout with proper module setup
2. **Basic TUI Skeleton**: Bubbletea framework with section navigation
3. **Theme Foundation**: Tokyo Night color scheme with theme switching capability
4. **Configuration Management**: Viper-based config with proper hierarchy
5. **Keyboard Navigation**: btop-inspired navigation (1-5 sections, hjkl, vim-like)
6. **Responsive Layout**: Multi-pane layout that adapts to terminal size

## Functional Requirements

### FR1: Project Structure Setup

- **FR1.1**: Initialize Go module with `go mod init ntx`
- **FR1.2**: Create standard Go project layout:
  ```
  ntx/
  ├── cmd/ntx/main.go          # Application entry point
  ├── internal/
  │   ├── app/                 # Application orchestration
  │   ├── ui/
  │   │   ├── dashboard/       # Main TUI components
  │   │   ├── themes/          # Color schemes
  │   │   └── components/      # Reusable UI elements
  │   └── config/              # Configuration management
  ├── configs/                 # Configuration templates
  ├── go.mod
  └── go.sum
  ```
- **FR1.3**: Install core dependencies: Bubbletea, Lipgloss, Viper
- **FR1.4**: Create `.gitignore` for Go projects

### FR2: Basic TUI Skeleton

- **FR2.1**: Implement Bubbletea Model-View-Update pattern
- **FR2.2**: Create 5 main sections:
  - [1] Overview: Portfolio summary and key statistics  
  - [2] Holdings: Current positions (default focus)
  - [3] Analysis: Placeholder for future metrics
  - [4] History: Placeholder for transaction history
  - [5] Market: Placeholder for market data
- **FR2.3**: Implement section switching with 1-5 number keys
- **FR2.4**: Add basic status bar showing current section and navigation help
- **FR2.5**: Display placeholder content for each section (no real data yet)

### FR3: Tokyo Night Theme Foundation

- **FR3.1**: Implement theme system with interface for future themes
- **FR3.2**: Create Tokyo Night color palette:
  ```
  Background: #1a1b26
  Foreground: #c0caf5
  Primary: #7aa2f7 (bright blue)
  Success: #9ece6a (green)
  Warning: #e0af68 (yellow)
  Error: #f7768e (red)
  Muted: #565f89 (dark blue-gray)
  ```
- **FR3.3**: Apply theme to UI components (borders, text, highlights)
- **FR3.4**: Add theme switching with 't' key
- **FR3.5**: Create theme interface for future theme additions

### FR4: Configuration Management

- **FR4.1**: Implement Viper configuration with hierarchy:
  - Command line flags > Environment variables > Config file > Defaults
- **FR4.2**: Create default config structure:
  ```toml
  [ui]
  theme = "tokyo_night"  
  default_section = "holdings"
  
  [display]
  refresh_interval = 30
  currency_symbol = "₹"
  ```
- **FR4.3**: Config file location: `~/.config/ntx/config.toml`
- **FR4.4**: Add command line flags: `--theme`, `--config`

### FR5: Keyboard Navigation (btop-inspired)

- **FR5.1**: Section switching: 1-5 keys for direct section access
- **FR5.2**: Vim-style movement within sections: h/j/k/l
- **FR5.3**: Tab/Shift+Tab for section cycling
- **FR5.4**: Basic navigation:
  - q: Quit application
  - ?: Show help/keybindings  
  - r: Refresh (placeholder for future data refresh)
  - t: Toggle theme
- **FR5.5**: Navigation state management (track current section, selected item)

### FR6: Responsive Multi-Pane Layout

- **FR6.1**: Create responsive layout system:
  - Wide terminals (>120 cols): 3-pane layout
  - Medium terminals (80-120 cols): 2-pane layout  
  - Narrow terminals (<80 cols): single pane with tab switching
- **FR6.2**: Layout components:
  - Header bar: Portfolio summary (always visible when space allows)
  - Main content area: Current section content
  - Status bar: Navigation help and current section indicator
- **FR6.3**: Handle terminal resize gracefully without restart
- **FR6.4**: Minimum terminal size handling (60x24)

## Technical Requirements

### TR1: Dependencies

```go
// Core TUI framework
github.com/charmbracelet/bubbletea v1.3.5
github.com/charmbracelet/lipgloss v1.1.0

// Configuration management  
github.com/spf13/viper v1.20.1

// Terminal utilities
golang.org/x/term v0.32.0
```

### TR2: Architecture Patterns

- **TR2.1**: Follow Bubbletea Model-View-Update pattern
- **TR2.2**: Use composition over inheritance for UI components
- **TR2.3**: Implement interfaces for themes and sections for extensibility
- **TR2.4**: Clean separation between UI logic and business logic (prepare for future data layer)

### TR3: Performance Requirements

- **TR3.1**: Application startup time: <200ms
- **TR3.2**: Section switching: <50ms response time
- **TR3.3**: Theme switching: <100ms response time
- **TR3.4**: Memory usage: <10MB baseline (no data loading yet)

### TR4: Code Quality

- **TR4.1**: Go standard formatting with `gofmt`
- **TR4.2**: No external runtime dependencies (single binary)
- **TR4.3**: Proper error handling and graceful degradation
- **TR4.4**: Clear, readable code with appropriate comments

## Implementation Plan

### Step 1: Project Bootstrap
1. Initialize Go module
2. Create directory structure
3. Install dependencies  
4. Create basic main.go entry point

### Step 2: Basic TUI Structure
1. Implement Bubbletea model with 5 sections
2. Add section switching logic
3. Create placeholder content for each section
4. Add status bar with navigation help

### Step 3: Theme System
1. Create theme interface and Tokyo Night implementation
2. Apply theme to all UI components
3. Add theme switching functionality
4. Style borders, text, and highlights

### Step 4: Configuration System
1. Implement Viper configuration setup
2. Create default config file template
3. Add command line flag parsing
4. Test configuration hierarchy

### Step 5: Navigation & Layout
1. Implement keyboard navigation handlers
2. Add responsive layout logic
3. Handle terminal resize events
4. Test on different terminal sizes

### Step 6: Polish & Testing
1. Test all navigation flows
2. Verify theme switching works
3. Test configuration loading
4. Add help documentation

## Acceptance Criteria

### AC1: Project Setup
- [x] `go mod init ntx` creates proper module
- [x] All dependencies install without errors
- [x] `go run cmd/ntx/main.go` launches application
- [x] Single binary builds with `go build cmd/ntx/main.go`

### AC2: TUI Functionality
- [x] Application starts with Holdings section focused by default
- [x] Keys 1-5 switch between sections (Overview, Holdings, Analysis, History, Market)
- [x] Each section shows placeholder content with section name
- [x] Status bar shows current section and key navigation hints
- [x] 'q' key quits application cleanly

### AC3: Theme System
- [x] Tokyo Night theme applied throughout interface
- [x] 't' key cycles through available themes
- [x] Theme colors consistent across all UI elements
- [x] Multiple theme options available (Tokyo Night, Rose Pine, Gruvbox, Default)

### AC4: Configuration
- [ ] Config file created at `~/.config/ntx/config.toml` on first run
- [ ] Command line flags override config file settings
- [ ] `--theme` flag changes theme
- [ ] Configuration loads without errors
- [ ] Theme preference persists (when config is saved)

### AC5: Navigation & Layout
- [ ] h/j/k/l keys work for navigation within sections
- [ ] Tab/Shift+Tab cycle between sections
- [ ] Layout adapts to terminal width (3-pane → 2-pane → 1-pane)
- [ ] Terminal resize handled without restart
- [ ] Minimum size handling shows appropriate message

### AC6: Code Quality
- [ ] Code passes `go fmt` check
- [ ] No runtime dependencies beyond stdlib
- [ ] Graceful error handling for invalid terminal sizes
- [ ] Help system ('?' key) shows all available keybindings

## Success Metrics

- **Startup Performance**: Application launches in <200ms
- **Responsiveness**: All navigation commands respond in <50ms
- **Memory Efficiency**: <10MB memory usage for basic UI
- **User Experience**: Clean, professional appearance matching btop aesthetics
- **Code Quality**: Passes all standard Go linting tools

## Constraints & Assumptions

### Constraints
- Must remain simple and focused (avoid over-engineering)
- No external data sources in Phase 1 (placeholder content only)
- Single binary deployment (no CGO dependencies)
- Support for standard terminal capabilities (256 colors minimum)

### Assumptions
- Users have Go 1.21+ installed for development
- Target terminals support 256-color palette
- Users are familiar with vim-style navigation
- Terminal size of at least 60x24 for full functionality

## Future Phase Preparation

This Phase 1 foundation prepares for:
- **Phase 2**: Database layer and data models
- **Phase 3**: Portfolio management and transaction entry  
- **Phase 4**: Additional themes and UI polish
- **Phase 5**: Market data integration
- **Phase 6**: Advanced analytics and reporting

The architecture should be extensible for these future phases while maintaining the simplicity established in Phase 1.