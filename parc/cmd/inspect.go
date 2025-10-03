/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/reader"
	"github.com/spf13/cobra"
)

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Investigate the contents of a Ponzu archive",
	Long: `Investigate and show the structure of the Ponzu archive,
including compression information and similar. `,
	Run: func(cmd *cobra.Command, args []string) {
		for _, filename := range args {
			inspectArchive(cmd, filename)
		}
	},
	Args: cobra.MinimumNArgs(1),
}

func inspectArchive(cmd *cobra.Command, path string) {

	//verbose, _ := rootCmd.Flags().GetBool("verbose")

	fh, err := os.OpenFile(path, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		cmd.PrintErrln("Failed to open file:", err)
	}
	defer fh.Close()

	r := reader.NewReader(fh)

	walkFun := func(p *format.Preamble, m *format.RecordInfo) error {
		fmt.Println("--- enter walkFun ---")
		spew.Dump(p)
		switch p.Rtype {
		case format.RECORD_TYPE_CONTROL:
			if p.Flags == format.RECORD_FLAG_CONTROL_START {
				spew.Dump(m.StartOfArchive)
			}
		}
		fmt.Println("--- exit walkFun ---")
		return nil
	}

	err = r.Walk(walkFun)

	if err != nil {
		panic(err)
	}
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
