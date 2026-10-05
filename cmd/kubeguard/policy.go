package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/Burhan-21/kubeguard/internal/rules/reliability"
	"github.com/Burhan-21/kubeguard/internal/rules/security"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage and view policies",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available rules",
	Run: func(cmd *cobra.Command, args []string) {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tCATEGORY\tDEFAULT SEVERITY")
		fmt.Fprintln(w, "--\t-----\t--------\t----------------")

		for _, r := range security.AllSecurityRules() {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.ID(), r.Title(), r.Category(), r.DefaultSeverity())
		}

		for _, r := range reliability.AllReliabilityRules() {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.ID(), r.Title(), r.Category(), r.DefaultSeverity())
		}

		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(policyCmd)
	policyCmd.AddCommand(listCmd)
}
