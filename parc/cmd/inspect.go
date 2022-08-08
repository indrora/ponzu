/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Investigate the contents of a Pitch archive",
	Long: `Investigate and show the structure of the Pitch archive,
including compression information and similar. `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("inspect called")
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// inspectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// inspectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
