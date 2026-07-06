package warp

import "github.com/charmbracelet/lipgloss"

// Gruvbox Dark (https://github.com/morhetz/gruvbox)
// Backgrounds
var (
	gbDark0  = lipgloss.Color("#282828")
	gbDark1  = lipgloss.Color("#3c3836")
	gbDark2  = lipgloss.Color("#504945")
	gbDark3  = lipgloss.Color("#665c54")
	gbDark4  = lipgloss.Color("#7c6f64")
	gbGray   = lipgloss.Color("#928374")
	gbLight1 = lipgloss.Color("#ebdbb2")
	gbRed    = lipgloss.Color("#fb4934")
	gbGreen  = lipgloss.Color("#b8bb26")
	gbYellow = lipgloss.Color("#fabd2f")
	gbBlue   = lipgloss.Color("#83a598")
)

// Tab bar colors
var (
	tabBarBg      = gbDark0
	activeTabBg   = gbDark2
	activeTabFg   = gbLight1
	inactiveTabFg = gbGray
	newTabFg      = gbGreen
	closeTabFg    = gbRed
)

// Split border colors
var (
	borderColor      = gbDark1
	borderDragColor  = gbYellow
	borderHoverColor = gbDark3
)

// BorderStyle returns the style used for normal split borders.
func BorderStyle() lipgloss.Style { return borderStyle }

// BorderDragStyle returns the style used while dragging a split border.
func BorderDragStyle() lipgloss.Style { return borderDragStyle }

// BorderHoverStyle returns the style used when the mouse hovers a split border.
func BorderHoverStyle() lipgloss.Style { return borderHoverStyle }

// Float pane colors
var (
	floatBorderColor = gbGray
	floatTitleBg     = gbDark1
	floatTitleFg     = gbLight1
	floatBg          = gbDark0
	floatCloseFg     = gbRed
)

// Tab bar styles
var (
    tabBarStyle = lipgloss.NewStyle().
            Background(tabBarBg)

    activeTabStyle = lipgloss.NewStyle().
            Background(activeTabBg).
            Foreground(activeTabFg).
            Bold(true)

    inactiveTabStyle = lipgloss.NewStyle().
            Background(tabBarBg).
            Foreground(inactiveTabFg)

    newTabStyle = lipgloss.NewStyle().
            Foreground(newTabFg)

    closeTabStyle = lipgloss.NewStyle().
            Foreground(closeTabFg)
)

// Border styles
var (
    borderStyle = lipgloss.NewStyle().
            Foreground(borderColor)

    borderHoverStyle = lipgloss.NewStyle().
            Foreground(borderHoverColor)

    borderDragStyle = lipgloss.NewStyle().
            Foreground(borderDragColor)
)

// Float pane styles
var (
    floatBorderStyle = lipgloss.NewStyle().
            Foreground(floatBorderColor)

    floatTitleStyle = lipgloss.NewStyle().
            Background(floatTitleBg).
            Foreground(floatTitleFg).
            Bold(true)

    floatCloseStyle = lipgloss.NewStyle().
            Foreground(floatCloseFg).
            Bold(true)

    floatBgStyle = lipgloss.NewStyle().
            Background(floatBg)
)

// Collapsible styles
var (
    collapsibleStyle = lipgloss.NewStyle().
            Foreground(gbLight1).
            Background(gbDark1)

    collapsibleBorderStyle = lipgloss.NewStyle().
            Foreground(gbDark4)
)

// Dropdown styles
var (
    dropdownButtonStyle = lipgloss.NewStyle().
            Background(gbDark2).
            Foreground(gbLight1)

    dropdownItemStyle = lipgloss.NewStyle().
            Background(gbDark0).
            Foreground(gbLight1)

    dropdownItemHoverStyle = lipgloss.NewStyle().
            Background(gbDark2).
            Foreground(gbYellow)

    dropdownItemSelectedStyle = lipgloss.NewStyle().
            Background(gbDark2).
            Foreground(gbGreen).
            Bold(true)
)

// Popover styles in popover.go
