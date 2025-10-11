/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"errors"
	"os"

	"github.com/indrora/ponzu/eflag"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/reader"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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

	localLogger := GlobalLogger.With(zap.String("archive", args[0]))

	fh, err := os.OpenFile(args[0], os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		localLogger.Fatal("Failed to open archive", zap.Error(err))
	}
	defer fh.Close()

	r := reader.NewReader(fh)

	// Get information about the archive

	mPreamble, mRecord, err := r.Next()

	if err != nil {
		localLogger.Fatal("Failed to read record", zap.Error(err))
	}

	// Verify that the first record we read is a control record and that it's a start flag.
	if mPreamble.Rtype == format.RECORD_TYPE_CONTROL && mPreamble.Flags == format.RECORD_FLAG_CONTROL_START {
		coa := mRecord.StartOfArchive

		// Do we need to override the prefix, or do we use the one from the COA?
		overridePrefix := cmd.Flags().Changed("prefix")
		if !overridePrefix {
			extractCmdOpts.Prefix = coa.Prefix
		}

		// Say something
		localLogger.Info("unpacking archive",
			zap.String("prefix", *extractCmdOpts.Prefix),
			zap.String("comment", *coa.Comment),
			zap.String("host", *coa.Host),
			zap.Int("version", *coa.Version),
			zap.Any("options", extractCmdOpts),
		)

	} else {
		// The first record was the wrong type, bail!
		localLogger.Fatal(
			"First record was not start of archive",
			zap.Uint8("rtype", uint8(mPreamble.Rtype)),
			zap.Uint8("flags", uint8(mPreamble.Flags)),
		)
	}

	walkFun := func(p *format.Preamble, m *format.RecordInfo) error {

		localLogger.Debug(
			"Processing record",
			zap.Binary("infoHash", p.InfoChecksum[:]),
			zap.Binary("dataHash", p.DataChecksum[:]),
			zap.Uint64("blocks", p.DataLen),
		)
		switch p.Rtype {

		case format.RECORD_TYPE_FILE:
			filename := m.File.Name
			filesize := m.File.Metadata.FileSize

			localLogger.Info("Extract regular file",
				zap.Stringp("filename", filename),
				zap.Uint64p("filesize", filesize),
			)

		case format.RECORD_TYPE_DIRECTORY:
			filename := m.Directory.Name
			localLogger.Info("Create directory", zap.Stringp("path", filename))
		case format.RECORD_TYPE_CONTROL:
			if p.Flags == format.RECORD_FLAG_CONTROL_START {
				localLogger.Fatal("Control record out of sequence.")
				return ErrMissingHeader
			}
			if p.Flags == format.RECORD_FLAG_CONTROL_END {

				localLogger.Info("Found end of archive record ")
				return nil
			}
		case format.RECORD_TYPE_OS_SPECIAL:

			specialType := m.OSSpecial.SpecialType

			path := m.OSSpecial.Name

			if *specialType == "mknod" {

				device := m.OSSpecial.Device
				mode := m.OSSpecial.Mode
				localLogger.Info("mknod device", zap.Stringp("path", path), zap.Uint32p("device", device), zap.Uint32p("mode", mode))
			} else {
				localLogger.Fatal("Unknown special type", zap.Stringp("type", specialType))
			}
		case format.RECORD_TYPE_CONTINUE:
			return nil
		case format.RECORD_TYPE_HARDLINK:
			path := m.Hardlink.Name
			target := m.Hardlink.Target
			localLogger.Info("Hardlink", zap.Stringp("path", path), zap.Stringp("target", target))
		case format.RECORD_TYPE_SYMLINK:
			path := m.Symlink.Name
			target := m.Symlink.Target
			localLogger.Info("symlink", zap.Stringp("path", path), zap.Stringp("target", target))
		default:
			localLogger.Warn("Unhandled record type type", zap.Uint8("type", uint8(p.Rtype)), zap.Any("info", m))
		}

		return nil
	}

	err = r.Walk(walkFun)

	if err != nil {
		panic(err)
	}

}

//var forcedPrefix *string

type FilterMode int

const (
	FilterModeInclude FilterMode = iota
	FilterModeExclude
)

var FilterModeMap = map[string]FilterMode{
	"include": FilterModeInclude,
	"exclude": FilterModeExclude,
}

type ExtractOptions struct {
	Prefix         *string
	ExtractPath    *string
	ShouldFilter   *bool
	FilterPatterns *[]string
	FilterMode     FilterMode
}

var extractCmdOpts = ExtractOptions{FilterMode: FilterModeInclude}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmdOpts.Prefix = extractCmd.Flags().String("prefix", "", "Force the specified prefix")
	extractCmdOpts.ExtractPath = extractCmd.Flags().String("path", ".", "Extract to specified root path (in addition to prefix)")
	extractCmdOpts.ShouldFilter = extractCmd.Flags().Bool("filter", false, "Filter paths")
	extractCmdOpts.FilterPatterns = extractCmd.Flags().StringArray("filter-pattern", []string{}, "Pattern to include/exclude from extraction")
	extractCmd.Flags().Var(eflag.NewEnumFlag(&extractCmdOpts.FilterMode, FilterModeInclude, "mode", FilterModeMap), "filter-mode", "Filter direction: include/exclude")
}
