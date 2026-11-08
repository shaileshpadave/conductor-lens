package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	conductorURL string
	conductorKey string
	anthropicKey string
)

var rootCmd = &cobra.Command{
	Use:   "conductor-lens",
	Short: "AI-powered failure analyst for Conductor OSS workflows",
	Long: `conductor-lens connects to your Conductor OSS instance,
fetches workflow executions, and uses AI to diagnose failures —
giving you a root cause, explanation, and concrete fix.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&conductorURL, "conductor-url", envOrDefault("CONDUCTOR_URL", "http://localhost:8080"), "Conductor OSS base URL")
	rootCmd.PersistentFlags().StringVar(&conductorKey, "conductor-key", os.Getenv("CONDUCTOR_API_KEY"), "Conductor API key (optional)")
	rootCmd.PersistentFlags().StringVar(&anthropicKey, "anthropic-key", os.Getenv("ANTHROPIC_API_KEY"), "Anthropic API key")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
