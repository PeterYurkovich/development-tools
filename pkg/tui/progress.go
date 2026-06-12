package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type OperationStatus int

const (
	OperationPending OperationStatus = iota
	OperationInProgress
	OperationComplete
	OperationFailed
)

type Operation struct {
	Name   string
	Status OperationStatus
	Error  error
}

type OperationUpdateMsg struct {
	Index  int
	Status OperationStatus
	Error  error
}

type ProgressModel struct {
	BaseModel
	title      string
	operations []Operation
	current    int
	done       bool
}

func NewProgressModel(title string, operations []string) *ProgressModel {
	ops := make([]Operation, len(operations))
	for i, name := range operations {
		ops[i] = Operation{
			Name:   name,
			Status: OperationPending,
		}
	}

	return &ProgressModel{
		title:      title,
		operations: ops,
		current:    0,
		done:       false,
	}
}

func (m ProgressModel) Init() tea.Cmd {
	return nil
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.done = true
			return m, tea.Quit
		}

	case OperationUpdateMsg:
		if msg.Index >= 0 && msg.Index < len(m.operations) {
			m.operations[msg.Index].Status = msg.Status
			m.operations[msg.Index].Error = msg.Error

			if msg.Status == OperationComplete || msg.Status == OperationFailed {
				m.current = msg.Index + 1
			}

			if msg.Status == OperationFailed {
				m.SetError(msg.Error)
			}

			if m.current >= len(m.operations) {
				m.done = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m ProgressModel) View() string {
	if m.done {
		return ""
	}

	s := TitleStyle.Render(m.title) + "\n\n"

	for _, op := range m.operations {
		var icon string
		var text string

		switch op.Status {
		case OperationPending:
			icon = "○"
			text = op.Name
		case OperationInProgress:
			icon = ProgressStyle.Render("⋯")
			text = ProgressStyle.Render(op.Name)
		case OperationComplete:
			icon = SuccessStyle.Render("✓")
			text = SuccessStyle.Render(op.Name)
		case OperationFailed:
			icon = ErrorStyle.Render("✗")
			text = ErrorStyle.Render(op.Name)
		}

		s += fmt.Sprintf("%s %s", icon, text)

		if op.Status == OperationFailed && op.Error != nil {
			s += " " + ErrorStyle.Render(fmt.Sprintf("- %s", op.Error.Error()))
		}

		s += "\n"
	}

	s += "\n" + HelpStyle.Render("ctrl+c/q: cancel")

	return s
}

func SendOperationUpdate(index int, status OperationStatus, err error) tea.Cmd {
	return func() tea.Msg {
		return OperationUpdateMsg{
			Index:  index,
			Status: status,
			Error:  err,
		}
	}
}
