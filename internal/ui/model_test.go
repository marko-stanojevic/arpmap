package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marko-stanojevic/hostr/internal/ui"
)

func TestNewModel_InitialState(t *testing.T) {
	t.Parallel()

	m := ui.NewModel()
	view := m.View()

	if view == "" {
		t.Error("expected non-empty view on initial model")
	}
}

func TestModel_QuitOnQ(t *testing.T) {
	t.Parallel()

	m := ui.NewModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	if cmd == nil {
		t.Error("expected a quit command after pressing q")
	}
}

func TestModel_WindowResize(t *testing.T) {
	t.Parallel()

	m := ui.NewModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if updated == nil {
		t.Error("expected updated model after WindowSizeMsg")
	}
}
