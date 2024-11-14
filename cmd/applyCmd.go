package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply Topology",
	Long:  `Apply Topology with Nodes list and Links list.`,
	Run: func(cmd *cobra.Command, args []string) {
		filepath, _ := cmd.Flags().GetString("from")
		err := Calculator.ApplyTopoConfig(filepath)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringP("from", "f", "", "Path to the topology configuration file")
	//applyCmd.MarkFlagRequired("from")
}
