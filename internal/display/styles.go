package display

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	LabelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	SecretStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
	CriticalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
	HighStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true)
	MediumStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	LowStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#4A9EFF"))
	BoxStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	CheckStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
)

// DisableColor strips ANSI styling from all lipgloss output globally.
func DisableColor() {
	lipgloss.DefaultRenderer().SetColorProfile(termenv.Ascii)
}

// SeverityStyle returns the lipgloss style for a severity label.
func SeverityStyle(severity string) lipgloss.Style {
	switch strings.ToUpper(severity) {
	case "CRITICAL":
		return CriticalStyle
	case "HIGH":
		return HighStyle
	case "MEDIUM":
		return MediumStyle
	case "LOW":
		return LowStyle
	default:
		return LabelStyle
	}
}

// Banner returns the boxed app title.
func Banner() string {
	return BoxStyle.Render(TitleStyle.Render("⚡ secrets-gen"))
}
