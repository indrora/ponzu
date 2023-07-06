/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"errors"
	"fmt"
	"io"
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

			fmt.Println(filename)
			fileh, err := os.Open(filename)
			if err != nil {
				fmt.Println(err)
				continue
			}
			archiveReader := reader.NewReader(fileh)

			err = nil
			for !errors.Is(err, io.EOF) {

				var preamble *format.Preamble
				var meta any

				preamble, meta, err = archiveReader.Next()

				if err != nil && !errors.Is(err, io.EOF) {
					fmt.Println("Failed to read record header:")
					fmt.Println(err)
					fileh.Close()
					os.Exit(1)
				}
				if !errors.Is(err, io.EOF) {

					if preamble != nil {
						explainRecord(*preamble, meta)
					} else {
						fmt.Printf("Preamble was nil... Something went wrong")
						fileh.Close()
						os.Exit(1)
					}
				} else {
					fmt.Println("Reached end of file.")
				}
			}

			fileh.Close()
		}

	},
}

func explainRecord(preamble format.Preamble, meta any) {

	fmt.Printf("======Record ======\n")
	fmt.Printf("Type: %d\n", preamble.Rtype)
	fmt.Printf("Flags: %d\n", preamble.Flags)
	fmt.Printf("Compression: %d\n", preamble.Compression)
	fmt.Printf("Length: %d, modulo %d\n", preamble.DataLen, preamble.Modulo)
	fmt.Printf("Checksum: %x\n", preamble.DataChecksum)
	fmt.Printf("Metadata Length: %d\n", preamble.MetadataLength)
	fmt.Printf("Metadata Checksum: %x\n", preamble.MetadataChecksum)

	if meta != nil {
		spew.Dump(meta)
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

func castOrNil[T any](k *any) *T {
	if k == nil {
		return nil
	}
	v, ok := (*k).(T)
	if !ok {
		return nil
	} else {
		return &v
	}
}
