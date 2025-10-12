/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"errors"
	"io"
	"os"
	"path"

	"github.com/indrora/ponzu/eflag"
	"github.com/indrora/ponzu/osmeta"
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

	fullPath := path.Join(*extractCmdOpts.ExtractPath, *extractCmdOpts.Prefix)

	_ = os.MkdirAll(fullPath, os.FileMode(*extractCmdOpts.FileMode))

	walkFun := func(p *format.Preamble, m *format.RecordInfo) error {

		localLogger.Debug(
			"Processing record",
			zap.Binary("infoHash", p.InfoChecksum[:]),
			zap.Binary("dataHash", p.DataChecksum[:]),
			zap.Uint64("blocks", p.DataLen),
		)

		fMode := GetFileMode(m, *extractCmdOpts.FileMode)

		switch p.Rtype {

		case format.RECORD_TYPE_FILE:
			filename := m.File.Name
			filesize := m.File.Metadata.FileSize

			localLogger.Info("Extract regular file",
				zap.Stringp("filename", filename),
				zap.Uint64p("filesize", filesize),
			)
			diskname := path.Join(fullPath, *filename)

			// Try and get the directory that this is in
			checkDir, _ := path.Split(diskname)

			var fHandle *os.File

			if chkstat, err := os.Stat(checkDir); err != nil {
				err = os.MkdirAll(checkDir, os.FileMode(*extractCmdOpts.FileMode))
				if err != nil {
					localLogger.Fatal("Couldn't create directories", zap.Error(err))
				}
			} else {
				if !chkstat.IsDir() {
					localLogger.Fatal("target path exists and is not a directory!", zap.String("dir", checkDir))
				}
			}

			fHandle, err = os.OpenFile(diskname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

			if err != nil {
				localLogger.Fatal("Failed to open output file", zap.Error(err))
			}
			err = r.CopyAll(fHandle, false)
			if err != nil && err != io.EOF {
				localLogger.Fatal("Failed to write output", zap.Error(err))
			}
			err = fHandle.Close()
			if err != nil {
				localLogger.Fatal("Failed to close output file", zap.Error(err))
			}

			if err = osmeta.SetMetadata(diskname, m.File.Metadata); err != nil {
				localLogger.Fatal("Failed setting metadata", zap.Error(err))
			}

		case format.RECORD_TYPE_DIRECTORY:
			filename := m.Directory.Name
			diskname := path.Join(fullPath, *filename)
			// Check if the directory already exists.

			localLogger.Info("Create directory", zap.Stringp("path", filename))
			if err = os.Mkdir(diskname, os.FileMode(fMode)); err != nil && !errors.Is(err, os.ErrExist) {
				localLogger.Fatal("failed to create directory", zap.Error(err))
			}
			if err = osmeta.SetMetadata(diskname, m.Directory.Metadata); err != nil {
				localLogger.Fatal("Failed to set OS metadata", zap.Error(err))
			}

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
			source := m.Hardlink.Name
			target := m.Hardlink.Target
			localLogger.Info("Hardlink", zap.Stringp("path", source), zap.Stringp("target", target))

			srcpath := path.Join(fullPath, *source)
			targetpath := path.Join(fullPath, *target)

			if err = os.Link(targetpath, srcpath); err != nil {
				localLogger.Fatal("Failed to link", zap.Error(err))
			}

		case format.RECORD_TYPE_SYMLINK:
			source := m.Symlink.Name
			target := m.Symlink.Target
			localLogger.Info("Symlink", zap.Stringp("path", source), zap.Stringp("target", target))

			srcpath := path.Join(fullPath, *source)
			targetpath := path.Join(fullPath, *target)

			if err = os.Link(targetpath, srcpath); err != nil {
				localLogger.Fatal("Failed to link", zap.Error(err))
			}

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
	FileMode       *uint32
	OverrideMode   *bool
}

var extractCmdOpts = ExtractOptions{FilterMode: FilterModeInclude}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmdOpts.Prefix = extractCmd.Flags().String("prefix", "", "Force the specified prefix")
	extractCmdOpts.ExtractPath = extractCmd.Flags().String("path", ".", "Extract to specified root path (in addition to prefix)")
	extractCmdOpts.ShouldFilter = extractCmd.Flags().Bool("filter", false, "Filter paths")
	extractCmdOpts.FilterPatterns = extractCmd.Flags().StringArray("filter-pattern", []string{}, "Pattern to include/exclude from extraction")
	extractCmd.Flags().Var(eflag.NewEnumFlag(&extractCmdOpts.FilterMode, FilterModeInclude, "mode", FilterModeMap), "filter-mode", "Filter direction: include/exclude")
	extractCmdOpts.FileMode = extractCmd.Flags().Uint32("file-mode", 0755, "Specify file mode (fallback). Use leading 0 for octal modes.")
}
