/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Unwrap a Ponzu archive",
	Long:  `Unwrap a given archive to the given path (default ".")`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("extract called")
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().String("force-prefix", "", "Force the specified prefix")
}
