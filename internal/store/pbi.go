package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ystsbry/bako/internal/model"
)

// pbiMeta is the YAML frontmatter shape of a pbi/*.md file.
type pbiMeta struct {
	Title   string `yaml:"title"`
	Status  string `yaml:"status"`
	Created string `yaml:"created"`
}

// pbiIDWidth is the zero-padding width of PBI sequence numbers.
const pbiIDWidth = 3

// ListPBIs returns the PBIs of a project sorted by ID ascending. A project
// with no pbi/ directory yields an empty slice.
func ListPBIs(slug string) ([]model.PBI, error) {
	dir, err := pbiDir(slug)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read pbi dir: %w", err)
	}
	var out []model.PBI
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		p, err := loadPBIFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// CreatePBI assigns the next sequence ID and writes a new PBI to the project.
// When status is invalid it defaults to Todo; when created is empty today's
// date is used. The stored PBI (with ID populated) is returned.
func CreatePBI(slug, title string, status model.Status, body string) (model.PBI, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model.PBI{}, fmt.Errorf("PBI title is required")
	}
	if !status.Valid() {
		status = model.StatusTodo
	}
	id, err := nextPBIID(slug)
	if err != nil {
		return model.PBI{}, err
	}
	p := model.PBI{
		ID:      id,
		Title:   title,
		Status:  status,
		Created: time.Now().Format("2006-01-02"),
		Body:    body,
	}
	if err := SavePBI(slug, p); err != nil {
		return model.PBI{}, err
	}
	return p, nil
}

// SavePBI writes a PBI to the project's pbi/ directory. The PBI's ID must be
// set. The on-disk filename is {id}-{slug(title)}.md; saving re-derives the
// filename, removing any stale file that shared the same ID but a different
// title slug.
func SavePBI(slug string, p model.PBI) error {
	if p.ID == "" {
		return fmt.Errorf("PBI ID is required")
	}
	if !p.Status.Valid() {
		return fmt.Errorf("invalid PBI status %q", p.Status)
	}
	dir, err := pbiDir(slug)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create pbi dir: %w", err)
	}
	if err := removePBIFilesForID(dir, p.ID); err != nil {
		return err
	}
	data, err := renderDoc(pbiMeta{
		Title:   p.Title,
		Status:  string(p.Status),
		Created: p.Created,
	}, p.Body)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, pbiFilename(p)), data, 0o644); err != nil {
		return fmt.Errorf("write pbi: %w", err)
	}
	return nil
}

// DeletePBI removes the PBI with the given ID from the project.
func DeletePBI(slug, id string) error {
	dir, err := pbiDir(slug)
	if err != nil {
		return err
	}
	return removePBIFilesForID(dir, id)
}

// pbiFilename returns the on-disk filename for a PBI.
func pbiFilename(p model.PBI) string {
	s := Slugify(p.Title)
	if s == "" {
		return p.ID + ".md"
	}
	return p.ID + "-" + s + ".md"
}

// loadPBIFile reads and parses a single pbi/*.md file, deriving the ID from
// the filename prefix.
func loadPBIFile(path string) (model.PBI, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.PBI{}, err
	}
	var meta pbiMeta
	body, err := parseDoc(data, &meta)
	if err != nil {
		return model.PBI{}, err
	}
	status := model.Status(meta.Status)
	if !status.Valid() {
		status = model.StatusTodo
	}
	return model.PBI{
		ID:      idFromFilename(filepath.Base(path)),
		Title:   strings.TrimSpace(meta.Title),
		Status:  status,
		Created: meta.Created,
		Body:    strings.TrimRight(body, "\n"),
	}, nil
}

// idFromFilename extracts the leading numeric ID from a filename like
// "003-login.md" or "003.md".
func idFromFilename(name string) string {
	name = strings.TrimSuffix(name, ".md")
	if i := strings.IndexByte(name, '-'); i >= 0 {
		return name[:i]
	}
	return name
}

// nextPBIID returns the next zero-padded sequence ID for a project.
func nextPBIID(slug string) (string, error) {
	pbis, err := ListPBIs(slug)
	if err != nil {
		return "", err
	}
	max := 0
	for _, p := range pbis {
		if n, err := strconv.Atoi(p.ID); err == nil && n > max {
			max = n
		}
	}
	return fmt.Sprintf("%0*d", pbiIDWidth, max+1), nil
}

// removePBIFilesForID deletes any pbi/*.md whose leading ID matches id.
func removePBIFilesForID(dir, id string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read pbi dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if idFromFilename(e.Name()) == id {
			if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
				return fmt.Errorf("remove pbi file: %w", err)
			}
		}
	}
	return nil
}
