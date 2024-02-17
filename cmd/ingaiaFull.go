/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/jvfrodrigues/realtor-info-getter/pkg/ingaia"
	"github.com/spf13/cobra"
)

// ingaiaFullCmd represents the ingaiaFull command
var ingaiaFullCmd = &cobra.Command{
	Use:   "ingaiaFull",
	Short: "Gets all properties from ingaia",
	Long:  `Access ingaia and get all properties`,
	Run: func(cmd *cobra.Command, args []string) {
		ingaia.GetIngaia(args)
	},
}

func init() {
	rootCmd.AddCommand(ingaiaFullCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ingaiaFullCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ingaiaFullCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
