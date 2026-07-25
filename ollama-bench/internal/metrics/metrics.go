package metrics

import (
	"math"
	"sort"
)

const varianceThreshold = 0.10 // flag when stddev/mean > 10%

// RunTelemetry is one generate call's dual telemetry.
type RunTelemetry struct {
	TTFTMs             float64 `json:"ttft_ms"`
	TotalDurationNs    int64   `json:"total_duration_ns"`
	LoadDurationNs     int64   `json:"load_duration_ns"`
	PromptEvalCount    int     `json:"prompt_eval_count"`
	PromptEvalDuration int64   `json:"prompt_eval_duration_ns"`
	EvalCount          int     `json:"eval_count"`
	EvalDuration       int64   `json:"eval_duration_ns"`
	PrefillTPS         float64 `json:"prefill_tps"`
	DecodeTPS          float64 `json:"decode_tps"`
}

// SuiteSummary aggregates a 5-run suite after discarding run 1.
type SuiteSummary struct {
	Runs         []RunTelemetry `json:"runs"`
	P50TPS       float64        `json:"p50_tps"`
	StddevTPS    float64        `json:"stddev_tps"`
	P50TTFTMs    float64        `json:"p50_ttft_ms,omitempty"`
	HighVariance bool           `json:"high_variance"`
}

// PrefillTPS computes prompt tokens/sec from Ollama nanosecond counters.
func PrefillTPS(promptEvalCount int, promptEvalDurationNs int64) float64 {
	if promptEvalDurationNs <= 0 || promptEvalCount <= 0 {
		return 0
	}
	return float64(promptEvalCount) / float64(promptEvalDurationNs) * 1e9
}

// DecodeTPS computes generation tokens/sec from Ollama nanosecond counters.
func DecodeTPS(evalCount int, evalDurationNs int64) float64 {
	if evalDurationNs <= 0 || evalCount <= 0 {
		return 0
	}
	return float64(evalCount) / float64(evalDurationNs) * 1e9
}

// Median returns the p50 of values (average of middle two when even length).
func Median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	cp := append([]float64(nil), values...)
	sort.Float64s(cp)
	mid := len(cp) / 2
	if len(cp)%2 == 0 {
		return (cp[mid-1] + cp[mid]) / 2
	}
	return cp[mid]
}

// SampleStdDev returns the sample standard deviation (n-1 denominator).
func SampleStdDev(values []float64) float64 {
	n := len(values)
	if n < 2 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(n)
	var sq float64
	for _, v := range values {
		d := v - mean
		sq += d * d
	}
	return math.Sqrt(sq / float64(n-1))
}

// Mean returns the arithmetic mean.
func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// HighVariance reports whether coefficient of variation exceeds 10%.
func HighVariance(values []float64) bool {
	mean := Mean(values)
	if mean == 0 {
		return false
	}
	return SampleStdDev(values)/math.Abs(mean) > varianceThreshold
}

// SummarizePrefill discards run 1 and aggregates prefill TPS over runs 2–5.
func SummarizePrefill(runs []RunTelemetry) SuiteSummary {
	s := SuiteSummary{Runs: runs}
	kept := keepAfterWarm(runs)
	tps := make([]float64, 0, len(kept))
	for _, r := range kept {
		tps = append(tps, r.PrefillTPS)
	}
	s.P50TPS = Median(tps)
	s.StddevTPS = SampleStdDev(tps)
	s.HighVariance = HighVariance(tps)
	return s
}

// SummarizeDecode discards run 1 and aggregates decode TPS + TTFT over runs 2–5.
func SummarizeDecode(runs []RunTelemetry) SuiteSummary {
	s := SuiteSummary{Runs: runs}
	kept := keepAfterWarm(runs)
	tps := make([]float64, 0, len(kept))
	ttfts := make([]float64, 0, len(kept))
	for _, r := range kept {
		tps = append(tps, r.DecodeTPS)
		ttfts = append(ttfts, r.TTFTMs)
	}
	s.P50TPS = Median(tps)
	s.StddevTPS = SampleStdDev(tps)
	s.P50TTFTMs = Median(ttfts)
	s.HighVariance = HighVariance(tps)
	return s
}

func keepAfterWarm(runs []RunTelemetry) []RunTelemetry {
	if len(runs) <= 1 {
		return nil
	}
	return runs[1:]
}
