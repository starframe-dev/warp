package warp

import "github.com/charmbracelet/lipgloss"

// ThemeColors contains semantic colors used by Warp overlays and controls.
type ThemeColors struct {
	Background          string
	Surface             string
	Raised              string
	Border              string
	BorderMuted         string
	Text                string
	TextMuted           string
	TextStrong          string
	Accent              string
	AccentMuted         string
	Error               string
	Success             string
	Warning             string
	SelectionBackground string
	SelectionForeground string
}

// SetTheme updates all package-level Warp styles used by overlays and controls.
func SetTheme(colors ThemeColors) {
	gbDark0 = lipgloss.Color(colors.Background)
	gbDark1 = lipgloss.Color(colors.Surface)
	gbDark2 = lipgloss.Color(colors.Raised)
	gbDark3 = lipgloss.Color(colors.BorderMuted)
	gbDark4 = lipgloss.Color(colors.Border)
	gbGray = lipgloss.Color(colors.TextMuted)
	gbLight1 = lipgloss.Color(colors.TextStrong)
	gbRed = lipgloss.Color(colors.Error)
	gbGreen = lipgloss.Color(colors.Success)
	gbYellow = lipgloss.Color(colors.Warning)
	gbBlue = lipgloss.Color(colors.Accent)

	tabBarBg = gbDark0
	activeTabBg = lipgloss.Color(colors.SelectionBackground)
	activeTabFg = lipgloss.Color(colors.SelectionForeground)
	inactiveTabFg = gbGray
	newTabFg = gbGreen
	closeTabFg = gbRed

	borderColor = gbDark1
	borderDragColor = gbYellow
	borderHoverColor = gbDark3

	floatBorderColor = gbGray
	floatTitleBg = gbDark1
	floatTitleFg = gbLight1
	floatBg = gbDark0
	floatCloseFg = gbRed

	tabBarStyle = lipgloss.NewStyle().Background(tabBarBg)
	activeTabStyle = lipgloss.NewStyle().Background(activeTabBg).Foreground(activeTabFg).Bold(true)
	inactiveTabStyle = lipgloss.NewStyle().Background(tabBarBg).Foreground(inactiveTabFg)
	newTabStyle = lipgloss.NewStyle().Foreground(newTabFg)
	closeTabStyle = lipgloss.NewStyle().Foreground(closeTabFg)

	borderStyle = lipgloss.NewStyle().Foreground(borderColor)
	borderHoverStyle = lipgloss.NewStyle().Foreground(borderHoverColor)
	borderDragStyle = lipgloss.NewStyle().Foreground(borderDragColor)
	collapseStyle = lipgloss.NewStyle().Foreground(borderColor)

	floatBorderStyle = lipgloss.NewStyle().Foreground(floatBorderColor)
	floatTitleStyle = lipgloss.NewStyle().Background(floatTitleBg).Foreground(floatTitleFg).Bold(true)
	floatCloseStyle = lipgloss.NewStyle().Foreground(floatCloseFg).Bold(true)
	floatBgStyle = lipgloss.NewStyle().Background(floatBg)

	collapsibleStyle = lipgloss.NewStyle().Foreground(gbLight1).Background(gbDark1)
	collapsibleBorderStyle = lipgloss.NewStyle().Foreground(gbDark4)

	dropdownButtonStyle = lipgloss.NewStyle().Background(gbDark2).Foreground(gbLight1)
	dropdownItemStyle = lipgloss.NewStyle().Background(gbDark0).Foreground(gbLight1)
	dropdownItemHoverStyle = lipgloss.NewStyle().Background(gbDark2).Foreground(gbYellow)
	dropdownItemSelectedStyle = lipgloss.NewStyle().Background(gbDark2).Foreground(gbGreen).Bold(true)

	popoverBaseStyle = lipgloss.NewStyle().Background(gbDark1).Foreground(gbLight1)
	popoverSelectedStyle = lipgloss.NewStyle().Background(activeTabBg).Foreground(activeTabFg)

	modalBorderStyle = lipgloss.NewStyle().
		Background(gbDark1).
		Foreground(gbLight1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(gbDark4).
		Padding(1, 2)
	dimStyle = lipgloss.NewStyle().Foreground(gbGray).Background(gbDark0)

	inputStyle = lipgloss.NewStyle().Foreground(gbLight1)
	inputBorderStyle = lipgloss.NewStyle().Foreground(gbDark4)
	inputFocusBorderStyle = lipgloss.NewStyle().Foreground(gbBlue)
}
