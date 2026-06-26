// Package tui implements bako's interactive terminal UI for registering
// projects, PBIs and GitHub repositories. It is built on Bubble Tea and
// persists everything through internal/store.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ystsbry/bako/internal/model"
	"github.com/ystsbry/bako/internal/store"
)

// screen identifies which view is currently active.
type screen int

const (
	screenProjects screen = iota
	screenProjectForm
	screenDetail
	screenPBIForm
	screenRepoForm
)

// App is the root Bubble Tea model.
type App struct {
	screen        screen
	width, height int

	// transient feedback shown on the status line.
	status string
	errMsg string

	// project list state.
	projects   []model.Project
	projCursor int

	// active project context (valid on detail and form screens).
	proj       model.Project
	tab        int // 0 = PBI, 1 = Repo
	pbis       []model.PBI
	pbiCursor  int
	repos      []model.Repo
	repoCursor int

	// forms.
	pf projectForm
	bf pbiForm
	rf repoForm
}

// Run starts the bako TUI and blocks until the user quits.
func Run() error {
	projects, err := store.ListProjects()
	if err != nil {
		return err
	}
	app := &App{screen: screenProjects, projects: projects}
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = m.Width, m.Height
		// Reflow only the textarea of the currently open form; the others
		// are zero-value until their form is opened (where width is set).
		switch a.screen {
		case screenProjectForm:
			a.pf.desc.SetWidth(m.Width - 4)
		case screenPBIForm:
			a.bf.body.SetWidth(m.Width - 4)
		case screenRepoForm:
			a.rf.overview.SetWidth(m.Width - 4)
		}
		return a, nil
	case tea.KeyMsg:
		// Ctrl+C always quits, regardless of screen.
		if m.String() == "ctrl+c" {
			return a, tea.Quit
		}
	}

	switch a.screen {
	case screenProjects:
		return a.updateProjects(msg)
	case screenProjectForm:
		return a.updateProjectForm(msg)
	case screenDetail:
		return a.updateDetail(msg)
	case screenPBIForm:
		return a.updatePBIForm(msg)
	case screenRepoForm:
		return a.updateRepoForm(msg)
	}
	return a, nil
}

// View implements tea.Model.
func (a *App) View() string {
	switch a.screen {
	case screenProjects:
		return a.viewProjects()
	case screenProjectForm:
		return a.viewProjectForm()
	case screenDetail:
		return a.viewDetail()
	case screenPBIForm:
		return a.viewPBIForm()
	case screenRepoForm:
		return a.viewRepoForm()
	}
	return ""
}

// --- shared helpers ---

// flash sets a transient success message and clears any error.
func (a *App) flash(format string, args ...any) {
	a.status = fmt.Sprintf(format, args...)
	a.errMsg = ""
}

// fail sets an error message and clears any success message.
func (a *App) fail(err error) {
	a.errMsg = err.Error()
	a.status = ""
}

// clearFeedback drops any transient status/error line.
func (a *App) clearFeedback() {
	a.status = ""
	a.errMsg = ""
}

// reloadProjects refreshes the project list, clamping the cursor.
func (a *App) reloadProjects() {
	projects, err := store.ListProjects()
	if err != nil {
		a.fail(err)
		return
	}
	a.projects = projects
	a.projCursor = clamp(a.projCursor, 0, len(projects)-1)
}

// reloadDetail refreshes the PBI and repo lists for the active project.
func (a *App) reloadDetail() {
	pbis, err := store.ListPBIs(a.proj.Slug)
	if err != nil {
		a.fail(err)
		return
	}
	repos, err := store.ListRepos(a.proj.Slug)
	if err != nil {
		a.fail(err)
		return
	}
	a.pbis = pbis
	a.repos = repos
	a.pbiCursor = clamp(a.pbiCursor, 0, len(pbis)-1)
	a.repoCursor = clamp(a.repoCursor, 0, len(repos)-1)
}

// statusLine renders the transient feedback line, or the given hint when empty.
func (a *App) statusLine(hint string) string {
	switch {
	case a.errMsg != "":
		return errStyle.Render("✗ " + a.errMsg)
	case a.status != "":
		return okStyle.Render("✓ " + a.status)
	default:
		return helpStyle.Render(hint)
	}
}

// frame composes a title bar, body and help footer into a full-screen view.
func (a *App) frame(title, body, help string) string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(title))
	b.WriteString("\n\n")
	b.WriteString(body)
	b.WriteString("\n\n")
	b.WriteString(help)
	return b.String()
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
