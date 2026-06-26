package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ystsbry/bako/internal/model"
	"github.com/ystsbry/bako/internal/store"
)

// projectForm holds the inputs for creating or editing a project.
type projectForm struct {
	editingSlug string // empty for a new project
	name        textinput.Model
	desc        textarea.Model
	focus       int // 0 = name, 1 = description
}

// updateProjects handles keys on the project list screen.
func (a *App) updateProjects(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}
	switch key.String() {
	case "q", "esc":
		return a, tea.Quit
	case "up", "k":
		a.projCursor = clamp(a.projCursor-1, 0, len(a.projects)-1)
	case "down", "j":
		a.projCursor = clamp(a.projCursor+1, 0, len(a.projects)-1)
	case "n":
		a.openProjectForm("")
		return a, textinput.Blink
	case "e":
		if len(a.projects) > 0 {
			a.openProjectForm(a.projects[a.projCursor].Slug)
			return a, textinput.Blink
		}
	case "d":
		if len(a.projects) > 0 {
			p := a.projects[a.projCursor]
			if err := store.DeleteProject(p.Slug); err != nil {
				a.fail(err)
			} else {
				a.flash("deleted project %q", p.Name)
				a.reloadProjects()
			}
		}
	case "enter":
		if len(a.projects) > 0 {
			a.openDetail(a.projects[a.projCursor])
		}
	}
	return a, nil
}

// openDetail switches to the project detail screen for p.
func (a *App) openDetail(p model.Project) {
	a.proj = p
	a.tab = 0
	a.pbiCursor = 0
	a.repoCursor = 0
	a.clearFeedback()
	a.reloadDetail()
	a.screen = screenDetail
}

// openProjectForm prepares the project form for a new project (slug == "") or
// editing an existing one.
func (a *App) openProjectForm(slug string) {
	name := textinput.New()
	name.Placeholder = "project name"
	name.CharLimit = 120
	name.Prompt = "› "

	desc := textarea.New()
	desc.Placeholder = "description (optional)"
	desc.ShowLineNumbers = false
	desc.SetWidth(a.width - 4)
	desc.SetHeight(6)

	a.pf = projectForm{editingSlug: slug, name: name, desc: desc, focus: 0}

	if slug != "" {
		if p, err := store.LoadProject(slug); err == nil {
			a.pf.name.SetValue(p.Name)
			a.pf.desc.SetValue(p.Description)
		}
	}
	a.pf.name.Focus()
	a.clearFeedback()
	a.screen = screenProjectForm
}

// updateProjectForm handles keys on the project create/edit form.
func (a *App) updateProjectForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			a.screen = screenProjects
			return a, nil
		case "tab", "shift+tab":
			a.pf.focus = (a.pf.focus + 1) % 2
			a.applyProjectFocus()
			return a, nil
		case "ctrl+s":
			return a.saveProjectForm()
		}
	}

	var cmd tea.Cmd
	if a.pf.focus == 0 {
		a.pf.name, cmd = a.pf.name.Update(msg)
	} else {
		a.pf.desc, cmd = a.pf.desc.Update(msg)
	}
	return a, cmd
}

// applyProjectFocus focuses the active field and blurs the other.
func (a *App) applyProjectFocus() {
	if a.pf.focus == 0 {
		a.pf.name.Focus()
		a.pf.desc.Blur()
	} else {
		a.pf.name.Blur()
		a.pf.desc.Focus()
	}
}

// saveProjectForm persists the project form, creating or updating as needed.
func (a *App) saveProjectForm() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(a.pf.name.Value())
	desc := a.pf.desc.Value()
	if a.pf.editingSlug == "" {
		p, err := store.CreateProject(name, desc)
		if err != nil {
			a.fail(err)
			return a, nil
		}
		a.flash("created project %q", p.Name)
	} else {
		p := model.Project{Slug: a.pf.editingSlug, Name: name, Description: desc}
		if err := store.SaveProject(p); err != nil {
			a.fail(err)
			return a, nil
		}
		a.flash("updated project %q", p.Name)
	}
	a.reloadProjects()
	a.screen = screenProjects
	return a, nil
}

// viewProjects renders the project list.
func (a *App) viewProjects() string {
	var body strings.Builder
	if len(a.projects) == 0 {
		body.WriteString(dimStyle.Render("No projects yet. Press 'n' to create one."))
	} else {
		for i, p := range a.projects {
			line := p.Name
			counts := projectCounts(p.Slug)
			row := padRight(line, 28) + dimStyle.Render(counts)
			if i == a.projCursor {
				body.WriteString(cursorRowStyle.Render("› " + row))
			} else {
				body.WriteString("  " + row)
			}
			body.WriteString("\n")
		}
	}
	help := a.statusLine("↑/↓ move · enter open · n new · e edit · d delete · q quit")
	return a.frame("bako — projects", body.String(), help)
}

// projectCounts returns a short "PBI x · Repo y" summary for the list.
func projectCounts(slug string) string {
	pbis, _ := store.ListPBIs(slug)
	repos, _ := store.ListRepos(slug)
	return fmt.Sprintf("PBI %d · Repo %d", len(pbis), len(repos))
}

// viewProjectForm renders the project create/edit form.
func (a *App) viewProjectForm() string {
	title := "bako — new project"
	if a.pf.editingSlug != "" {
		title = "bako — edit project"
	}
	var body strings.Builder
	body.WriteString(fieldLabelStyle.Render("Name"))
	body.WriteString("\n")
	body.WriteString(a.pf.name.View())
	body.WriteString("\n\n")
	body.WriteString(fieldLabelStyle.Render("Description"))
	body.WriteString("\n")
	body.WriteString(a.pf.desc.View())
	help := a.statusLine("tab switch field · ctrl+s save · esc cancel")
	return a.frame(title, body.String(), help)
}
