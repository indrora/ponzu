/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"errors"
	"os"

	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/reader"
	"github.com/spf13/cobra"
)

var (
	ErrMissingHeader = errors.New("archive is missing start control record")
	ErrBadMetadata   = errors.New("failed to cast")
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
	fh, err := os.OpenFile(args[0], os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		cmd.PrintErrln("Failed to open file:", err)
	}
	defer fh.Close()

	r := reader.NewReader(fh)

	// Get information about the archive

	var cSOA *format.StartOfArchive

	// make sure it's nil at the start.
	cSOA = nil

	walkFun := func(p *format.Preamble, m *format.RecordInfo) error {

		if cSOA == nil {
			if p.Rtype != format.RECORD_TYPE_CONTROL && p.Flags != format.RECORD_FLAG_CONTROL_START {
				return ErrMissingHeader
			} else {
				cSOA = &(m.StartOfArchive)
				// patch up the prefix if we have a change
				if forcedPrefix != nil {
					cmd.Println("Overriding prefix with " + *forcedPrefix)
					cSOA.Prefix = forcedPrefix
				}
			}

			cmd.Printf("Unpacking archive (version %v) with prefix %v \n", cSOA.Version, cSOA.Prefix)

		} else {

			switch p.Rtype {

			case format.RECORD_TYPE_FILE:
				filename := m.File.Name
				filesize := m.File.Metadata.FileSize
				cmd.Printf("Extract %s (size %s)\n", filename, filesize)

			case format.RECORD_TYPE_DIRECTORY:
				filename := m.Directory.Name
				cmd.Printf("Create directory %s\n", filename)
			case format.RECORD_TYPE_CONTROL:
				if p.Flags == format.RECORD_FLAG_CONTROL_START {
					cmd.PrintErrln("Encountered a start control record out of sequence.")
					return ErrMissingHeader
				}
				if p.Flags == format.RECORD_FLAG_CONTROL_END {
					cmd.Println("End of archive record found.")
					cSOA = nil
					return nil
				}
			case format.RECORD_TYPE_OS_SPECIAL:

				specialType := m.OSSpecial.SpecialType

				path := m.OSSpecial.Name

				if *specialType == "mknod" {

					device := m.OSSpecial.Device
					mode := m.OSSpecial.Mode
					cmd.Printf("mknod device %s (dev=%v, mode=%v)\n", path, device, mode)
				}
			case format.RECORD_TYPE_CONTINUE:
				return nil
			case format.RECORD_TYPE_HARDLINK:
				path := m.Hardlink.Name
				target := m.Hardlink.Target
				cmd.Printf("Hard link: %s -> %s\n", path, target)
			case format.RECORD_TYPE_SYMLINK:
				path := m.Symlink.Name
				target := m.Symlink.Target
				cmd.Printf("Symlink: %s -> %s\n", path, target)
			default:
				cmd.PrintErrln("Encountered unknown record type... skipping")
			}
		}

		return nil
	}

	err = r.Walk(walkFun)

	if err != nil {
		panic(err)
	}

}

var forcedPrefix *string

func init() {
	rootCmd.AddCommand(extractCmd)
	forcedPrefix = extractCmd.Flags().String("force-prefix", "", "Force the specified prefix")
	extractCmd.Flags().String("root", "", "Extract to specified root path (in addition to prefix)")
}
