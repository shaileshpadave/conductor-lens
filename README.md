# conductor-lens

AI-powered failure analyst for [Conductor OSS](https://github.com/conductor-oss/conductor) workflows.

Point it at a failed workflow ID and get back a root cause, plain-English explanation, and concrete fix suggestion — powered by Claude (Haiku).

---

## Demo

```
$ conductor-lens search --conductor-url http://localhost:8080

  RECENT FAILURES (3 found)
  ────────────────────────────────────────────────────────────
  1. payment_processing_flow  50ffffc9-aa3f-444c-8ca0-e7bae533b200
     FAILED  2026-11-08T18:45:36.187Z
     Task validate_card_ref failed: Connection refused on POST...

  2. order_fulfillment  a3c12e88-bb01-4f2a-9d1e-cc3344556677
     FAILED  2026-11-08T17:12:04.003Z
     Task notify_warehouse timed out after 30s


$ conductor-lens analyze --workflow-id 50ffffc9-aa3f-444c-8ca0-e7bae533b200

WORKFLOW EXECUTION
────────────────────────────────────────────────────────────
  Name     : payment_processing_flow (v1)
  ID       : 50ffffc9-aa3f-444c-8ca0-e7bae533b200
  Status   : FAILED
  Started  : 2026-11-08 18:45:36
  Duration : 2m7.873s
  Tasks    : 3 total, 3 failed

AI ANALYSIS
────────────────────────────────────────────────────────────
  Severity  : ● LOW

  ROOT CAUSE
  The HTTP task is attempting to connect to a validation
  service at localhost:9999 that is not running or accessible.

  WHAT HAPPENED
  The workflow failed because validate_card_ref could not
  establish a connection to http://localhost:9999/validate.
  The task was retried 3 times but all attempts failed with
  'Connection refused'. Since all executions failed for the
  same reason, the entire workflow failed.

  SUGGESTED FIX
  Verify the validation service is running and listening on
  port 9999. If running in containers, avoid localhost —
  use the service's container name or internal DNS instead.
  Add exponential backoff (initialInterval: 1000ms,
  maxInterval: 10000ms) to handle transient service restarts.

  FAILED TASKS
  ✗ validate_card_ref [HTTP]
    Connection refused on POST http://localhost:9999/validate
    retried 2 time(s)
────────────────────────────────────────────────────────────
```

---

## Installation

**Build from source** (requires Go 1.21+):

```bash
git clone https://github.com/shaileshpadave/conductor-lens
cd conductor-lens
go build -o conductor-lens .
```

**Or install directly:**

```bash
go install github.com/shaileshpadave/conductor-lens@latest
```

---

## Setup

You need:
- A running [Conductor OSS](https://github.com/conductor-oss/conductor) instance
- An [Anthropic API key](https://console.anthropic.com)

```bash
cp .env.example .env
# edit .env and add your ANTHROPIC_API_KEY
```

Or pass everything as flags — no config file needed.

**Quick start with Docker (Conductor OSS):**

```bash
docker run -d --name conductor \
  -p 8080:8080 \
  -e conductor.db.type=memory \
  -e conductor.queue.type=internal \
  -e conductor.indexing.enabled=false \
  conductoross/conductor:latest
```

---

## Commands

### `search` — list recent failures

```bash
conductor-lens search
conductor-lens search --workflow-name payment_flow
conductor-lens search --workflow-name payment_flow --limit 20
```

Shows the most recent failed workflows so you can pick one to drill into.

### `analyze` — diagnose a specific failure

```bash
conductor-lens analyze --workflow-id <id>
```

Fetches the full workflow execution, extracts failed tasks and retry history, then asks Claude to identify the root cause and suggest a fix.

---

## Flags

| Flag | Env var | Default | Description |
|---|---|---|---|
| `--conductor-url` | `CONDUCTOR_URL` | `http://localhost:8080` | Conductor OSS base URL |
| `--conductor-key` | `CONDUCTOR_API_KEY` | _(none)_ | API key if your instance requires auth |
| `--anthropic-key` | `ANTHROPIC_API_KEY` | _(none)_ | Anthropic API key |

---

## How it works

```
Conductor OSS REST API          Claude (Haiku)
        │                              │
  GET /api/workflow/{id}               │
  ?includeTasks=true                   │
        │                              │
        ▼                              │
  Full execution JSON                  │
  - task statuses                      │
  - retry counts                       │   structured prompt
  - error messages         ───────────▶  (workflow + failed tasks
  - task timeline                      │   + error messages)
        │                              │
        └──────────────────────────────▶
                                       │
                              JSON response:
                              root_cause
                              explanation        colored
                              suggested_fix  ──▶ terminal
                                               output
```

1. Fetch the full workflow execution via `GET /api/workflow/{id}?includeTasks=true`
2. Extract failed tasks, retry counts, and error messages
3. Build a structured prompt and send to Claude Haiku (fast, cheap — typically < 1s)
4. Parse the structured JSON response and render it as color-coded CLI output

---

## Project layout

```
conductor-lens/
├── cmd/
│   ├── root.go       # global flags, env var bindings
│   ├── analyze.go    # analyze command
│   └── search.go     # search command
├── internal/
│   ├── conductor/
│   │   ├── client.go # Conductor REST API client
│   │   └── types.go  # Workflow, Task, SearchResult structs
│   ├── analyzer/
│   │   └── analyzer.go  # Claude integration, prompt builder
│   └── display/
│       └── display.go   # colored terminal output
└── main.go
```

---

## Roadmap

- [ ] `watch` — tail live workflow executions, alert on new failures
- [ ] `compare` — diff two executions of the same workflow to spot regressions
- [ ] SQLite history — store past analyses, surface similar past failures
- [ ] `--json` — machine-readable output for piping into other tools
- [ ] Batch mode — analyze all failures from a search in one pass
