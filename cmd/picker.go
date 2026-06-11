package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gray-snyk/secrets-gen/internal/display"
)

// providerItem is a single selectable provider in the picker list.
type providerItem string

func (p providerItem) FilterValue() string { return string(p) }

// providerDelegate renders provider items as a compact single-line list.
type providerDelegate struct{}

func (d providerDelegate) Height() int                             { return 1 }
func (d providerDelegate) Spacing() int                            { return 0 }
func (d providerDelegate) Update(tea.Msg, *list.Model) tea.Cmd     { return nil }
func (d providerDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	p, ok := item.(providerItem)
	if !ok {
		return
	}
	line := "  " + string(p)
	if index == m.Index() {
		line = display.SecretStyle.Render("> " + string(p))
	} else {
		line = display.LabelStyle.Render(line)
	}
	fmt.Fprint(w, line)
}

// pickerModel is the Bubble Tea model for the interactive provider picker.
type pickerModel struct {
	list     list.Model
	chosen   string
	quitting bool
}

func newPickerModel(providers []string) pickerModel {
	items := make([]list.Item, len(providers))
	for i, p := range providers {
		items[i] = providerItem(p)
	}

	l := list.New(items, providerDelegate{}, 0, 0)
	l.Title = "Select a provider"
	l.Styles.Title = display.TitleStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	return pickerModel{list: l}
}

func (m pickerModel) Init() tea.Cmd { return nil }

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-1)
		return m, nil
	case tea.KeyMsg:
		// Don't intercept keys while the user is typing a filter.
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "enter":
				if item, ok := m.list.SelectedItem().(providerItem); ok {
					m.chosen = string(item)
				}
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pickerModel) View() string {
	if m.quitting || m.chosen != "" {
		return ""
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(m.list.View())
}

// runPicker launches the interactive provider picker and returns the chosen
// provider. The second return value is false if the user quit without
// selecting anything.
func runPicker(providers []string) (string, bool, error) {
	p := tea.NewProgram(newPickerModel(providers), tea.WithAltScreen())
	res, err := p.Run()
	if err != nil {
		return "", false, err
	}
	m, ok := res.(pickerModel)
	if !ok || m.chosen == "" {
		return "", false, nil
	}
	return strings.TrimSpace(m.chosen), true, nil
}
