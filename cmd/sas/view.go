package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/plexusone/systemspec-architecture/cli"
)

func newViewCmd() *cobra.Command {
	var (
		viewFlag              string
		formatFlag            string
		groupByFlag           string
		includeKindsFlag      string
		includeRelationsFlag  string
		includeBoundariesFlag string
	)

	cmd := &cobra.Command{
		Use:   "view <architecture.json>",
		Short: "Render a view of a SAS architecture document as Mermaid, D2, or Graphviz DOT",
		Long: "Render a view of a SAS architecture document as Mermaid, D2, or Graphviz DOT.\n\n" +
			"With --view, looks up a view declared in the architecture's own \"views\" list.\n" +
			"Without it, builds an ad-hoc view from --group-by/--include-kinds/--include-relations/--include-boundaries.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := cli.View(cli.ViewOptions{
				ArchitecturePath:  args[0],
				ViewID:            viewFlag,
				GroupBy:           groupByFlag,
				IncludeKinds:      splitCSV(includeKindsFlag),
				IncludeRelations:  splitCSV(includeRelationsFlag),
				IncludeBoundaries: splitCSV(includeBoundariesFlag),
				Format:            formatFlag,
			})
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), out)
			return err
		},
	}

	cmd.Flags().StringVar(&viewFlag, "view", "", "ID of a view declared in the architecture's own views list")
	cmd.Flags().StringVar(&formatFlag, "format", "mermaid", "output format: mermaid, d2, or dot")
	cmd.Flags().StringVar(&groupByFlag, "group-by", "", "boundary kind to group nodes by (ad-hoc view only)")
	cmd.Flags().StringVar(&includeKindsFlag, "include-kinds", "", "comma-separated node kinds to include (ad-hoc view only)")
	cmd.Flags().StringVar(&includeRelationsFlag, "include-relations", "", "comma-separated relationship kinds to include (ad-hoc view only)")
	cmd.Flags().StringVar(&includeBoundariesFlag, "include-boundaries", "", "comma-separated boundary IDs to include (ad-hoc view only)")

	return cmd
}

func splitCSV(flag string) []string {
	if flag == "" {
		return nil
	}
	parts := strings.Split(flag, ",")
	values := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		values = append(values, p)
	}
	return values
}
