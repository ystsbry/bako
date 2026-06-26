// Package model holds bako's core domain types: projects, PBIs (Product
// Backlog Items) and the GitHub repositories/organizations attached to a
// project. These types are persisted as markdown files by internal/store.
package model

import "fmt"

// Status is the lifecycle state of a PBI. bako tracks exactly three states.
type Status string

const (
	StatusTodo     Status = "todo"
	StatusProgress Status = "progress"
	StatusDone     Status = "done"
)

// Statuses lists the three valid states in display order. The order is also
// the cycle order used by the TUI when the user advances a PBI's status.
func Statuses() []Status {
	return []Status{StatusTodo, StatusProgress, StatusDone}
}

// Valid reports whether s is one of the three known states.
func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusProgress, StatusDone:
		return true
	default:
		return false
	}
}

// Label returns a short human-facing label for the status.
func (s Status) Label() string {
	switch s {
	case StatusTodo:
		return "Todo"
	case StatusProgress:
		return "Progress"
	case StatusDone:
		return "Done"
	default:
		return string(s)
	}
}

// Next returns the following status in cycle order (Done wraps to Todo).
// Used by the TUI to advance a PBI through its lifecycle with a single key.
func (s Status) Next() Status {
	switch s {
	case StatusTodo:
		return StatusProgress
	case StatusProgress:
		return StatusDone
	default:
		return StatusTodo
	}
}

// Project groups a set of PBIs together with the GitHub repositories or
// organizations relevant to that work.
type Project struct {
	// Slug is the on-disk directory name; it is derived from Name and is
	// stable once the project is created. It is not stored in the file.
	Slug string
	// Name is the human-facing project name.
	Name string
	// Description is free-form markdown describing the project.
	Description string
}

// PBI is a single Product Backlog Item belonging to a project.
type PBI struct {
	// ID is a zero-padded sequence number unique within the project; it is
	// also the leading part of the on-disk filename.
	ID string
	// Title is the one-line summary shown in lists.
	Title string
	// Status is the current lifecycle state.
	Status Status
	// Created is the creation date (YYYY-MM-DD).
	Created string
	// Body is the full PBI content (markdown), typically pasted in via the TUI.
	Body string
}

// RepoKind distinguishes a single repository entry from an organization entry.
type RepoKind string

const (
	RepoKindRepo RepoKind = "repo"
	RepoKindOrg  RepoKind = "org"
)

// Valid reports whether k is one of the known kinds.
func (k RepoKind) Valid() bool {
	return k == RepoKindRepo || k == RepoKindOrg
}

// Repo is a GitHub repository or organization registered to a project,
// together with a free-form overview.
type Repo struct {
	// Kind is "repo" or "org".
	Kind RepoKind
	// Owner is the GitHub owner (user or organization login).
	Owner string
	// Name is the repository name. Empty when Kind is org.
	Name string
	// URL is the GitHub URL. Derived when blank.
	URL string
	// Overview is free-form markdown describing the repo/org.
	Overview string
}

// Slug returns a stable identifier for the repo entry, used as the on-disk
// filename stem and for de-duplication. Orgs are keyed by owner alone.
func (r Repo) Slug() string {
	if r.Kind == RepoKindOrg {
		return r.Owner
	}
	return r.Owner + "__" + r.Name
}

// DefaultURL returns the canonical github.com URL for the entry.
func (r Repo) DefaultURL() string {
	if r.Kind == RepoKindOrg {
		return "https://github.com/" + r.Owner
	}
	return fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Name)
}

// Display returns "owner/name" for a repo, or "owner (org)" for an org.
func (r Repo) Display() string {
	if r.Kind == RepoKindOrg {
		return r.Owner + " (org)"
	}
	return r.Owner + "/" + r.Name
}
