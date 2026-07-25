# ollama-bench

Benchmark [Ollama](https://ollama.com/) inference hosts for **prefill throughput** and **decode throughput**, with dual telemetry from both the client (TTFT) and Ollama's streaming API counters.

Each host runs the same deterministic pipeline in isolation. Multiple hosts are benchmarked concurrently (one worker goroutine per host).

## What it measures

| Test | Prompt | `num_predict` | Metric |
| --- | --- | --- | --- |
| Warmup | `"say hi"` | 16 | Discarded |
| A — Prefill | ~25 KB embedded prose (`assets/prefill_context.txt`) | 16 | Prompt eval TPS |
| B — Decode | Fixed ~20-token technical prompt | 512 | Generation TPS + TTFT |

Each test case runs **5 times**. Run 1 is discarded; the reported value is the **p50 median of runs 2–5**. Hosts with `stddev / mean > 10%` are flagged as `high-variance`.

Generation options are locked for reproducibility:

```json
{
  "temperature": 0.0,
  "seed": 42,
  "top_k": 1,
  "top_p": 1.0
}
```

## Requirements

- Go 1.26+ (see `go.work` at repo root)
- One or more reachable Ollama servers (`GET /api/version`)

## Build

```bash
cd ollama-bench
go build -o ollama-bench ./cmd/ollama-bench
```

## Usage

```bash
./ollama-bench -config config.yaml
```

| Flag | Default | Description |
| --- | --- | --- |
| `-config` | `config.yaml` | Path to YAML configuration |

The tool prints a summary table to stdout and writes structured JSON to the path configured in `output`. Progress and errors are logged to stderr.

Exit code `1` if any host is `failed` or `unreachable`. `high-variance` hosts still exit `0`.

`SIGINT` / `SIGTERM` cancel in-flight work; hosts that already pulled the model attempt best-effort unload/delete before exit.

## Configuration

Example `config.yaml`:

```yaml
hosts:
  - http://127.0.0.1:11434
  - http://192.168.1.10:11434
model: llama3.2:latest
output: results.json
timeout_sec: 900
cleanup: true
```

| Field | Required | Description |
| --- | --- | --- |
| `hosts` | yes | Ollama base URLs (trailing `/` is stripped) |
| `model` | yes | Model tag passed to `/api/pull` and `/api/generate` |
| `output` | no | JSON report path (default: `results.json`) |
| `timeout_sec` | yes | Per-inference request timeout in seconds |
| `cleanup` | no | If `true`, unload VRAM (`keep_alive: 0`) and `DELETE /api/delete` after the benchmark |

### Timeout tuning

Prefill evaluates a large embedded prompt (~25 KB, ~4k tokens depending on tokenizer). Slower models or CPU-only hosts may need a high `timeout_sec` (e.g. 600–900). A `context deadline exceeded` on prefill run 1 usually means the timeout is too low, not a bug in the benchmark.

When `cleanup: true`, a failed run still removes the pulled model; set `cleanup: false` while iterating on timeouts to avoid re-pulling each attempt.

## Per-host lifecycle

```
GET /api/version  (5s probe)
  → POST /api/pull
  → warmup ("say hi")
  → prefill × 5
  → decode × 5
  → unload (keep_alive: 0)
  → DELETE /api/delete  (if cleanup)
```

Failed hosts do not stop other workers in the pool.

## Output

### Console table

```
HOST                      STATUS          PREFILL TPS  DECODE TPS  TTFT ms  NOTES
http://127.0.0.1:11434    ok              21.85        45.12       120.50
```

Status values: `ok`, `high-variance`, `failed`, `unreachable`.

### JSON report

Written to `output`. Top-level fields:

- `model` — model tag from config
- `generated_at` — UTC timestamp
- `hosts[]` — per-host results with raw run telemetry and aggregates

Each run records client `ttft_ms` plus Ollama nanosecond counters (`prompt_eval_*`, `eval_*`, `total_duration_ns`, `load_duration_ns`). Aggregates include `p50_tps`, `stddev_tps`, `high_variance`, and for decode also `p50_ttft_ms`.

### Interpret results

Pretty-print a human-readable summary (stdlib Python only):

```bash
python3 interpret_results.py results.json
```

## Tests

```bash
go test ./...
```

## Further reading

See [docs/SPEC.md](docs/SPEC.md) for the full architectural specification.
