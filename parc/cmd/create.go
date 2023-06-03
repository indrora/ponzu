/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Ponzu archive",
	Long: `Create an archive from a specified set of paths.


	
`,
	Run: func(cmd *cobra.Command, args []string) {
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
			// walk every file in those paths.

			// First, clean the path:

			pathn_clean := filepath.Clean(pathn)
			fmt.Println(pathn_clean)

			// check to see if it's a glob

			if strings.ContainsRune(pathn, '*') {
				// Get the base of the path
				fmt.Println("Path is a glob")

				path_base, _ := strings.CutSuffix(pathn, "*")
				if strings.HasSuffix(path_base, string(filepath.Separator)) {
					path_base = filepath.Dir(path_base)
				}

			} else {

				pathstat, err := os.Lstat(pathn)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Couldn't stat %v, %v\n", pathn, err)
					continue
				}

				if pathstat.IsDir() {
					fmt.Println("Found a directory: " + pathstat.Name())
				} else {

					if pathstat.Mode()&os.ModeSymlink != 0 {
						target, _ := os.Readlink(pathn)
						fmt.Printf("%v -> link to %v\n", pathn, target)
					} else {
						spew.Dump(pathstat)
					}

				}
			}

		}

		for k, v := range files {
			fmt.Printf("%s -> %v", k, v)
		}

	},
	Example: "parc create myarchive.pzarc a/ foo",
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().String("comment", "", "Add comment to archive")
	createCmd.Flags().String("prefix", "", "Archive prefix")
}
