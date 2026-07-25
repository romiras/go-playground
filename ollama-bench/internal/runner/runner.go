package runner

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"play.ground/ollama-bench/assets"
	"play.ground/ollama-bench/internal/client"
	"play.ground/ollama-bench/internal/config"
	"play.ground/ollama-bench/internal/metrics"
	"play.ground/ollama-bench/internal/reporter"
)

const (
	numRuns           = 5
	prefillNumPredict = 16
	decodeNumPredict  = 512
	warmupPrompt      = "say hi"
	decodePrompt      = "Write a detailed technical explanation of how TCP congestion control works, covering slow start, congestion avoidance, fast retransmit, and fast recovery."
)

// Pool orchestrates concurrent per-host benchmark workers.
type Pool struct {
	cfg *config.Config
}

// New creates a worker pool for cfg.
func New(cfg *config.Config) *Pool {
	return &Pool{cfg: cfg}
}

// Run executes the full benchmark across all hosts concurrently.
func (p *Pool) Run(ctx context.Context) []reporter.HostResult {
	results := make([]reporter.HostResult, len(p.cfg.Hosts))
	var wg sync.WaitGroup

	for i, host := range p.cfg.Hosts {
		wg.Add(1)
		go func(i int, host string) {
			defer wg.Done()
			results[i] = p.runHost(ctx, host)
		}(i, host)
	}
	wg.Wait()
	return results
}

func (p *Pool) runHost(ctx context.Context, host string) (result reporter.HostResult) {
	result.Host = host
	cli := client.New(host, time.Duration(p.cfg.TimeoutSec)*time.Second)

	// Always attempt teardown if we got past health (model may be loaded).
	var pulled bool
	defer func() {
		if !pulled {
			return
		}
		if err := cli.Unload(context.Background(), p.cfg.Model); err != nil {
			log.Printf("[%s] unload: %v", host, err)
		}
		if p.cfg.Cleanup {
			if err := cli.Delete(context.Background(), p.cfg.Model); err != nil {
				log.Printf("[%s] delete: %v", host, err)
			}
		}
	}()

	if err := ctx.Err(); err != nil {
		result.Status = reporter.StatusFailed
		result.Error = err.Error()
		return result
	}

	if _, err := cli.Version(ctx); err != nil {
		result.Status = reporter.StatusUnreachable
		result.Error = err.Error()
		log.Printf("[%s] unreachable: %v", host, err)
		return result
	}

	log.Printf("[%s] pulling %s", host, p.cfg.Model)
	if err := cli.Pull(ctx, p.cfg.Model); err != nil {
		result.Status = reporter.StatusFailed
		result.Error = fmt.Sprintf("pull: %v", err)
		log.Printf("[%s] %s", host, result.Error)
		return result
	}
	pulled = true

	if err := ctx.Err(); err != nil {
		result.Status = reporter.StatusFailed
		result.Error = err.Error()
		return result
	}

	log.Printf("[%s] warmup", host)
	if _, err := cli.Generate(ctx, p.cfg.Model, warmupPrompt, prefillNumPredict); err != nil {
		result.Status = reporter.StatusFailed
		result.Error = fmt.Sprintf("warmup: %v", err)
		log.Printf("[%s] %s", host, result.Error)
		return result
	}

	prefillRuns, err := p.runSuite(ctx, cli, assets.PrefillContext, prefillNumPredict)
	if err != nil {
		result.Status = reporter.StatusFailed
		result.Error = fmt.Sprintf("prefill: %v", err)
		log.Printf("[%s] %s", host, result.Error)
		return result
	}
	result.Prefill = metrics.SummarizePrefill(prefillRuns)

	decodeRuns, err := p.runSuite(ctx, cli, decodePrompt, decodeNumPredict)
	if err != nil {
		result.Status = reporter.StatusFailed
		result.Error = fmt.Sprintf("decode: %v", err)
		log.Printf("[%s] %s", host, result.Error)
		return result
	}
	result.Decode = metrics.SummarizeDecode(decodeRuns)

	if result.Prefill.HighVariance || result.Decode.HighVariance {
		result.Status = reporter.StatusHighVar
	} else {
		result.Status = reporter.StatusOK
	}
	log.Printf("[%s] done status=%s prefill_p50=%.2f decode_p50=%.2f",
		host, result.Status, result.Prefill.P50TPS, result.Decode.P50TPS)
	return result
}

func (p *Pool) runSuite(ctx context.Context, cli *client.Client, prompt string, numPredict int) ([]metrics.RunTelemetry, error) {
	runs := make([]metrics.RunTelemetry, 0, numRuns)
	for i := 0; i < numRuns; i++ {
		if err := ctx.Err(); err != nil {
			return runs, err
		}
		gen, err := cli.Generate(ctx, p.cfg.Model, prompt, numPredict)
		if err != nil {
			return runs, fmt.Errorf("run %d: %w", i+1, err)
		}
		runs = append(runs, metrics.RunTelemetry{
			TTFTMs:             float64(gen.TTFT.Microseconds()) / 1000.0,
			TotalDurationNs:    gen.TotalDuration,
			LoadDurationNs:     gen.LoadDuration,
			PromptEvalCount:    gen.PromptEvalCount,
			PromptEvalDuration: gen.PromptEvalDur,
			EvalCount:          gen.EvalCount,
			EvalDuration:       gen.EvalDuration,
			PrefillTPS:         metrics.PrefillTPS(gen.PromptEvalCount, gen.PromptEvalDur),
			DecodeTPS:          metrics.DecodeTPS(gen.EvalCount, gen.EvalDuration),
		})
	}
	return runs, nil
}
