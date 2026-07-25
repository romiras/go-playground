package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"play.ground/ollama-bench/internal/metrics"
)

// Status constants for host results.
const (
	StatusOK          = "ok"
	StatusHighVar     = "high-variance"
	StatusFailed      = "failed"
	StatusUnreachable = "unreachable"
)

// HostResult is one host's benchmark outcome.
type HostResult struct {
	Host    string               `json:"host"`
	Status  string               `json:"status"`
	Error   string               `json:"error,omitempty"`
	Prefill metrics.SuiteSummary `json:"prefill"`
	Decode  metrics.SuiteSummary `json:"decode"`
}

// Report is the full JSON export schema.
type Report struct {
	Model       string       `json:"model"`
	GeneratedAt time.Time    `json:"generated_at"`
	Hosts       []HostResult `json:"hosts"`
}

// WriteTable prints a console summary table.
func WriteTable(w io.Writer, results []HostResult) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "HOST\tSTATUS\tPREFILL TPS\tDECODE TPS\tTTFT ms\tNOTES")
	for _, r := range results {
		notes := r.Error
		prefill := formatFloat(r.Prefill.P50TPS)
		decode := formatFloat(r.Decode.P50TPS)
		ttft := formatFloat(r.Decode.P50TTFTMs)
		if r.Status == StatusOK || r.Status == StatusHighVar {
			if r.Prefill.HighVariance || r.Decode.HighVariance {
				if notes != "" {
					notes += "; "
				}
				notes += "variance>10%"
			}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			r.Host, r.Status, prefill, decode, ttft, notes)
	}
	return tw.Flush()
}

// WriteJSON writes the structured report to path.
func WriteJSON(path string, model string, results []HostResult) error {
	rep := Report{
		Model:       model,
		GeneratedAt: time.Now().UTC(),
		Hosts:       results,
	}
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func formatFloat(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f", v)
}
