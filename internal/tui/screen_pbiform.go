package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ystsbry/bako/internal/model"
	"github.com/ystsbry/bako/internal/store"
)

// pbiForm holds the inputs for creating or editing a PBI.
type pbiForm struct {
	editingID string // empty for a new PBI
	created   string // preserved when editing
	title     textinput.Model
	status    model.Status
	body      textarea.Model
	focus     int // 0 = title, 1 = status, 2 = body
}

// openPBIForm prepares the PBI form for a new PBI (id == "") or editing one.
func (a *App) openPBIForm(id string) {
	title := textinput.New()
	title.Placeholder = "PBI title"
	title.CharLimit = 200
	title.Prompt = "› "

	body := textarea.New()
	body.Placeholder = "Paste the PBI content here..."
	body.ShowLineNumbers = false
	body.CharLimit = 0
	body.SetWidth(a.width - 4)
	body.SetHeight(a.bodyHeight())

	a.bf = pbiForm{editingID: id, title: title, status: model.StatusTodo, body: body, focus: 0}

	if id != "" {
		if p := a.findPBI(id); p != nil {
			a.bf.title.SetValue(p.Title)
			a.bf.status = p.Status
			a.bf.body.SetValue(p.Body)
			a.bf.created = p.Created
		}
	}
	a.bf.title.Focus()
	a.clearFeedback()
	a.screen = screenPBIForm
}

// findPBI returns a pointer to the loaded PBI with the given ID, or nil.
func (a *App) findPBI(id string) *model.PBI {
	for i := range a.pbis {
		if a.pbis[i].ID == id {
			return &a.pbis[i]
		}
	}
	return nil
}

// bodyHeight returns a sensible textarea height for the current window.
func (a *App) bodyHeight() int {
	h := a.height - 12
	if h < 6 {
		return 6
	}
	if h > 24 {
		return 24
	}
	return h
}

// updatePBIForm handles keys on the PBI create/edit form.
func (a *App) updatePBIForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			a.screen = screenDetail
			return a, nil
		case "tab":
			a.bf.focus = (a.bf.focus + 1) % 3
			a.applyPBIFocus()
			return a, nil
		case "shift+tab":
			a.bf.focus = (a.bf.focus + 2) % 3
			a.applyPBIFocus()
			return a, nil
		case "ctrl+s":
			return a.savePBIForm()
		}
		// On the status field, left/right/space cycle the value.
		if a.bf.focus == 1 {
			switch key.String() {
			case "left", "right", " ", "enter":
				a.bf.status = a.bf.status.Next()
				return a, nil
			}
		}
	}

	var cmd tea.Cmd
	switch a.bf.focus {
	case 0:
		a.bf.title, cmd = a.bf.title.Update(msg)
	case 2:
		a.bf.body, cmd = a.bf.body.Update(msg)
	}
	return a, cmd
}

// applyPBIFocus focuses the active field and blurs the text inputs otherwise.
func (a *App) applyPBIFocus() {
	a.bf.title.Blur()
	a.bf.body.Blur()
	switch a.bf.focus {
	case 0:
		a.bf.title.Focus()
	case 2:
		a.bf.body.Focus()
	}
}

// savePBIForm persists the PBI form, creating or updating as needed.
func (a *App) savePBIForm() (tea.Model, tea.Cmd) {
	title := strings.TrimSpace(a.bf.title.Value())
	bodyText := a.bf.body.Value()
	if a.bf.editingID == "" {
		p, err := store.CreatePBI(a.proj.Slug, title, a.bf.status, bodyText)
		if err != nil {
			a.fail(err)
			return a, nil
		}
		a.flash("created PBI #%s", p.ID)
	} else {
		p := model.PBI{
			ID:      a.bf.editingID,
			Title:   title,
			Status:  a.bf.status,
			Created: a.bf.created,
			Body:    bodyText,
		}
		if err := store.SavePBI(a.proj.Slug, p); err != nil {
			a.fail(err)
			return a, nil
		}
		a.flash("updated PBI #%s", p.ID)
	}
	a.reloadDetail()
	a.screen = screenDetail
	return a, nil
}

// viewPBIForm renders the PBI create/edit form.
func (a *App) viewPBIForm() string {
	title := "bako — new PBI"
	if a.bf.editingID != "" {
		title = "bako — edit PBI #" + a.bf.editingID
	}
	var body strings.Builder

	body.WriteString(a.fieldLabel("Title", a.bf.focus == 0))
	body.WriteString("\n")
	body.WriteString(a.bf.title.View())
	body.WriteString("\n\n")

	body.WriteString(a.fieldLabel("Status", a.bf.focus == 1))
	body.WriteString("\n")
	body.WriteString(statusSelector(a.bf.status))
	body.WriteString("\n\n")

	body.WriteString(a.fieldLabel("Content", a.bf.focus == 2))
	body.WriteString("\n")
	body.WriteString(a.bf.body.View())

	help := a.statusLine("tab next field · status: ←/→ change · ctrl+s save · esc cancel")
	return a.frame(title+"  "+dimStyle.Render(a.proj.Name), body.String(), help)
}

// statusSelector renders the three statuses with the active one highlighted.
func statusSelector(active model.Status) string {
	var parts []string
	for _, s := range model.Statuses() {
		if s == active {
			parts = append(parts, statusStyle(s).Render(" "+s.Label()+" "))
		} else {
			parts = append(parts, inactiveTabStyle.Render(s.Label()))
		}
	}
	return strings.Join(parts, " ")
}

// fieldLabel renders a field label, marking the focused field.
func (a *App) fieldLabel(name string, focused bool) string {
	if focused {
		return fieldLabelStyle.Render("▸ " + name)
	}
	return dimStyle.Render("  " + name)
}
