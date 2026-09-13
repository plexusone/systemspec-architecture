package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-architecture/cli"
)

func newBindCmd() *cobra.Command {
	var (
		pidlFlag   string
		formatFlag string
	)

	cmd := &cobra.Command{
		Use:   "bind <architecture.json>",
		Short: "Verify protocol bindings against a PIDL protocol document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := cli.Bind(cli.BindOptions{
				ArchitecturePath: args[0],
				PIDLPath:         pidlFlag,
			})
			if err != nil {
				return err
			}

			report, err := cli.FormatBindReport(result, formatFlag)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprint(cmd.OutOrStdout(), report); err != nil {
				return err
			}

			if !result.OK() {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&pidlFlag, "pidl", "", "path to a PIDL protocol JSON document (required)")
	cmd.Flags().StringVar(&formatFlag, "format", "console", "output format: console or json")
	if err := cmd.MarkFlagRequired("pidl"); err != nil {
		panic(err) // programming error: "pidl" flag was just registered above
	}

	return cmd
}
