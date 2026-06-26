package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ystsbry/bako/internal/model"
)

// repoMeta is the YAML frontmatter shape of a repo/*.md file.
type repoMeta struct {
	Kind  string `yaml:"kind"`
	Owner string `yaml:"owner"`
	Name  string `yaml:"name,omitempty"`
	URL   string `yaml:"url"`
}

// ListRepos returns the repositories/organizations registered to a project,
// sorted by display name.
func ListRepos(slug string) ([]model.Repo, error) {
	dir, err := repoDir(slug)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read repo dir: %w", err)
	}
	var out []model.Repo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		r, err := loadRepoFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Display()) < strings.ToLower(out[j].Display())
	})
	return out, nil
}

// SaveRepo validates and writes a repo/org entry to the project. A blank URL
// is filled with the canonical github.com URL. The entry is keyed by its
// Slug(), so re-saving the same owner/name overwrites the existing file.
func SaveRepo(slug string, r model.Repo) error {
	r.Owner = strings.TrimSpace(r.Owner)
	r.Name = strings.TrimSpace(r.Name)
	r.URL = strings.TrimSpace(r.URL)
	if !r.Kind.Valid() {
		return fmt.Errorf("invalid repo kind %q", r.Kind)
	}
	if r.Owner == "" {
		return fmt.Errorf("owner is required")
	}
	if r.Kind == model.RepoKindRepo && r.Name == "" {
		return fmt.Errorf("repository name is required")
	}
	if r.Kind == model.RepoKindOrg {
		r.Name = ""
	}
	if r.URL == "" {
		r.URL = r.DefaultURL()
	}
	dir, err := repoDir(slug)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create repo dir: %w", err)
	}
	data, err := renderDoc(repoMeta{
		Kind:  string(r.Kind),
		Owner: r.Owner,
		Name:  r.Name,
		URL:   r.URL,
	}, r.Overview)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, r.Slug()+".md"), data, 0o644); err != nil {
		return fmt.Errorf("write repo: %w", err)
	}
	return nil
}

// DeleteRepo removes the repo/org entry identified by its slug (as returned by
// model.Repo.Slug).
func DeleteRepo(slug, repoSlug string) error {
	dir, err := repoDir(slug)
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(dir, repoSlug+".md")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete repo %q: %w", repoSlug, err)
	}
	return nil
}

// loadRepoFile reads and parses a single repo/*.md file.
func loadRepoFile(path string) (model.Repo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Repo{}, err
	}
	var meta repoMeta
	body, err := parseDoc(data, &meta)
	if err != nil {
		return model.Repo{}, err
	}
	kind := model.RepoKind(meta.Kind)
	if !kind.Valid() {
		kind = model.RepoKindRepo
	}
	r := model.Repo{
		Kind:     kind,
		Owner:    strings.TrimSpace(meta.Owner),
		Name:     strings.TrimSpace(meta.Name),
		URL:      strings.TrimSpace(meta.URL),
		Overview: strings.TrimRight(body, "\n"),
	}
	if r.URL == "" {
		r.URL = r.DefaultURL()
	}
	return r, nil
}
