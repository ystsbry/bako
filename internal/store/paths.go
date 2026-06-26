// Package store persists bako projects, PBIs and repositories as markdown
// files with YAML frontmatter under the bako home directory.
//
// Layout:
//
//	~/.bako/
//	  {project-slug}/
//	    project.md            # name in frontmatter, description in body
//	    pbi/
//	      {id}-{slug}.md       # title/status/created in frontmatter, content in body
//	    repo/
//	      {owner}__{name}.md   # kind/owner/name/url in frontmatter, overview in body
//
// The BAKO_HOME environment variable overrides the default ~/.bako root
// (used by tests and for relocating storage).
package store

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	projectFile = "project.md"
	pbiDirName  = "pbi"
	repoDirName = "repo"
)

// Home returns the bako home directory. Defaults to ~/.bako; the BAKO_HOME
// env var overrides this.
func Home() (string, error) {
	if v := os.Getenv("BAKO_HOME"); v != "" {
		return v, nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(h, ".bako"), nil
}

// projectDir returns the directory for a project slug.
func projectDir(slug string) (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	if slug == "" {
		return "", fmt.Errorf("empty project slug")
	}
	return filepath.Join(home, slug), nil
}

// pbiDir returns the pbi/ subdirectory for a project.
func pbiDir(slug string) (string, error) {
	d, err := projectDir(slug)
	if err != nil {
		return "", err
	}
	return filepath.Join(d, pbiDirName), nil
}

// repoDir returns the repo/ subdirectory for a project.
func repoDir(slug string) (string, error) {
	d, err := projectDir(slug)
	if err != nil {
		return "", err
	}
	return filepath.Join(d, repoDirName), nil
}
