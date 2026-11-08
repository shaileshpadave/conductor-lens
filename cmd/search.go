package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/shaileshpadave/conductor-lens/internal/conductor"
	"github.com/shaileshpadave/conductor-lens/internal/display"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search recent failed workflows",
	Example: `  conductor-lens search
  conductor-lens search --workflow-name payment_flow --limit 20`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("workflow-name")
		limit, _ := cmd.Flags().GetInt("limit")

		client := conductor.NewClient(conductorURL, conductorKey)

		display.PrintSpinner(fmt.Sprintf("Searching failed workflows on %s", conductorURL))
		results, err := client.SearchFailed(name, limit)
		if err != nil {
			display.PrintError(err.Error())
			return nil
		}

		display.PrintSearchResults(results.Results)
		return nil
	},
}

func init() {
	searchCmd.Flags().String("workflow-name", "", "Filter by workflow name (optional)")
	searchCmd.Flags().Int("limit", 10, "Max results to return")
	rootCmd.AddCommand(searchCmd)
}
