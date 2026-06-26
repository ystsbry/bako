package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ystsbry/bako/internal/model"
	"github.com/ystsbry/bako/internal/store"
)

func newRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Register and inspect a project's GitHub repositories/organizations",
	}
	cmd.AddCommand(newRepoAddCmd())
	cmd.AddCommand(newRepoListCmd())
	return cmd
}

func newRepoAddCmd() *cobra.Command {
	var (
		project      string
		owner        string
		name         string
		org          bool
		url          string
		overview     string
		overviewFile string
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Register (or update) a repository/organization with an overview",
		Long: `Register a GitHub repository or organization to a project, together with a
free-form overview. Re-running with the same owner/name updates the entry.

The overview text may be supplied via --overview, --overview-file, or piped on
stdin (use --overview-file - to read stdin explicitly). This makes it easy for
an agent to generate a markdown overview and register it in one step:

  bako repo add --project apex-rewrite --owner ystsbry --repo bako \
    --overview-file -  < overview.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := store.ResolveProject(project)
			if err != nil {
				return err
			}
			body, err := resolveOverview(overview, overviewFile)
			if err != nil {
				return err
			}
			kind := model.RepoKindRepo
			if org {
				kind = model.RepoKindOrg
			}
			r := model.Repo{
				Kind:     kind,
				Owner:    owner,
				Name:     name,
				URL:      url,
				Overview: body,
			}
			if err := store.SaveRepo(proj.Slug, r); err != nil {
				return err
			}
			// Re-read to report the canonical (URL-filled) entry.
			r.URL = firstNonEmpty(url, r.DefaultURL())
			fmt.Fprintf(cmd.OutOrStdout(), "registered %s to project %q (%s)\n", r.Display(), proj.Name, r.URL)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&project, "project", "", "project slug or name (required)")
	f.StringVar(&owner, "owner", "", "GitHub owner / login (required)")
	f.StringVar(&name, "repo", "", "repository name (required unless --org)")
	f.BoolVar(&org, "org", false, "register an organization instead of a repository")
	f.StringVar(&url, "url", "", "GitHub URL (defaults to github.com/owner[/repo])")
	f.StringVar(&overview, "overview", "", "overview text")
	f.StringVar(&overviewFile, "overview-file", "", "read overview from a file ('-' for stdin)")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("owner")
	return cmd
}

func newRepoListCmd() *cobra.Command {
	var (
		project string
		asJSON  bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the repositories/organizations of a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			proj, err := store.ResolveProject(project)
			if err != nil {
				return err
			}
			repos, err := store.ListRepos(proj.Slug)
			if err != nil {
				return err
			}
			if asJSON {
				type row struct {
					Kind     string `json:"kind"`
					Owner    string `json:"owner"`
					Name     string `json:"name,omitempty"`
					URL      string `json:"url"`
					Overview string `json:"overview"`
				}
				out := make([]row, 0, len(repos))
				for _, r := range repos {
					out = append(out, row{
						Kind:     string(r.Kind),
						Owner:    r.Owner,
						Name:     r.Name,
						URL:      r.URL,
						Overview: r.Overview,
					})
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}
			if len(repos) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no repositories)")
				return nil
			}
			for _, r := range repos {
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\t%s\n", r.Kind, r.Display(), r.URL)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&project, "project", "", "project slug or name (required)")
	f.BoolVar(&asJSON, "json", false, "output as JSON")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

// resolveOverview returns the overview text from, in priority order: an
// explicit --overview-file (with '-' meaning stdin), --overview, or piped
// stdin when nothing else is given and stdin is not a terminal.
func resolveOverview(overview, overviewFile string) (string, error) {
	if overviewFile == "-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read overview from stdin: %w", err)
		}
		return string(b), nil
	}
	if overviewFile != "" {
		b, err := os.ReadFile(overviewFile)
		if err != nil {
			return "", fmt.Errorf("read overview file: %w", err)
		}
		return string(b), nil
	}
	if overview != "" {
		return overview, nil
	}
	if st, err := os.Stdin.Stat(); err == nil && (st.Mode()&os.ModeCharDevice) == 0 {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read overview from stdin: %w", err)
		}
		return string(b), nil
	}
	return "", nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
