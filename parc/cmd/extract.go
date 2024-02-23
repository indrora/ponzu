/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Unwrap a Ponzu archive",
	Long:  `Unwrap a given archive to the given path (default ".")`,
	Run:   run,
	Args:  cobra.ExactArgs(1),
}

func run(cmd *cobra.Command, args []string) {

	if len(args) != 1 {
		cmd.PrintErrln("Expected 1 argument, got something else.")
		return
	}

}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().String("force-prefix", "", "Force the specified prefix")
}
