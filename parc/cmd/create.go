/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/indrora/ponzu/eflag"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/writer"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/bmatcuk/doublestar/v4"
)

func getFiles(searchStart string, searchPattern string) (map[string]string, error) {

	searchPattern = filepath.ToSlash(searchPattern)

	if !doublestar.ValidatePathPattern(searchPattern) {
		//GlobalLogger.Panic("Invalid search pattern", zap.String("pattern", pathn))
		return nil, fmt.Errorf("invalid search pattern %s", searchPattern)
	}

	mid, pattern := doublestar.SplitPattern(searchPattern)

	combinedSearchPath := filepath.Join(searchStart, mid)
	searchFS := os.DirFS(combinedSearchPath)

	foundPaths, err := doublestar.Glob(searchFS, pattern)

	if err != nil {
		return nil, err
	}

	files := make(map[string]string, len(foundPaths))
	for _, path := range foundPaths {
		archivePath := filepath.Clean(filepath.Join(mid, path))
		abspath, err := filepath.Abs(filepath.Join(searchStart, mid, path))
		if err != nil {
			return nil, errors.Join(errors.New("failed to get absolute path for "+abspath), err)
		}
		files[archivePath] = abspath
	}

	return files, nil
}

func createMain(cmd *cobra.Command, args []string) {

	// Check that we have enough arguments
	if len(args) < 2 {
		fmt.Fprintln(cmd.ErrOrStderr(), "Expected 2 arguments, at least")
		cmd.Usage()
		return
	}

	// some things we use
	archiveFilename := args[0]
	archiveSearchPaths := args[1:]

	// Search for the files in the 	search paths
	files := make(map[string]string)
	for _, searchPath := range archiveSearchPaths {
		foundFiles, err := getFiles(*createCmdOpts.SearchPath, searchPath)
		if err != nil {
			GlobalLogger.Fatal("invalid path specifier", zap.String("pattern", searchPath), zap.Error(err))
		} else {
			for localName, archiveName := range foundFiles {
				files[localName] = archiveName
			}
		}
	}

	GlobalLogger.Info("Done collecting files", zap.Int("count", len(files)))

	// If the archive will be empty, say so.
	if (len(files)) < 1 {
		GlobalLogger.Warn("Archive will contain no records!")
	}

	// Copy the list of files over to a new array so we can sort them
	archive_files := make([]string, 0, len(files))
	for k := range files {
		archive_files = append(archive_files, k)
	}

	// sort the keys for deterministic output
	sort.Strings(archive_files)

	// open the archive file. We want to make sure that we're creating, truncating, and creating a one way handle
	fhandle, err := os.OpenFile(archiveFilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		GlobalLogger.Fatal("failed to open output", zap.String("path", archiveFilename), zap.Error(err))
		return
	}

	GlobalLogger.Info("Creating archive",
		zap.String("name", archiveFilename),
		zap.String("prefix", *createCmdOpts.ArchivePrefix),
		zap.String("comment", *createCmdOpts.ArchiveComment),
		zap.String("root", *createCmdOpts.SearchPath),
	)
	writer := writer.NewWriter(fhandle, (*createCmdOpts.BuffSize)*format.BLOCK_SIZE)

	// Clean up after ourselves when we're done here.
	defer writer.AppendEnd()
	defer fhandle.Close()

	GlobalLogger.Info("Starting archive", zap.String("filename", archiveFilename))

	// Write start of archive header.
	writer.AppendStart(*createCmdOpts.ArchivePrefix, *createCmdOpts.ArchiveComment)

	//zstdDict, _ := cmd.Flags().GetString("zstandard-dictionary")

	if *createCmdOpts.ZstdDictPath != "" {
		GlobalLogger.Debug("Adding Zstandard dictionary", zap.String("filename", *createCmdOpts.ZstdDictPath))
		// try and open the file
		dict, err := os.Open(*createCmdOpts.ZstdDictPath)
		if err != nil {
			GlobalLogger.Fatal("Failed to open zstd dictionary", zap.String("filename", *createCmdOpts.ZstdDictPath), zap.Error(err))
			return
		}
		buff := new(bytes.Buffer)

		_, err = io.Copy(buff, dict)

		if err != nil {
			GlobalLogger.Fatal("Failed to read zstd dictionary", zap.String("filename", *createCmdOpts.ZstdDictPath), zap.Error(err))
			return
		}

		dictBytes := buff.Bytes()

		// we can't defer this call because that would defer at the end of the function, not here.
		dict.Close()
		GlobalLogger.Info("Appending ZStandard dictionary", zap.Int("size", len(dictBytes)))
		writer.AppendZstdDict(dictBytes)
	}

	GlobalLogger.Info("files collected", zap.Int("count", len(archive_files)))

	mask := os.ModeDir | os.ModeSymlink

	// Work through each of the files found in the search path
	for _, archiveFilePath := range archive_files {
		localFilePath := files[archiveFilePath]

		cFileStat, err := os.Lstat(localFilePath)
		if err != nil {
			GlobalLogger.Fatal("could not stat fle path", zap.String("path", localFilePath))
		}

		localLog := GlobalLogger.With(zap.String("localPath", localFilePath), zap.String("archivePath", archiveFilePath))

		localLog.Debug("got file stat", zap.Any("stat", cFileStat))

		switch mode := cFileStat.Mode(); mode & mask {
		case os.ModeDir:
			localLog.Info("Append Directory")
			writer.AppendDirectory(archiveFilePath, cFileStat)
		case os.ModeSymlink:
			linkinfo, err := os.Readlink(localFilePath)
			if err != nil {
				localLog.Fatal("Failed to read link information", zap.String("localPath", localFilePath), zap.Error(err))
			} else {
				localLog.Info("Append Symlink", zap.String("localPath", localFilePath), zap.String("linkInfo", linkinfo), zap.String("archivePath", archiveFilePath))
				writer.AppendSymlink(archiveFilePath, linkinfo, cFileStat)
			}
		default:
			localLog.Info("Append regular file")
			cRecordCompression := *(createCmdOpts.CompressionType)

			// If we are told to compress but the file is smaller than the size of a block, there is no reason to do so.
			// Warn that we're going to skip compression
			if cRecordCompression != format.COMPRESSION_NONE && cFileStat.Size() < int64(format.BLOCK_SIZE) {
				localLog.Warn("File is smaller than single block, not compressing", zap.Int64("size", cFileStat.Size()))
				cRecordCompression = format.COMPRESSION_NONE
			}

			if err = writer.AppendFile(archiveFilePath, localFilePath, cRecordCompression, cFileStat); err != nil {
				localLog.Fatal("Failed to append file to archive", zap.Error(err))
				return
			}
		}
	}

}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [flags] ARCHIVE ...FILES",
	Short: "Create a Ponzu archive",
	Long: `Create an archive from a specified series of glob patterns

Example globbing patterns:

* foo/ (Selects the directory "foo" but no contents)
* foo/* (Selects all contents of "foo")
* foo/** (Selects all contents of "foo" recursively)
* foo/*.txt (Selects all ".txt" files in "foo")
* foo/*/*.txt (Selects all ".txt" files in subdirectories of "foo")
* foo/*.{txt,md} (Selects all ".txt" and ".md" files in "foo")
* foo/{a,b,c}/* (Selects all contents of "foo/a", "foo/b" and "foo/c" non-recursively)

Use ? to specify a single character (foo/??/* selects all contents of two-character subdirectories of "foo")

Double stars act mostly like bash's globstar: **.txt is the same as *.txt, but foo/**/*.txt selects all .txt files in any depth subdirectory of foo.

Depending on your shell, you may have to enclose globbing patterns in single quotes('foo/**').
`,
	Run:     createMain,
	Example: "parc create myarchive.pzarc a/** foo",
	Args:    cobra.MinimumNArgs(2),
}

var compressTypes = map[string]format.CompressionType{
	"zstd":   format.COMPRESSION_ZSTD,
	"brotli": format.COMPRESSION_BROTLI,
	"none":   format.COMPRESSION_NONE,
}

type createOpts struct {
	BuffSize        *uint64
	CompressionType *format.CompressionType
	ForceCompress   *bool
	ArchiveComment  *string
	ArchivePrefix   *string
	ZstdDictPath    *string
	SearchPath      *string
}

var createCmdOpts = createOpts{
	CompressionType: new(format.CompressionType),
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Minutiae
	createCmdOpts.BuffSize = createCmd.Flags().Uint64("buff-size", 5000, "Number of blocks to read into memory at once (default 5000, 2GB)")
	createCmdOpts.SearchPath = createCmd.Flags().StringP("chdir", "C", ".", "Search this path to find relative paths")

	// Archive information
	createCmdOpts.ArchiveComment = createCmd.Flags().String("comment", "", "Add comment to archive")
	createCmdOpts.ArchivePrefix = createCmd.Flags().String("prefix", "", "Archive prefix")

	// Compression options

	createCmdOpts.ZstdDictPath = createCmd.Flags().String("zstandard-dictionary", "", "Path to ZStandard Dictionary to use")
	createCmd.Flags().Var(eflag.NewEnumFlag(createCmdOpts.CompressionType, format.COMPRESSION_ZSTD, "type", compressTypes), "compression", "Specify compression (none,zstd,brotli) to use")
}
