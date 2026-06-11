package display

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#B39DDB"))
	LabelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	SecretStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#C0C0C0")).Bold(true)
	CriticalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
	HighStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB347")).Bold(true)
	MediumStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFE066"))
	LowStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#C0C0C0"))
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
