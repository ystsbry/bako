package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ystsbry/bako/internal/store"
)

// key builds a KeyMsg for a single rune or named key.
func runes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func send(t *testing.T, a *App, msg tea.Msg) *App {
	t.Helper()
	m, _ := a.Update(msg)
	return m.(*App)
}

func typeText(t *testing.T, a *App, s string) *App {
	t.Helper()
	for _, r := range s {
		a = send(t, a, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return a
}

// TestCreateProjectFlow drives the TUI through creating a project, adding a
// PBI and registering a repo, then asserts the store reflects it.
func TestCreateProjectFlow(t *testing.T) {
	t.Setenv("BAKO_HOME", t.TempDir())

	a := &App{screen: screenProjects}
	a = send(t, a, tea.WindowSizeMsg{Width: 100, Height: 40})

	// 'n' opens the new-project form.
	a = send(t, a, runes("n"))
	if a.screen != screenProjectForm {
		t.Fatalf("screen = %v, want project form", a.screen)
	}
	a = typeText(t, a, "Demo")
	a = send(t, a, tea.KeyMsg{Type: tea.KeyCtrlS})
	if a.screen != screenProjects {
		t.Fatalf("after save screen = %v, want projects", a.screen)
	}
	if len(a.projects) != 1 || a.projects[0].Name != "Demo" {
		t.Fatalf("projects = %+v", a.projects)
	}

	// Enter the project detail.
	a = send(t, a, tea.KeyMsg{Type: tea.KeyEnter})
	if a.screen != screenDetail {
		t.Fatalf("screen = %v, want detail", a.screen)
	}

	// Add a PBI: 'n', type title, tab to status, advance, tab to body.
	a = send(t, a, runes("n"))
	if a.screen != screenPBIForm {
		t.Fatalf("screen = %v, want pbi form", a.screen)
	}
	a = typeText(t, a, "Build the thing")
	a = send(t, a, tea.KeyMsg{Type: tea.KeyTab})   // -> status
	a = send(t, a, tea.KeyMsg{Type: tea.KeyRight}) // todo -> progress
	a = send(t, a, tea.KeyMsg{Type: tea.KeyTab})   // -> body
	a = typeText(t, a, "pasted content")
	a = send(t, a, tea.KeyMsg{Type: tea.KeyCtrlS})
	if a.screen != screenDetail {
		t.Fatalf("after pbi save screen = %v", a.screen)
	}

	pbis, _ := store.ListPBIs(a.proj.Slug)
	if len(pbis) != 1 {
		t.Fatalf("pbis = %d, want 1", len(pbis))
	}
	if pbis[0].Title != "Build the thing" || pbis[0].Status.Label() != "Progress" {
		t.Errorf("pbi = %+v", pbis[0])
	}
	if !strings.Contains(pbis[0].Body, "pasted content") {
		t.Errorf("body = %q", pbis[0].Body)
	}

	// Switch to Repo tab and register a repo.
	a = send(t, a, tea.KeyMsg{Type: tea.KeyTab})
	if a.tab != 1 {
		t.Fatalf("tab = %d, want 1 (repo)", a.tab)
	}
	a = send(t, a, runes("n"))
	if a.screen != screenRepoForm {
		t.Fatalf("screen = %v, want repo form", a.screen)
	}
	a = send(t, a, tea.KeyMsg{Type: tea.KeyTab}) // kind -> owner
	a = typeText(t, a, "ystsbry")
	a = send(t, a, tea.KeyMsg{Type: tea.KeyTab}) // owner -> name
	a = typeText(t, a, "bako")
	a = send(t, a, tea.KeyMsg{Type: tea.KeyCtrlS})
	if a.screen != screenDetail {
		t.Fatalf("after repo save screen = %v", a.screen)
	}
	repos, _ := store.ListRepos(a.proj.Slug)
	if len(repos) != 1 || repos[0].Display() != "ystsbry/bako" {
		t.Fatalf("repos = %+v", repos)
	}

	// View() on every screen must not panic and must be non-empty.
	for _, s := range []screen{screenProjects, screenDetail, screenPBIForm, screenRepoForm, screenProjectForm} {
		a.screen = s
		if a.View() == "" {
			t.Errorf("empty view for screen %v", s)
		}
	}
}
