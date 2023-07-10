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
	"github.com/indrora/ponzu/ponzu/format/metadata"
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
			inspectArchive(filename)
		}
	},
}

func inspectArchive(path string) {

	verbose, _ := rootCmd.Flags().GetBool("verbose")

	fileh, err := os.Open(path)
	if err != nil {
		return
	}
	defer fileh.Close()
	archiveReader := reader.NewReader(fileh)

	err = nil
	for !errors.Is(err, io.EOF) {

		var preamble *format.Preamble
		var meta any

		preamble, meta, err = archiveReader.Next()

		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Println("Failed to read record header:")
			fmt.Println(err)
			return
		}
		if !errors.Is(err, io.EOF) {

			if preamble != nil {
				if verbose {
					fmt.Println("Preamble:")
					spew.Dump(preamble)
					fmt.Println("Metadata:")
					spew.Dump(meta)
				} else {
					explainRecord(*preamble, meta)
				}
			} else {
				fmt.Printf("Preamble was nil... Something went wrong")

				return

			}
		} else {
			fmt.Println("No more records.")
		}
	}
}

func explainRecord(preamble format.Preamble, meta any) {

	switch preamble.Rtype {
	case format.RECORD_TYPE_CONTROL:
		fmt.Print("Control record: ")
		if preamble.Flags == format.RECORD_FLAG_CONTROL_START {
			fmt.Println("Begin archive.", "ponzu version", meta.(*format.StartOfArchive).Version)
		} else if preamble.Flags == format.RECORD_FLAG_CONTROL_END {
			fmt.Println("End of archive marker")
		} else {
			fmt.Println("Unknown control record.")
		}
	case format.RECORD_TYPE_DIRECTORY:
		fmt.Println("Directory: ", meta.(*format.Directory).Name)
	case format.RECORD_TYPE_FILE:
		fmeta := meta.(*format.File)
		mmeta, ok := metadata.TransmogrifyCbor[metadata.CommonMetadata](fmeta.Metadata.(map[any]any))
		fmt.Println("File ", fmeta.Name, "modtime ", fmeta.ModTime)

		if verbose {

			if ok {
				if mmeta.FileSize != nil {
					fmt.Printf("Size %d bytes\n", *mmeta.FileSize)
				}
				if mmeta.MimeType != nil {
					fmt.Printf("Mimetype %s\n", *mmeta.MimeType)
				}
			}
			fmt.Printf("Body checksum: %x\n", preamble.DataChecksum)
		}

	case format.RECORD_TYPE_CONTINUE:
		fmt.Println("[Previous record continues]")
	default:
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
		fmt.Println("======= Record ===== ")
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
