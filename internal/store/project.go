package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ystsbry/bako/internal/model"
)

// projectMeta is the YAML frontmatter shape of project.md.
type projectMeta struct {
	Name string `yaml:"name"`
}

// ListProjects returns every project under the bako home directory, sorted by
// name (case-insensitive). A missing home directory yields an empty slice.
func ListProjects() ([]model.Project, error) {
	home, err := Home()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(home)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read bako home: %w", err)
	}
	var out []model.Project
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := LoadProject(e.Name())
		if err != nil {
			// Skip directories that are not valid projects.
			continue
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

// LoadProject reads the project identified by slug.
func LoadProject(slug string) (model.Project, error) {
	dir, err := projectDir(slug)
	if err != nil {
		return model.Project{}, err
	}
	data, err := os.ReadFile(filepath.Join(dir, projectFile))
	if err != nil {
		return model.Project{}, fmt.Errorf("read project %q: %w", slug, err)
	}
	var meta projectMeta
	body, err := parseDoc(data, &meta)
	if err != nil {
		return model.Project{}, fmt.Errorf("project %q: %w", slug, err)
	}
	name := strings.TrimSpace(meta.Name)
	if name == "" {
		name = slug
	}
	return model.Project{
		Slug:        slug,
		Name:        name,
		Description: strings.TrimRight(body, "\n"),
	}, nil
}

// ProjectExists reports whether a project directory with a project.md exists
// for the given slug.
func ProjectExists(slug string) bool {
	dir, err := projectDir(slug)
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, projectFile))
	return err == nil
}

// CreateProject creates a new project with a slug derived from name. The slug
// is made unique by appending -2, -3, ... on collision. The created project
// (with its assigned Slug) is returned.
func CreateProject(name, description string) (model.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.Project{}, fmt.Errorf("project name is required")
	}
	slug, err := uniqueProjectSlug(name)
	if err != nil {
		return model.Project{}, err
	}
	p := model.Project{Slug: slug, Name: name, Description: description}
	if err := SaveProject(p); err != nil {
		return model.Project{}, err
	}
	return p, nil
}

// SaveProject writes the project to disk, creating its directory if needed.
// The slug must already be set.
func SaveProject(p model.Project) error {
	if p.Slug == "" {
		return fmt.Errorf("project slug is required")
	}
	dir, err := projectDir(p.Slug)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create project dir: %w", err)
	}
	data, err := renderDoc(projectMeta{Name: p.Name}, p.Description)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, projectFile), data, 0o644); err != nil {
		return fmt.Errorf("write project: %w", err)
	}
	return nil
}

// DeleteProject removes a project and everything under it.
func DeleteProject(slug string) error {
	dir, err := projectDir(slug)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete project %q: %w", slug, err)
	}
	return nil
}

// uniqueProjectSlug returns a slug for name that does not collide with an
// existing project directory.
func uniqueProjectSlug(name string) (string, error) {
	base := Slugify(name)
	if base == "" {
		base = "project"
	}
	home, err := Home()
	if err != nil {
		return "", err
	}
	candidate := base
	for i := 2; ; i++ {
		if _, err := os.Stat(filepath.Join(home, candidate)); os.IsNotExist(err) {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
}
