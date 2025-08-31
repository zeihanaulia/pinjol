package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "pinjol",
	Short: "Pinjol service CLI",
	Long:  `Pinjol is a loan management service with CLI and server capabilities.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Configure Viper to read environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Add subcommands here
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(dbInitCmd)
	rootCmd.AddCommand(scenarioCmd)
}
