/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/jvfrodrigues/realtor-info-getter/pkg/c2s"
	"github.com/spf13/cobra"
)

// c2sCmd represents the c2s command
var c2sCmd = &cobra.Command{
	Use:   "c2s",
	Short: "Get leads from c2s",
	Long:  `Get all leads from c2s`,
	Run: func(cmd *cobra.Command, args []string) {
		c2s.GetC2S()
	},
}

func init() {
	rootCmd.AddCommand(c2sCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// c2sCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// c2sCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
