package cmd

import (
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show Resources",
	Long:  `Show the resources of the topology.`,
	Run: func(cmd *cobra.Command, args []string) {
		class := cmd.Flag("class").Value.String()
		if class == "nodes" {
			Calculator.ShowNodes()
		} else if class == "links" {
			Calculator.ShowLinks()
		} else {
			print("Invalid class")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().String("class", "nodes", "Class of the element to show")
}
