package metrics

import (
	"math"
	"testing"
)

func TestPrefillDecodeTPS(t *testing.T) {
	// 100 tokens in 1e9 ns => 100 TPS
	if got := PrefillTPS(100, 1e9); got != 100 {
		t.Errorf("PrefillTPS = %v", got)
	}
	if got := DecodeTPS(50, 5e8); math.Abs(got-100) > 1e-9 {
		t.Errorf("DecodeTPS = %v", got)
	}
	if PrefillTPS(10, 0) != 0 {
		t.Error("zero duration should yield 0")
	}
}

func TestMedian(t *testing.T) {
	if got := Median([]float64{1, 3, 2}); got != 2 {
		t.Errorf("odd median = %v", got)
	}
	if got := Median([]float64{1, 2, 3, 4}); got != 2.5 {
		t.Errorf("even median = %v", got)
	}
	if Median(nil) != 0 {
		t.Error("empty median")
	}
}

func TestSampleStdDev(t *testing.T) {
	// values 2,4 mean=3, variance=((1)^2+(1)^2)/1 = 2, stddev=sqrt(2)
	got := SampleStdDev([]float64{2, 4})
	want := math.Sqrt(2)
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("stddev = %v want %v", got, want)
	}
}

func TestHighVariance(t *testing.T) {
	stable := []float64{100, 101, 99, 100.5}
	if HighVariance(stable) {
		t.Error("stable should not flag")
	}
	noisy := []float64{100, 130, 70, 140}
	if !HighVariance(noisy) {
		t.Error("noisy should flag")
	}
}

func TestSummarizeDiscardRun1(t *testing.T) {
	runs := []RunTelemetry{
		{PrefillTPS: 999, DecodeTPS: 999, TTFTMs: 999}, // discarded
		{PrefillTPS: 10, DecodeTPS: 20, TTFTMs: 40},
		{PrefillTPS: 12, DecodeTPS: 22, TTFTMs: 42},
		{PrefillTPS: 11, DecodeTPS: 21, TTFTMs: 41},
		{PrefillTPS: 13, DecodeTPS: 23, TTFTMs: 43},
	}
	pre := SummarizePrefill(runs)
	if pre.P50TPS != 11.5 {
		t.Errorf("prefill p50 = %v", pre.P50TPS)
	}
	dec := SummarizeDecode(runs)
	if dec.P50TPS != 21.5 {
		t.Errorf("decode p50 = %v", dec.P50TPS)
	}
	if dec.P50TTFTMs != 41.5 {
		t.Errorf("ttft p50 = %v", dec.P50TTFTMs)
	}
}
