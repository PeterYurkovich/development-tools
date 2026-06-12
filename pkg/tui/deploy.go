package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type DeploySelectionModel struct {
	BaseModel
	choices  []string
	selected map[int]bool
	cursor   int
	done     bool
}

func NewDeploySelectionModel(choices []string) *DeploySelectionModel {
	return &DeploySelectionModel{
		choices:  choices,
		selected: make(map[int]bool),
		cursor:   0,
		done:     false,
	}
}

func (m DeploySelectionModel) Init() tea.Cmd {
	return nil
}

func (m DeploySelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]

		case "a":
			allSelected := true
			for i := range m.choices {
				if !m.selected[i] {
					allSelected = false
					break
				}
			}

			for i := range m.choices {
				m.selected[i] = !allSelected
			}

		case "enter":
			selectedChoices := []string{}
			for i, choice := range m.choices {
				if m.selected[i] {
					selectedChoices = append(selectedChoices, choice)
				}
			}
			m.SetResult(selectedChoices)
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m DeploySelectionModel) View() string {
	if m.done {
		return ""
	}

	s := TitleStyle.Render("Select components to deploy") + "\n\n"

	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = SelectedStyle.Render("> ")
		}

		checked := "[ ]"
		if m.selected[i] {
			checked = CheckedStyle.Render("[✓]")
		}

		s += fmt.Sprintf("%s%s %s\n", cursor, checked, choice)
	}

	s += "\n" + HelpStyle.Render("space: toggle • a: toggle all • enter: confirm • q: quit")

	return s
}
