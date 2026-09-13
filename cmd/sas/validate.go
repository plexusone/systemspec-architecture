package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-architecture/cli"
	"github.com/plexusone/systemspec-architecture/validate"
)

func newValidateCmd() *cobra.Command {
	var (
		profileFlag string
		formatFlag  string
	)

	cmd := &cobra.Command{
		Use:   "validate <architecture.json>",
		Short: "Validate a SAS architecture document against referential integrity and optional profiles",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := cli.Validate(cli.ValidateOptions{
				ArchitecturePath: args[0],
				Profiles:         parseProfiles(profileFlag),
			})
			if err != nil {
				return err
			}

			report, err := cli.FormatValidateReport(result, formatFlag)
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

	cmd.Flags().StringVar(&profileFlag, "profile", "", fmt.Sprintf("comma-separated profiles to validate against (%s)", joinProfiles(validate.KnownProfiles())))
	cmd.Flags().StringVar(&formatFlag, "format", "console", "output format: console or json")

	return cmd
}

func parseProfiles(flag string) []validate.Profile {
	values := splitCSV(flag)
	profiles := make([]validate.Profile, len(values))
	for i, v := range values {
		profiles[i] = validate.Profile(v)
	}
	return profiles
}

func joinProfiles(profiles []validate.Profile) string {
	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = string(p)
	}
	return strings.Join(names, ", ")
}
