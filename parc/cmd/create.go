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

	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/writer"
	"github.com/spf13/cobra"

	"github.com/bmatcuk/doublestar/v4"
)

func getFiles(relroot string, pathn string) (map[string]string, error) {

	pathn = filepath.ToSlash(pathn)

	if !doublestar.ValidatePathPattern(pathn) {
		fmt.Println("Invalid path pattern: ", pathn)
		return nil, nil
	}

	mid, pattern := doublestar.SplitPattern(pathn)

	absroot := filepath.Join(relroot, mid)
	dir_fs := os.DirFS(absroot)

	foundpaths, err := doublestar.Glob(dir_fs, pattern)

	if err != nil {
		return nil, err
	}

	files := make(map[string]string, len(foundpaths))
	for _, path := range foundpaths {
		archivePath := filepath.Clean(filepath.Join(mid, path))
		abspath, err := filepath.Abs(filepath.Join(relroot, mid, path))
		if err != nil {
			return nil, errors.Join(errors.New("failed to get absolute path for "+abspath), err)
		}
		files[archivePath] = abspath
	}

	return files, nil
}

func createMain(cmd *cobra.Command, args []string) {
	verbose, _ = rootCmd.Flags().GetBool("verbose")

	if len(args) < 2 {
		fmt.Fprintln(cmd.ErrOrStderr(), "Expected 2 arguments, at least")
		return
	}

	archiveFname := args[0]
	archivePaths := args[1:]

	prefix, _ := cmd.Flags().GetString("prefix")
	comment, _ := cmd.Flags().GetString("comment")
	relroot, _ := cmd.Flags().GetString("chdir")

	files := make(map[string]string)

	if verbose {
		fmt.Printf("archive name = \"%v\", prefix = \"%v\", comment = \"%v\", searchroot=\"%v\"\n", archiveFname, prefix, comment, relroot)
	}
	for _, pathn := range archivePaths {
		nfiles, err := getFiles(relroot, pathn)
		if err != nil {
			cmd.PrintErr(err)
		} else {
			for lname, rname := range nfiles {
				files[lname] = rname
			}
		}
	}

	// open the archive
	fhandle, err := os.OpenFile(archiveFname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		cmd.PrintErr(err)
		return
	}
	defer fhandle.Close()
	writer := writer.NewWriter(fhandle, (*BuffSize)*format.BLOCK_SIZE)

	writer.AppendStart(prefix, comment)

	zstdDict, _ := cmd.Flags().GetString("zstandard-dictionary")
	if zstdDict != "" {
		// try and open the file
		dict, err := os.Open(zstdDict)
		if err != nil {
			cmd.PrintErr(err)
			return
		}
		buff := new(bytes.Buffer)
		_, err = io.Copy(buff, dict)
		if err != nil {
			cmd.PrintErr(err)
			return
		}
		dictBytes := buff.Bytes()

		dict.Close()
		writer.AppendZstdDict(dictBytes)
	}

	archive_files := make([]string, 0, len(files))
	for k := range files {
		archive_files = append(archive_files, k)
	}
	// sort the keys for deterministic output
	sort.Strings(archive_files)

	for _, archiveFilePath := range archive_files {
		localFilePath := files[archiveFilePath]

		fmt.Printf("%s -> %v\n", archiveFilePath, localFilePath)

		mask := os.ModeDir | os.ModeSymlink

		statn, err := os.Lstat(localFilePath)
		if err == nil {
			switch mode := statn.Mode(); mode & mask {
			case os.ModeDir:
				if verbose {
					fmt.Println("Directory")
				}
				writer.AppendDirectory(archiveFilePath, statn)
			case os.ModeSymlink:
				linkinfo, err := os.Readlink(localFilePath)
				if err != nil {
					cmd.PrintErrf("Failed to read symlink %v: %v\n", localFilePath, err)
				} else {
					if verbose {
						fmt.Printf("Symlink to %v\n", linkinfo)
					}
					writer.AppendSymlink(archiveFilePath, linkinfo, statn)
				}
			default:
				if verbose {
					fmt.Printf("Regular file, size=%v, modtime=%v\n", statn.Size(), statn.ModTime())
				}

				compression := format.COMPRESSION_NONE
				if !*NoCompress && statn.Size() > int64(format.BLOCK_SIZE) {
					if *UseBrotli {
						compression = format.COMPRESSION_BROTLI
					} else {
						compression = format.COMPRESSION_ZSTD
					}
				} else {
					if verbose {
						fmt.Println("File is smaller than single block, not compressing.")
					}
				}

				if err = writer.AppendFile(archiveFilePath, localFilePath, compression, statn); err != nil {
					cmd.PrintErr(err)
					return
				}
			}
		} else {
			cmd.PrintErrf("Failed to stat file: %v", err)
		}
	}
	writer.AppendEnd()
	fhandle.Close()

}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
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
	Args:    cobra.MinimumNArgs(1),
}

var BuffSize *uint64
var UseBrotli *bool
var NoCompress *bool
var verbose bool

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().String("comment", "", "Add comment to archive")
	createCmd.Flags().String("prefix", "", "Archive prefix")
	BuffSize = createCmd.Flags().Uint64("buff-size", 5000, "Number of blocks to read into memory at once (default 5000, 2GB)")
	createCmd.Flags().String("chdir", ".", "Search this path to find relative paths")
	NoCompress = createCmd.Flags().Bool("no-compress", false, "Disable compression")
	UseBrotli = createCmd.Flags().Bool("brotli", false, "use Brotli compression vs. ZStandard")
	createCmd.Flags().String("zstandard-dictionary", "", "Path to ZStandard Dictionary to use")
}
