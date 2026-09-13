package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-architecture/cli"
)

func newAssureCmd() *cobra.Command {
	var formatFlag string

	cmd := &cobra.Command{
		Use:   "assure <architecture.json>",
		Short: "Report assurance-reference coverage: which nodes and relationships declare tests, metrics, detections, and deployment evidence",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := cli.Assure(cli.AssureOptions{ArchitecturePath: args[0]})
			if err != nil {
				return err
			}

			out, err := cli.FormatAssureReport(report, formatFlag)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), out)
			return err
		},
	}

	cmd.Flags().StringVar(&formatFlag, "format", "console", "output format: console or json")

	return cmd
}
