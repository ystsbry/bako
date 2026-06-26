// Command bako is a local-first project tracker. Each project bundles its
// PBIs (Product Backlog Items) with the GitHub repositories or organizations
// relevant to it, all stored as markdown files under ~/.bako. Running bako
// with no arguments opens the interactive TUI for registering content.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ystsbry/bako/internal/store"
	"github.com/ystsbry/bako/internal/tui"
)

var (
	version = "0.1.0-dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "bako",
		Short:         "Local-first tracker for project PBIs and GitHub repositories",
		SilenceUsage:  true,
		SilenceErrors: false,
		// With no subcommand, launch the TUI.
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run()
		},
	}
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newHomeCmd())
	cmd.AddCommand(newProjectCmd())
	cmd.AddCommand(newRepoCmd())
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print bako version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "bako %s (commit %s, built %s)\n", version, commit, date)
			return nil
		},
	}
}

func newHomeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "home",
		Short: "Print the bako storage directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := store.Home()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), home)
			return nil
		},
	}
}
