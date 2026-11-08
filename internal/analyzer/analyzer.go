package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/shaileshpadave/conductor-lens/internal/conductor"
)

type Analysis struct {
	WorkflowID    string
	WorkflowName  string
	Status        string
	FailedTasks   []FailedTask
	RootCause     string
	Explanation   string
	SuggestedFix  string
	Severity      string // LOW, MEDIUM, HIGH
}

type FailedTask struct {
	Name   string
	Type   string
	Reason string
	Retries int
}

type Analyzer struct {
	client *anthropic.Client
}

func New(apiKey string) *Analyzer {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Analyzer{client: &client}
}

func (a *Analyzer) Analyze(ctx context.Context, wf *conductor.Workflow) (*Analysis, error) {
	failed := extractFailedTasks(wf)

	analysis := &Analysis{
		WorkflowID:   wf.WorkflowID,
		WorkflowName: wf.WorkflowName,
		Status:       string(wf.Status),
		FailedTasks:  failed,
	}

	prompt := buildPrompt(wf, failed)

	msg, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("calling Claude: %w", err)
	}

	if len(msg.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	raw := msg.Content[0].Text
	if err := parseResponse(raw, analysis); err != nil {
		// fallback: use raw text
		analysis.RootCause = "See explanation below"
		analysis.Explanation = raw
		analysis.SuggestedFix = "Review the explanation above"
	}

	analysis.Severity = computeSeverity(wf, failed)
	return analysis, nil
}

func extractFailedTasks(wf *conductor.Workflow) []FailedTask {
	var failed []FailedTask
	for _, t := range wf.Tasks {
		if t.Status == "FAILED" || t.Status == "TIMED_OUT" {
			failed = append(failed, FailedTask{
				Name:    t.ReferenceTaskName,
				Type:    t.TaskType,
				Reason:  t.ReasonForIncompletion,
				Retries: t.RetryCount,
			})
		}
	}
	return failed
}

func buildPrompt(wf *conductor.Workflow, failed []FailedTask) string {
	duration := ""
	if wf.EndTime > 0 && wf.StartTime > 0 {
		d := time.Duration(wf.EndTime-wf.StartTime) * time.Millisecond
		duration = d.String()
	}

	var sb strings.Builder
	sb.WriteString("You are an expert in distributed workflow systems, specifically Conductor OSS.\n")
	sb.WriteString("Analyze this failed workflow execution and provide a diagnosis.\n\n")

	sb.WriteString(fmt.Sprintf("WORKFLOW: %s (version %d)\n", wf.WorkflowName, wf.WorkflowVersion))
	sb.WriteString(fmt.Sprintf("STATUS: %s\n", wf.Status))
	if duration != "" {
		sb.WriteString(fmt.Sprintf("DURATION: %s\n", duration))
	}
	if wf.ReasonForIncompletion != "" {
		sb.WriteString(fmt.Sprintf("TOP-LEVEL REASON: %s\n", wf.ReasonForIncompletion))
	}

	sb.WriteString(fmt.Sprintf("\nTOTAL TASKS: %d\n", len(wf.Tasks)))
	sb.WriteString(fmt.Sprintf("FAILED TASKS: %d\n\n", len(failed)))

	for _, ft := range failed {
		sb.WriteString(fmt.Sprintf("- Task: %s (type: %s)\n", ft.Name, ft.Type))
		if ft.Reason != "" {
			sb.WriteString(fmt.Sprintf("  Reason: %s\n", ft.Reason))
		}
		sb.WriteString(fmt.Sprintf("  Retries: %d\n", ft.Retries))
	}

	// include a compact task timeline
	sb.WriteString("\nTASK TIMELINE (reference_name → status):\n")
	for _, t := range wf.Tasks {
		sb.WriteString(fmt.Sprintf("  %s [%s] → %s\n", t.ReferenceTaskName, t.TaskType, t.Status))
	}

	sb.WriteString(`
Respond in this exact JSON format:
{
  "root_cause": "one sentence identifying the root cause",
  "explanation": "2-3 sentences explaining what happened and why",
  "suggested_fix": "concrete actionable fix — specific config, retry policy, timeout value, or code change"
}
`)
	return sb.String()
}

func parseResponse(raw string, out *Analysis) error {
	// extract JSON from response — Claude sometimes adds markdown fences
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 {
		return fmt.Errorf("no JSON found in response")
	}

	var result struct {
		RootCause    string `json:"root_cause"`
		Explanation  string `json:"explanation"`
		SuggestedFix string `json:"suggested_fix"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
		return err
	}

	out.RootCause = result.RootCause
	out.Explanation = result.Explanation
	out.SuggestedFix = result.SuggestedFix
	return nil
}

func computeSeverity(wf *conductor.Workflow, failed []FailedTask) string {
	if wf.Status == "TIMED_OUT" {
		return "HIGH"
	}
	if len(failed) > 3 {
		return "HIGH"
	}
	for _, ft := range failed {
		if ft.Retries > 2 {
			return "MEDIUM"
		}
	}
	return "LOW"
}
