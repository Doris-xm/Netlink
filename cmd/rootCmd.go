package cmd

import (
	"Netlink/pkg"
	"github.com/spf13/cobra"
)

var Calculator *pkg.Calculator
var rootCmd = &cobra.Command{
	Use:   "net",
	Short: "net Management CLI",
	Long:  "A command-line tool for managing network topologies.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(c *pkg.Calculator) error {
	Calculator = c
	err := rootCmd.Execute()
	return err
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.AddCommand(applyCmd)
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
