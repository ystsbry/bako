package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ystsbry/bako/internal/model"
	"github.com/ystsbry/bako/internal/store"
)

// repoForm field indices.
const (
	rfKind = iota
	rfOwner
	rfName
	rfURL
	rfOverview
	rfFieldCount
)

// repoForm holds the inputs for registering or editing a repo/org.
type repoForm struct {
	editingSlug string // empty for a new entry
	kind        model.RepoKind
	owner       textinput.Model
	name        textinput.Model
	url         textinput.Model
	overview    textarea.Model
	focus       int
}

// openRepoForm prepares the repo form. existing == nil starts a new entry.
func (a *App) openRepoForm(existing *model.Repo) {
	mk := func(ph string) textinput.Model {
		t := textinput.New()
		t.Placeholder = ph
		t.CharLimit = 200
		t.Prompt = "› "
		return t
	}
	overview := textarea.New()
	overview.Placeholder = "Repository / organization overview (optional)..."
	overview.ShowLineNumbers = false
	overview.SetWidth(a.width - 4)
	overview.SetHeight(a.bodyHeight())

	a.rf = repoForm{
		kind:     model.RepoKindRepo,
		owner:    mk("owner (e.g. ystsbry)"),
		name:     mk("repository (e.g. bako)"),
		url:      mk("https://github.com/... (auto if blank)"),
		overview: overview,
		focus:    rfKind,
	}
	if existing != nil {
		a.rf.editingSlug = existing.Slug()
		a.rf.kind = existing.Kind
		a.rf.owner.SetValue(existing.Owner)
		a.rf.name.SetValue(existing.Name)
		a.rf.url.SetValue(existing.URL)
		a.rf.overview.SetValue(existing.Overview)
	}
	a.applyRepoFocus()
	a.clearFeedback()
	a.screen = screenRepoForm
}

// updateRepoForm handles keys on the repo create/edit form.
func (a *App) updateRepoForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			a.screen = screenDetail
			return a, nil
		case "tab":
			a.rf.focus = (a.rf.focus + 1) % rfFieldCount
			a.applyRepoFocus()
			return a, nil
		case "shift+tab":
			a.rf.focus = (a.rf.focus + rfFieldCount - 1) % rfFieldCount
			a.applyRepoFocus()
			return a, nil
		case "ctrl+s":
			return a.saveRepoForm()
		}
		if a.rf.focus == rfKind {
			switch key.String() {
			case "left", "right", " ", "enter":
				a.toggleRepoKind()
				return a, nil
			}
		}
	}

	var cmd tea.Cmd
	switch a.rf.focus {
	case rfOwner:
		a.rf.owner, cmd = a.rf.owner.Update(msg)
	case rfName:
		a.rf.name, cmd = a.rf.name.Update(msg)
	case rfURL:
		a.rf.url, cmd = a.rf.url.Update(msg)
	case rfOverview:
		a.rf.overview, cmd = a.rf.overview.Update(msg)
	}
	return a, cmd
}

// toggleRepoKind flips between repo and org.
func (a *App) toggleRepoKind() {
	if a.rf.kind == model.RepoKindRepo {
		a.rf.kind = model.RepoKindOrg
	} else {
		a.rf.kind = model.RepoKindRepo
	}
}

// applyRepoFocus focuses the active field and blurs all text inputs otherwise.
func (a *App) applyRepoFocus() {
	a.rf.owner.Blur()
	a.rf.name.Blur()
	a.rf.url.Blur()
	a.rf.overview.Blur()
	switch a.rf.focus {
	case rfOwner:
		a.rf.owner.Focus()
	case rfName:
		a.rf.name.Focus()
	case rfURL:
		a.rf.url.Focus()
	case rfOverview:
		a.rf.overview.Focus()
	}
}

// saveRepoForm persists the repo form. When editing and the identity (slug)
// changed, the old file is removed so no stale entry lingers.
func (a *App) saveRepoForm() (tea.Model, tea.Cmd) {
	r := model.Repo{
		Kind:     a.rf.kind,
		Owner:    strings.TrimSpace(a.rf.owner.Value()),
		Name:     strings.TrimSpace(a.rf.name.Value()),
		URL:      strings.TrimSpace(a.rf.url.Value()),
		Overview: a.rf.overview.Value(),
	}
	if err := store.SaveRepo(a.proj.Slug, r); err != nil {
		a.fail(err)
		return a, nil
	}
	if a.rf.editingSlug != "" && a.rf.editingSlug != r.Slug() {
		_ = store.DeleteRepo(a.proj.Slug, a.rf.editingSlug)
	}
	a.flash("saved %s", r.Display())
	a.reloadDetail()
	a.screen = screenDetail
	return a, nil
}

// viewRepoForm renders the repo create/edit form.
func (a *App) viewRepoForm() string {
	title := "bako — register repo"
	if a.rf.editingSlug != "" {
		title = "bako — edit repo"
	}
	var body strings.Builder

	body.WriteString(a.fieldLabel("Kind", a.rf.focus == rfKind))
	body.WriteString("\n")
	body.WriteString(repoKindSelector(a.rf.kind))
	body.WriteString("\n\n")

	body.WriteString(a.fieldLabel("Owner", a.rf.focus == rfOwner))
	body.WriteString("\n")
	body.WriteString(a.rf.owner.View())
	body.WriteString("\n\n")

	if a.rf.kind == model.RepoKindRepo {
		body.WriteString(a.fieldLabel("Repository", a.rf.focus == rfName))
		body.WriteString("\n")
		body.WriteString(a.rf.name.View())
		body.WriteString("\n\n")
	} else {
		body.WriteString(dimStyle.Render("  Repository (n/a for org)"))
		body.WriteString("\n\n")
	}

	body.WriteString(a.fieldLabel("URL", a.rf.focus == rfURL))
	body.WriteString("\n")
	body.WriteString(a.rf.url.View())
	body.WriteString("\n\n")

	body.WriteString(a.fieldLabel("Overview", a.rf.focus == rfOverview))
	body.WriteString("\n")
	body.WriteString(a.rf.overview.View())

	help := a.statusLine("tab next field · kind: ←/→ change · ctrl+s save · esc cancel")
	return a.frame(title+"  "+dimStyle.Render(a.proj.Name), body.String(), help)
}

// repoKindSelector renders the repo/org toggle.
func repoKindSelector(active model.RepoKind) string {
	render := func(k model.RepoKind, label string) string {
		if k == active {
			return activeTabStyle.Render(label)
		}
		return inactiveTabStyle.Render(label)
	}
	return render(model.RepoKindRepo, "repository") + " " + render(model.RepoKindOrg, "organization")
}
