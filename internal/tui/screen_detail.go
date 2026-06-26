package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ystsbry/bako/internal/model"
	"github.com/ystsbry/bako/internal/store"
)

// updateDetail handles keys on the project detail screen (PBI / Repo tabs).
func (a *App) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}
	switch key.String() {
	case "esc", "q":
		a.reloadProjects()
		a.screen = screenProjects
		return a, nil
	case "tab", "left", "right", "h", "l":
		a.tab = (a.tab + 1) % 2
		a.clearFeedback()
		return a, nil
	}
	if a.tab == 0 {
		return a.updatePBITab(key)
	}
	return a.updateRepoTab(key)
}

// updatePBITab handles keys while the PBI tab is active.
func (a *App) updatePBITab(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		a.pbiCursor = clamp(a.pbiCursor-1, 0, len(a.pbis)-1)
	case "down", "j":
		a.pbiCursor = clamp(a.pbiCursor+1, 0, len(a.pbis)-1)
	case "n":
		a.openPBIForm("")
		return a, nil
	case "enter", "e":
		if len(a.pbis) > 0 {
			a.openPBIForm(a.pbis[a.pbiCursor].ID)
		}
	case "s":
		if len(a.pbis) > 0 {
			p := a.pbis[a.pbiCursor]
			p.Status = p.Status.Next()
			if err := store.SavePBI(a.proj.Slug, p); err != nil {
				a.fail(err)
			} else {
				a.flash("PBI #%s → %s", p.ID, p.Status.Label())
				a.reloadDetail()
			}
		}
	case "d":
		if len(a.pbis) > 0 {
			p := a.pbis[a.pbiCursor]
			if err := store.DeletePBI(a.proj.Slug, p.ID); err != nil {
				a.fail(err)
			} else {
				a.flash("deleted PBI #%s", p.ID)
				a.reloadDetail()
			}
		}
	}
	return a, nil
}

// updateRepoTab handles keys while the Repo tab is active.
func (a *App) updateRepoTab(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		a.repoCursor = clamp(a.repoCursor-1, 0, len(a.repos)-1)
	case "down", "j":
		a.repoCursor = clamp(a.repoCursor+1, 0, len(a.repos)-1)
	case "n":
		a.openRepoForm(nil)
		return a, nil
	case "enter", "e":
		if len(a.repos) > 0 {
			r := a.repos[a.repoCursor]
			a.openRepoForm(&r)
		}
	case "d":
		if len(a.repos) > 0 {
			r := a.repos[a.repoCursor]
			if err := store.DeleteRepo(a.proj.Slug, r.Slug()); err != nil {
				a.fail(err)
			} else {
				a.flash("deleted %s", r.Display())
				a.reloadDetail()
			}
		}
	}
	return a, nil
}

// viewDetail renders the project detail screen with tabbed PBI/Repo lists.
func (a *App) viewDetail() string {
	var tabs strings.Builder
	pbiTab := "PBI (" + itoa(len(a.pbis)) + ")"
	repoTab := "Repo (" + itoa(len(a.repos)) + ")"
	if a.tab == 0 {
		tabs.WriteString(activeTabStyle.Render(pbiTab))
		tabs.WriteString(inactiveTabStyle.Render(repoTab))
	} else {
		tabs.WriteString(inactiveTabStyle.Render(pbiTab))
		tabs.WriteString(activeTabStyle.Render(repoTab))
	}

	var body strings.Builder
	body.WriteString(tabs.String())
	body.WriteString("\n\n")
	if a.tab == 0 {
		body.WriteString(a.viewPBIList())
	} else {
		body.WriteString(a.viewRepoList())
	}

	var hint string
	if a.tab == 0 {
		hint = "↑/↓ move · tab Repo · n new · e edit · s status · d delete · esc back"
	} else {
		hint = "↑/↓ move · tab PBI · n new · e edit · d delete · esc back"
	}
	help := a.statusLine(hint)
	return a.frame("bako — "+a.proj.Name, body.String(), help)
}

// viewPBIList renders the PBI rows.
func (a *App) viewPBIList() string {
	if len(a.pbis) == 0 {
		return dimStyle.Render("No PBIs yet. Press 'n' to add one (paste content in the form).")
	}
	var b strings.Builder
	for i, p := range a.pbis {
		row := "#" + p.ID + " " + statusBadge(p.Status) + " " + p.Title
		if i == a.pbiCursor {
			b.WriteString(cursorRowStyle.Render("› " + row))
		} else {
			b.WriteString("  " + row)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// viewRepoList renders the repo/org rows.
func (a *App) viewRepoList() string {
	if len(a.repos) == 0 {
		return dimStyle.Render("No repositories yet. Press 'n' to register a repo or org.")
	}
	var b strings.Builder
	for i, r := range a.repos {
		kind := "repo"
		if r.Kind == model.RepoKindOrg {
			kind = "org "
		}
		row := dimStyle.Render("["+kind+"] ") + padRight(r.Display(), 30)
		if r.Overview != "" {
			row += dimStyle.Render("• has overview")
		}
		if i == a.repoCursor {
			b.WriteString(cursorRowStyle.Render("› " + row))
		} else {
			b.WriteString("  " + row)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// itoa is a tiny non-allocating-ish int formatter used in row rendering.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
