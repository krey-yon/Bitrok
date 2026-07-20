package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDashboardRefreshKey(t *testing.T) {
	called := false
	m := NewDashboard("https://example.bitrok.tech", "localhost:3000")
	m.Refresh = func() tea.Msg {
		called = true
		return nil
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("refresh key did not return the refresh command")
	}
	if called {
		t.Fatal("refresh command ran during Update")
	}
	cmd()
	if !called {
		t.Fatal("refresh command was not invoked")
	}
	if updated.(DashboardModel).Quitting {
		t.Fatal("refresh key quit the dashboard")
	}
}

func TestDashboardViewIncludesRefreshHint(t *testing.T) {
	m := NewDashboard("https://example.bitrok.tech", "localhost:3000")
	view := m.View()
	if !strings.Contains(view, "[r] refresh") {
		t.Fatal("dashboard view does not include the refresh key hint")
	}
}
