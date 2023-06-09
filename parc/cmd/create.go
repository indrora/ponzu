/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
)

func createMain(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(cmd.ErrOrStderr(), "Expected 2 arguments, at least")
		return
	}

	// we're going to just list the files out so far

	archiveFname := args[0]
	archivePaths := args[1:]

	prefix := cmd.Flag("prefix")
	comment := cmd.Flag("comment")

	files := make(map[string]string)

	fmt.Printf("archive name = \"%v\", prefix = \"%v\", comment = \"%v\"\n", archiveFname, prefix.Value, comment.Value)
	for _, pathn := range archivePaths {
		fstat, err := os.Stat(pathn)
		if err != nil {
			cmd.PrintErrln(err)
		}

		// if it's a directory, get the files under it
		if fstat.IsDir() {

			//prefix := filepath.Clean(pathn)

			fs.WalkDir(os.DirFS("."), pathn, func(path string, d fs.DirEntry, err error) error {
				fmt.Println(path)

				return nil
			})

		}
	}

	for k, v := range files {
		fmt.Printf("%s -> %v", k, v)
	}

}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Ponzu archive",
	Long: `Create an archive from a specified set of content roots.

	Directories will have their contents added. Two paths that contain the same content will be merged.
	Files are added by their base name (foo/bar.txt -> bar.txt)

	Paths containing 

`,
	Run:     createMain,
	Example: "parc create myarchive.pzarc a/ foo",
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().String("comment", "", "Add comment to archive")
	createCmd.Flags().String("prefix", "", "Archive prefix")
}
