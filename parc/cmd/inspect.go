/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

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
		metastruct, ok := meta.(map[interface{}]interface{})
		if ok {
			metajson, _ := json.MarshalIndent(metastruct, "", "  ")
			fmt.Printf("%v", metajson)
		}

		/*

				if meta != nil {

					switch preamble.Rtype {
					case format.RECORD_TYPE_CONTROL:
						fmt.Println("-- Control record Metadata -- ")
						if preamble.Flags == format.RECORD_FLAG_CONTROL_START {
							if soa := castOrNil[format.StartOfArchive](&meta); soa != nil {
								fmt.Printf("Prefix: %s\nComment:%s\n", soa.Prefix, soa.Comment)
								fmt.Printf("Host OS: %s\n", soa.Host)
								fmt.Printf("Archive version: %d", soa.Version)
							}
						}
					case format.RECORD_TYPE_FILE:
						fmt.Println("-> File metadata")
						if file := castOrNil[format.File](&meta); file != nil {
							fmt.Printf("File Path: %s\n", file.Name)
							fmt.Printf("File ModTime: %s\n", file.ModTime)
						} else {
							encoder := json.NewEncoder(os.Stdout)
							encoder.Encode(meta)
						}
					case format.RECORD_TYPE_DIRECTORY:
			v			if dir := castOrNil[format.Directory](&meta); dir != nil {
							fmt.Printf("Directory Path: %s\n", dir.Name)
						} else {
							encoder := json.NewEncoder(os.Stdout)
							encoder.Encode(meta)
						}
					case format.RECORD_TYPE_SYMLINK:
						fmt.Println("-> symlink")
						if sym := castOrNil[format.Symlink](&meta); sym != nil {
							fmt.Printf("Symlink Path: %s -> %s\n", sym.Name, sym.Target)
						} else {
							encoder := json.NewEncoder(os.Stdout)
							encoder.Encode(meta)
						}
					case format.RECORD_TYPE_CONTINUE:
						// No metadata for continue records.
					default:
						fmt.Printf("Unknown record type %d\n", preamble.Rtype)
						// turn the metadata into json
						encoder := json.NewEncoder(os.Stdout)
						encoder.Encode(meta)
					}
				} */
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
