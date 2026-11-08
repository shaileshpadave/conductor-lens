package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/shaileshpadave/conductor-lens/internal/analyzer"
	"github.com/shaileshpadave/conductor-lens/internal/conductor"
)

var (
	bold    = color.New(color.Bold)
	red     = color.New(color.FgRed, color.Bold)
	yellow  = color.New(color.FgYellow)
	green   = color.New(color.FgGreen)
	cyan    = color.New(color.FgCyan)
	faint   = color.New(color.Faint)
	heading = color.New(color.FgWhite, color.Bold, color.Underline)
)

func PrintWorkflowSummary(wf *conductor.Workflow) {
	fmt.Println()
	heading.Println("WORKFLOW EXECUTION")
	fmt.Println(strings.Repeat("─", 60))

	bold.Printf("  Name     ")
	fmt.Printf(": %s (v%d)\n", wf.WorkflowName, wf.WorkflowVersion)

	bold.Printf("  ID       ")
	fmt.Printf(": %s\n", wf.WorkflowID)

	bold.Printf("  Status   ")
	fmt.Printf(": ")
	printStatus(string(wf.Status))
	fmt.Println()

	if wf.StartTime > 0 {
		bold.Printf("  Started  ")
		fmt.Printf(": %s\n", formatTime(wf.StartTime))
	}
	if wf.EndTime > 0 {
		bold.Printf("  Duration ")
		d := time.Duration(wf.EndTime-wf.StartTime) * time.Millisecond
		fmt.Printf(": %s\n", d)
	}

	total := len(wf.Tasks)
	failed := 0
	for _, t := range wf.Tasks {
		if t.Status == "FAILED" || t.Status == "TIMED_OUT" {
			failed++
		}
	}

	bold.Printf("  Tasks    ")
	fmt.Printf(": %d total, ", total)
	if failed > 0 {
		red.Printf("%d failed", failed)
	} else {
		green.Printf("0 failed")
	}
	fmt.Println()
}

func PrintAnalysis(a *analyzer.Analysis) {
	fmt.Println()
	heading.Println("AI ANALYSIS")
	fmt.Println(strings.Repeat("─", 60))

	// Severity badge
	fmt.Printf("  ")
	bold.Printf("Severity  : ")
	switch a.Severity {
	case "HIGH":
		red.Printf("● HIGH\n")
	case "MEDIUM":
		yellow.Printf("● MEDIUM\n")
	default:
		green.Printf("● LOW\n")
	}

	fmt.Println()

	// Root cause
	bold.Println("  ROOT CAUSE")
	fmt.Printf("  %s\n", wrap(a.RootCause, 56, "  "))

	fmt.Println()

	// Explanation
	bold.Println("  WHAT HAPPENED")
	fmt.Printf("  %s\n", wrap(a.Explanation, 56, "  "))

	fmt.Println()

	// Fix
	bold.Println("  SUGGESTED FIX")
	cyan.Printf("  %s\n", wrap(a.SuggestedFix, 56, "  "))

	// Failed tasks
	if len(a.FailedTasks) > 0 {
		fmt.Println()
		bold.Println("  FAILED TASKS")
		for _, ft := range a.FailedTasks {
			red.Printf("  ✗ ")
			bold.Printf("%s", ft.Name)
			faint.Printf(" [%s]", ft.Type)
			fmt.Println()
			if ft.Reason != "" {
				faint.Printf("    %s\n", ft.Reason)
			}
			if ft.Retries > 0 {
				faint.Printf("    retried %d time(s)\n", ft.Retries)
			}
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
}

func PrintSearchResults(results []conductor.WorkflowSummary) {
	fmt.Println()
	heading.Printf("RECENT FAILURES (%d found)\n", len(results))
	fmt.Println(strings.Repeat("─", 60))

	if len(results) == 0 {
		green.Println("  No failures found.")
		return
	}

	for i, r := range results {
		fmt.Printf("  %d. ", i+1)
		bold.Printf("%s", r.WorkflowName)
		fmt.Printf("  %s\n", faintStr(r.WorkflowID))
		fmt.Printf("     ")
		printStatus(r.Status)
		if r.StartTime != "" {
			fmt.Printf("  %s", faintStr(r.StartTime))
		}
		if r.ReasonForIncompletion != "" {
			fmt.Printf("\n     ")
			faint.Printf("%s", truncate(r.ReasonForIncompletion, 55))
		}
		fmt.Println()
	}
	fmt.Println()
}

func PrintError(msg string) {
	red.Fprintf(color.Error, "\n  ✗ Error: %s\n\n", msg)
}

func PrintSpinner(msg string) {
	faint.Printf("  ⟳ %s...\n", msg)
}

// helpers

func printStatus(status string) {
	switch status {
	case "FAILED", "TIMED_OUT", "TERMINATED":
		red.Printf("%s", status)
	case "COMPLETED":
		green.Printf("%s", status)
	case "RUNNING":
		cyan.Printf("%s", status)
	default:
		yellow.Printf("%s", status)
	}
}

func formatTime(ms int64) string {
	t := time.Unix(0, ms*int64(time.Millisecond))
	return t.Format("2006-01-02 15:04:05")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func faintStr(s string) string {
	return color.New(color.Faint).Sprint(s)
}

func wrap(text string, width int, indent string) string {
	words := strings.Fields(text)
	var lines []string
	current := ""
	for _, w := range words {
		if len(current)+len(w)+1 > width {
			lines = append(lines, current)
			current = w
		} else {
			if current == "" {
				current = w
			} else {
				current += " " + w
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return strings.Join(lines, "\n"+indent)
}
