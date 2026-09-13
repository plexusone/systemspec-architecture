package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/plexusone/systemspec-architecture/assure"
)

// FormatAssureReport renders an assure.Report as either human-readable
// console text (format == "console" or "") or JSON (format == "json").
func FormatAssureReport(report assure.Report, format string) (string, error) {
	switch format {
	case "", "console":
		return formatAssureConsole(report), nil
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal assure report: %w", err)
		}
		return string(data) + "\n", nil
	default:
		return "", fmt.Errorf("unknown format %q (want console or json)", format)
	}
}

func formatAssureConsole(report assure.Report) string {
	var b strings.Builder

	b.WriteString("Assurance Coverage\n\n")
	writeCoverageLine(&b, "Node tests", report.Nodes, assure.EvidenceTests)
	writeCoverageLine(&b, "Node metrics", report.Nodes, assure.EvidenceMetrics)
	writeCoverageLine(&b, "Node detections", report.Nodes, assure.EvidenceDetections)
	writeCoverageLine(&b, "Node deployment", report.Nodes, assure.EvidenceDeployment)
	writeCoverageLine(&b, "Relationship tests", report.Relationships, assure.EvidenceTests)
	writeCoverageLine(&b, "Relationship metrics", report.Relationships, assure.EvidenceMetrics)
	writeCoverageLine(&b, "Relationship detections", report.Relationships, assure.EvidenceDetections)
	writeCoverageLine(&b, "Relationship deployment", report.Relationships, assure.EvidenceDeployment)

	if len(report.Gaps) == 0 {
		b.WriteString("\n✅ No gaps found\n")
		return b.String()
	}

	fmt.Fprintf(&b, "\nGaps (%d):\n", len(report.Gaps))
	for _, g := range report.Gaps {
		missing := make([]string, len(g.Missing))
		for i, m := range g.Missing {
			missing[i] = string(m)
		}
		fmt.Fprintf(&b, "  %s %q missing: %s\n", g.Kind, g.ID, strings.Join(missing, ", "))
	}

	return b.String()
}

func writeCoverageLine(b *strings.Builder, label string, counts assure.Counts, category assure.EvidenceCategory) {
	if counts.Total == 0 {
		fmt.Fprintf(b, "  %-24s n/a (no elements)\n", label)
		return
	}
	fmt.Fprintf(b, "  %-24s %5.1f%%\n", label, counts.Coverage(category)*100)
}
