package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ystsbry/bako/internal/store"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Inspect bako projects",
	}
	cmd.AddCommand(newProjectListCmd())
	return cmd
}

func newProjectListCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects (slug and name)",
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := store.ListProjects()
			if err != nil {
				return err
			}
			if asJSON {
				type row struct {
					Slug string `json:"slug"`
					Name string `json:"name"`
				}
				out := make([]row, 0, len(projects))
				for _, p := range projects {
					out = append(out, row{Slug: p.Slug, Name: p.Name})
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}
			if len(projects) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "(no projects)")
				return nil
			}
			for _, p := range projects {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", p.Slug, p.Name)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}
