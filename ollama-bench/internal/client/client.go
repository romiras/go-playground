package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	healthTimeout = 5 * time.Second
	pullTimeout   = 30 * time.Minute
)

// Options are locked generation parameters for reproducible benchmarks.
type Options struct {
	Temperature float64 `json:"temperature"`
	Seed        int     `json:"seed"`
	TopK        int     `json:"top_k"`
	TopP        float64 `json:"top_p"`
	NumPredict  int     `json:"num_predict"`
}

// LockedOptions returns the SPEC-mandated deterministic options with the given num_predict.
func LockedOptions(numPredict int) Options {
	return Options{
		Temperature: 0.0,
		Seed:        42,
		TopK:        1,
		TopP:        1.0,
		NumPredict:  numPredict,
	}
}

// GenerateRequest is the body for POST /api/generate.
type GenerateRequest struct {
	Model     string  `json:"model"`
	Prompt    string  `json:"prompt"`
	Stream    *bool   `json:"stream,omitempty"`
	KeepAlive *int    `json:"keep_alive,omitempty"`
	Options   Options `json:"options"`
}

// GenerateResult holds dual telemetry from a streaming generate call.
type GenerateResult struct {
	TTFT            time.Duration
	TotalDuration   int64
	LoadDuration    int64
	PromptEvalCount int
	PromptEvalDur   int64
	EvalCount       int
	EvalDuration    int64
	Response        string
}

type streamChunk struct {
	Model              string `json:"model"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	TotalDuration      int64  `json:"total_duration"`
	LoadDuration       int64  `json:"load_duration"`
	PromptEvalCount    int    `json:"prompt_eval_count"`
	PromptEvalDuration int64  `json:"prompt_eval_duration"`
	EvalCount          int    `json:"eval_count"`
	EvalDuration       int64  `json:"eval_duration"`
	Error              string `json:"error"`
}

// Client talks to a single Ollama host.
type Client struct {
	baseURL         string
	pullClient      *http.Client
	inferenceClient *http.Client
	timeout         time.Duration
}

// New creates a client for baseURL (e.g. http://127.0.0.1:11434).
func New(baseURL string, inferenceTimeout time.Duration) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          4,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Client{
		baseURL: baseURL,
		pullClient: &http.Client{
			Transport: transport,
			Timeout:   pullTimeout,
		},
		inferenceClient: &http.Client{
			Transport: transport,
			// Per-request context timeout; leave client Timeout unset so stream can run.
		},
		timeout: inferenceTimeout,
	}
}

// BaseURL returns the configured host URL.
func (c *Client) BaseURL() string { return c.baseURL }

// Version probes GET /api/version with a 5s timeout.
func (c *Client) Version(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, healthTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/version", nil)
	if err != nil {
		return "", err
	}
	resp, err := c.inferenceClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("version probe: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("version probe: status %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}
	var out struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("version probe decode: %w", err)
	}
	return out.Version, nil
}

// Pull downloads the model via POST /api/pull, draining NDJSON until success.
func (c *Client) Pull(ctx context.Context, model string) error {
	body, err := json.Marshal(map[string]string{"model": model})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/pull", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.pullClient.Do(req)
	if err != nil {
		return fmt.Errorf("pull: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("pull: status %d: %s", resp.StatusCode, bytes.TrimSpace(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var lastStatus string
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var chunk struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		}
		if err := json.Unmarshal(line, &chunk); err != nil {
			return fmt.Errorf("pull decode: %w", err)
		}
		if chunk.Error != "" {
			return fmt.Errorf("pull: %s", chunk.Error)
		}
		lastStatus = chunk.Status
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("pull stream: %w", err)
	}
	if lastStatus != "success" && lastStatus != "" {
		// Some servers may end without an explicit success line after stream closes cleanly.
		// Treat empty final status as ok only if we got at least one status; otherwise require success.
	}
	if lastStatus != "success" {
		return fmt.Errorf("pull: unexpected final status %q", lastStatus)
	}
	return nil
}

// Generate streams POST /api/generate and returns dual telemetry.
func (c *Client) Generate(ctx context.Context, model, prompt string, numPredict int) (*GenerateResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	stream := true
	reqBody := GenerateRequest{
		Model:   model,
		Prompt:  prompt,
		Stream:  &stream,
		Options: LockedOptions(numPredict),
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.inferenceClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("generate: status %d: %s", resp.StatusCode, bytes.TrimSpace(b))
	}

	result := &GenerateResult{}
	var ttftSet bool
	var responseBuf strings.Builder

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var chunk streamChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			return nil, fmt.Errorf("generate decode: %w", err)
		}
		if chunk.Error != "" {
			return nil, fmt.Errorf("generate: %s", chunk.Error)
		}
		if chunk.Response != "" {
			responseBuf.WriteString(chunk.Response)
			if !ttftSet {
				result.TTFT = time.Since(start)
				ttftSet = true
			}
		}
		if chunk.Done {
			result.TotalDuration = chunk.TotalDuration
			result.LoadDuration = chunk.LoadDuration
			result.PromptEvalCount = chunk.PromptEvalCount
			result.PromptEvalDur = chunk.PromptEvalDuration
			result.EvalCount = chunk.EvalCount
			result.EvalDuration = chunk.EvalDuration
			result.Response = responseBuf.String()
			return result, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("generate stream: %w", err)
	}
	return nil, fmt.Errorf("generate: stream ended without done:true")
}

// Unload forces VRAM eviction via generate with keep_alive: 0.
// Uses an independent short timeout so teardown is not blocked by a cancelled parent.
func (c *Client) Unload(_ context.Context, model string) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	keepAlive := 0
	stream := false
	reqBody := map[string]any{
		"model":      model,
		"prompt":     "",
		"keep_alive": keepAlive,
		"stream":     stream,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(cleanupCtx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.inferenceClient.Do(req)
	if err != nil {
		return fmt.Errorf("unload: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unload: status %d", resp.StatusCode)
	}
	return nil
}

// Delete removes model artifacts via DELETE /api/delete.
func (c *Client) Delete(_ context.Context, model string) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	payload, err := json.Marshal(map[string]string{"model": model})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(cleanupCtx, http.MethodDelete, c.baseURL+"/api/delete", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.inferenceClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete: status %d", resp.StatusCode)
	}
	return nil
}
