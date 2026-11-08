# conductor-lens

AI-powered failure analyst for [Conductor OSS](https://github.com/conductor-oss/conductor) workflows.

Point it at a failed workflow ID and get back a root cause, explanation, and concrete fix — powered by Claude.

```
$ conductor-lens analyze --workflow-id abc-123

  ⟳ Fetching workflow abc-123 from http://localhost:8080...

WORKFLOW EXECUTION
────────────────────────────────────────────────────────────
  Name     : payment_processing_flow (v2)
  ID       : abc-123
  Status   : FAILED
  Started  : 2026-11-08 14:32:01
  Duration : 32.4s
  Tasks    : 12 total, 1 failed

AI ANALYSIS
────────────────────────────────────────────────────────────
  Severity  : ● HIGH

  ROOT CAUSE
  Task "charge_payment" timed out waiting for the payment
  gateway response after 30s.

  WHAT HAPPENED
  The fraud_check task upstream returned slowly (28s) due
  to a third-party API degradation, leaving only 2s for
  the downstream charge_payment task — not enough time
  to complete before its 30s timeout fired.

  SUGGESTED FIX
  Increase charge_payment timeout to 60s, or add a retry
  policy on fraud_check with exponential backoff (3
  attempts, 2s base delay) to handle upstream slowness.

  FAILED TASKS
  ✗ charge_payment [HTTP]
    Read timeout after 30000ms
    retried 1 time(s)
────────────────────────────────────────────────────────────
```

## Installation

```bash
go install github.com/shaileshpadave/conductor-lens@latest
```

Or build from source:

```bash
git clone https://github.com/shaileshpadave/conductor-lens
cd conductor-lens
go build -o conductor-lens .
```

## Setup

```bash
cp .env.example .env
# add your ANTHROPIC_API_KEY
```

Or pass flags directly:

```bash
conductor-lens analyze \
  --workflow-id abc-123 \
  --conductor-url http://my-conductor:8080 \
  --anthropic-key sk-ant-...
```

## Commands

### `analyze` — diagnose a specific failure

```bash
conductor-lens analyze --workflow-id <id>
```

Fetches the workflow execution, identifies failed tasks, and asks Claude for a root cause and fix.

### `search` — find recent failures

```bash
conductor-lens search
conductor-lens search --workflow-name payment_flow --limit 20
```

Lists recent failed workflows so you can pick one to analyze.

## Requirements

- Go 1.21+
- A running [Conductor OSS](https://github.com/conductor-oss/conductor) instance
- An [Anthropic API key](https://console.anthropic.com)

## How it works

1. Fetches the full workflow execution from Conductor's REST API (`/api/workflow/{id}?includeTasks=true`)
2. Extracts failed tasks, retry counts, error messages, and task timeline
3. Builds a structured prompt and sends it to Claude (Haiku — fast and cheap)
4. Parses the structured JSON response into a clean CLI output

## Roadmap

- [ ] `watch` command — tail live workflow executions and alert on failure
- [ ] `compare` command — diff two executions of the same workflow
- [ ] SQLite history — store past analyses and find similar failures
- [ ] `--json` flag — machine-readable output for piping into other tools
