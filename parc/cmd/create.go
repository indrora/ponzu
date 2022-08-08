/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Pitch archive",
	Long: `Create an archive from a specified set of paths.
	
example:

parc create myarchive.pitch a/* b/*`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create called")

	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().BoolP("stream", "s", false, "Create streamed (noncompressed) archives")
	createCmd.Flags().BoolP("no-checksum", "n", false, "Skip checksum calculation")
	createCmd.Flags().String("comment", "", "Add comment to archive")
}
