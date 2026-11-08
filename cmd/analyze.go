package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/shaileshpadave/conductor-lens/internal/analyzer"
	"github.com/shaileshpadave/conductor-lens/internal/conductor"
	"github.com/shaileshpadave/conductor-lens/internal/display"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a failed workflow execution",
	Example: `  conductor-lens analyze --workflow-id abc123
  conductor-lens analyze --workflow-id abc123 --conductor-url http://my-conductor:8080`,
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowID, _ := cmd.Flags().GetString("workflow-id")
		if workflowID == "" {
			return fmt.Errorf("--workflow-id is required")
		}
		if anthropicKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY not set (or use --anthropic-key)")
		}

		client := conductor.NewClient(conductorURL, conductorKey)
		ai := analyzer.New(anthropicKey)

		display.PrintSpinner(fmt.Sprintf("Fetching workflow %s from %s", workflowID, conductorURL))
		wf, err := client.GetWorkflow(workflowID)
		if err != nil {
			display.PrintError(err.Error())
			return nil
		}

		display.PrintWorkflowSummary(wf)

		display.PrintSpinner("Analyzing with Claude")
		analysis, err := ai.Analyze(context.Background(), wf)
		if err != nil {
			display.PrintError(err.Error())
			return nil
		}

		display.PrintAnalysis(analysis)
		return nil
	},
}

func init() {
	analyzeCmd.Flags().String("workflow-id", "", "Conductor workflow execution ID to analyze")
	rootCmd.AddCommand(analyzeCmd)
}
