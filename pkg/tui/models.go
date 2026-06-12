package tui

import tea "github.com/charmbracelet/bubbletea"

type Model interface {
	tea.Model
	Error() error
	Result() interface{}
}

type BaseModel struct {
	err    error
	result interface{}
}

func (m *BaseModel) Error() error {
	return m.err
}

func (m *BaseModel) Result() interface{} {
	return m.result
}

func (m *BaseModel) SetError(err error) {
	m.err = err
}

func (m *BaseModel) SetResult(result interface{}) {
	m.result = result
}
