package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-architecture/cli"
)

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a SAS architecture document for another specification",
	}
	cmd.AddCommand(newExportThreatModelCmd())
	return cmd
}

func newExportThreatModelCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "threat-model <architecture.json>",
		Short: "Export the system-under-analysis (DiagramIR) for github.com/grokify/threat-model-spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := cli.ExportThreatModel(cli.ExportThreatModelOptions{ArchitecturePath: args[0]})
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), out)
			return err
		},
	}
}
