package ui

import (
	"fmt"
	"strings"

	"github.com/bastiblast/boiler-deploy/internal/ansible"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TagSelector struct {
	categories   []ansible.TagCategory
	actionType   string // "provision" ou "deploy"
	focusIndex   int
	categoryIdx  int
	tagIdx       int
	confirmed    bool
	cancelled    bool
	width        int
	height       int
}

func NewTagSelector(actionType string) TagSelector {
	return NewTagSelectorWithDefaults(actionType, nil)
}

func NewTagSelectorWithDefaults(actionType string, defaultTags []string) TagSelector {
	var categories []ansible.TagCategory
	if actionType == "provision" {
		categories = ansible.GetProvisionTags()
	} else {
		categories = ansible.GetDeployTags()
	}

	// Pre-select tags based on defaults
	if defaultTags != nil && len(defaultTags) > 0 {
		for i := range categories {
			for j := range categories[i].Tags {
				tagName := categories[i].Tags[j].Name
				for _, dt := range defaultTags {
					if dt == tagName || dt == "all" {
						categories[i].Tags[j].Selected = true
						break
					}
				}
			}
		}
	}

	return TagSelector{
		categories:  categories,
		actionType:  actionType,
		focusIndex:  0,
		categoryIdx: 0,
		tagIdx:      0,
		confirmed:   false,
		cancelled:   false,
	}
}

func (m TagSelector) Init() tea.Cmd {
	return nil
}

func (m TagSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit

		case "esc":
			m.cancelled = true
			return m, nil

		case "enter":
			m.confirmed = true
			return m, nil

		case "up", "k":
			m.navigateUp()
			return m, nil

		case "down", "j":
			m.navigateDown()
			return m, nil

		case " ":
			m.toggleTag()
			return m, nil

		case "a":
			m.selectAll()
			return m, nil

		case "n":
			m.selectNone()
			return m, nil
		}
	}

	return m, nil
}

func (m TagSelector) View() string {
	// Use a minimum width if not set yet
	if m.width == 0 {
		m.width = 80 // Fallback to reasonable default
	}

	var b strings.Builder

	// Header
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Padding(0, 1)

	title := fmt.Sprintf("Select Tags for %s", strings.ToUpper(m.actionType))
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Instructions
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 2)

	help := "↑/↓: Navigate • Space: Toggle • a: Select All • n: Select None • Enter: Confirm • Esc: Cancel"
	b.WriteString(helpStyle.Render(help))
	b.WriteString("\n\n")

	// Categories and tags
	currentIndex := 0
	for catIdx, category := range m.categories {
		// Category header
		catStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33")).
			Padding(0, 2)

		b.WriteString(catStyle.Render(fmt.Sprintf("▸ %s", category.Name)))
		b.WriteString("\n")

		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 4)

		b.WriteString(descStyle.Render(category.Description))
		b.WriteString("\n\n")

		// Tags
		for _, tag := range category.Tags {
			cursor := "  "
			if currentIndex == m.focusIndex {
				cursor = "▶ "
			}

			checkbox := "☐"
			if tag.Selected {
				checkbox = "☑"
			}

			tagStyle := lipgloss.NewStyle().Padding(0, 4)
			if currentIndex == m.focusIndex {
				tagStyle = tagStyle.
					Background(lipgloss.Color("240")).
					Foreground(lipgloss.Color("230"))
			}

			line := fmt.Sprintf("%s%s %s - %s", cursor, checkbox, tag.Name, tag.Description)
			b.WriteString(tagStyle.Render(line))
			b.WriteString("\n")

			currentIndex++
		}
		b.WriteString("\n")

		if catIdx < len(m.categories)-1 {
			b.WriteString("\n")
		}
	}

	// Footer with selected tags summary
	selectedTags := ansible.GetAllTags(m.categories)
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33")).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	footer := fmt.Sprintf("Selected: %d tags", len(selectedTags))
	if len(selectedTags) > 0 {
		footer += fmt.Sprintf(" (%s)", strings.Join(selectedTags, ", "))
	}
	b.WriteString("\n")
	b.WriteString(footerStyle.Render(footer))

	return b.String()
}

func (m *TagSelector) navigateUp() {
	if m.focusIndex > 0 {
		m.focusIndex--
		m.updateIndices()
	}
}

func (m *TagSelector) navigateDown() {
	maxIndex := m.getTotalTags() - 1
	if m.focusIndex < maxIndex {
		m.focusIndex++
		m.updateIndices()
	}
}

func (m *TagSelector) toggleTag() {
	catIdx, tagIdx := m.getCurrentPosition()
	if catIdx >= 0 && tagIdx >= 0 {
		m.categories[catIdx].Tags[tagIdx].Selected = !m.categories[catIdx].Tags[tagIdx].Selected
	}
}

func (m *TagSelector) selectAll() {
	for i := range m.categories {
		for j := range m.categories[i].Tags {
			m.categories[i].Tags[j].Selected = true
		}
	}
}

func (m *TagSelector) selectNone() {
	for i := range m.categories {
		for j := range m.categories[i].Tags {
			m.categories[i].Tags[j].Selected = false
		}
	}
}

func (m *TagSelector) getTotalTags() int {
	total := 0
	for _, cat := range m.categories {
		total += len(cat.Tags)
	}
	return total
}

func (m *TagSelector) getCurrentPosition() (catIdx, tagIdx int) {
	currentIndex := 0
	for ci, cat := range m.categories {
		for ti := range cat.Tags {
			if currentIndex == m.focusIndex {
				return ci, ti
			}
			currentIndex++
		}
	}
	return -1, -1
}

func (m *TagSelector) updateIndices() {
	catIdx, _ := m.getCurrentPosition()
	m.categoryIdx = catIdx
}

func (m TagSelector) GetSelectedTags() []string {
	return ansible.GetAllTags(m.categories)
}

func (m TagSelector) GetTagString() string {
	return ansible.FormatTagsForAnsible(m.categories)
}

func (m TagSelector) IsConfirmed() bool {
	return m.confirmed
}

func (m TagSelector) IsCancelled() bool {
	return m.cancelled
}
