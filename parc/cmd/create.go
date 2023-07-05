/*
Copyright © 2022 Morgan Gangwere <morgan.gangwere@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/writer"
	"github.com/spf13/cobra"
)

func fixLocalPath(path string) string {
	if filepath.IsAbs(path) {
		// strip off any front matter
		vol := filepath.VolumeName(path)
		return path[len(vol)+1:]
	} else {
		clean := filepath.Clean(path)
		// strip off any leading ..
		for strings.HasPrefix(clean, "..") {
			clean = clean[3:]
		}
		return clean
	}
}

func getFiles(relroot string, pathn string) (map[string]string, error) {

	files := make(map[string]string)
	var path string
	if filepath.IsAbs(pathn) {
		path = pathn
	} else {
		path = filepath.Join(relroot, pathn)
	}

	if containsGlob(pathn) {
		globbed, err := filepath.Glob(path)
		if err != nil {
			return nil, err
		}
		for _, gpath := range globbed {
			nfiles, err := getFiles(relroot, gpath)
			if err != nil {
				return nil, err
			}
			for k, v := range nfiles {
				files[k] = v
			}
		}
	} else {

		fi, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				abspath, err := filepath.Abs(path)
				if err != nil {
					return err
				}

				if filepath.IsAbs(pathn) {
					files[fixLocalPath(path)] = abspath
				} else {
					rel, err := filepath.Rel(relroot, path)
					if err != nil {
						return err
					}
					files[fixLocalPath(rel)] = abspath
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			abspath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}

			if filepath.IsAbs(pathn) {
				files[fixLocalPath(path)] = abspath
			} else {

				rel, err := filepath.Rel(relroot, path)
				if err != nil {
					return nil, err
				}
				files[fixLocalPath(rel)] = abspath
			}
		}
	}
	return files, nil
}

func containsGlob(ex string) bool {
	return strings.ContainsRune(ex, '*')
}

func createMain(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(cmd.ErrOrStderr(), "Expected 2 arguments, at least")
		return
	}

	archiveFname := args[0]
	archivePaths := args[1:]

	prefix := cmd.Flag("prefix")
	comment := cmd.Flag("comment")
	relroot := cmd.Flag("chdir")

	files := make(map[string]string)

	fmt.Printf("archive name = \"%v\", prefix = \"%v\", comment = \"%v\", searchroot=\"%v\"\n", archiveFname, prefix.Value, comment.Value, relroot.Value.String())
	for _, pathn := range archivePaths {
		nfiles, err := getFiles(relroot.Value.String(), pathn)
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

	writer.AppendStart(prefix.Value.String(), comment.Value.String())

	archive_files := make([]string, 0, len(files))
	for k := range files {
		archive_files = append(archive_files, k)
	}
	// sort the keys for deterministic output
	sort.Strings(archive_files)

	for _, archiveFilePath := range archive_files {
		localFilePath := files[archiveFilePath]

		fmt.Printf("%s -> %v: ", archiveFilePath, localFilePath)

		mask := os.ModeDir | os.ModeSymlink

		statn, err := os.Lstat(localFilePath)
		if err == nil {
			switch mode := statn.Mode(); mode & mask {
			case os.ModeDir:
				fmt.Println("Mode is directory")
				writer.AppendDirectory(archiveFilePath, statn)
			case os.ModeSymlink:
				fmt.Println("Mode is Symlink. ")
				linkinfo, err := os.Readlink(localFilePath)
				if err != nil {
					cmd.PrintErrf("Failed to read symlink %v: %v\n", localFilePath, err)
				} else {
					fmt.Printf("Symlink points to %v\n", linkinfo)
					writer.AppendSymlink(archiveFilePath, linkinfo, statn)
				}
			default:
				fmt.Printf("Regular file, size=%v, modtime=%v\n", statn.Size(), statn.ModTime())

				compression := format.COMPRESSION_NONE
				if !*NoCompress && statn.Size() > int64(format.BLOCK_SIZE) {
					if *UseBrotli {
						compression = format.COMPRESSION_BROTLI
					} else {
						compression = format.COMPRESSION_ZSTD
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
	Long: `Create an archive from a specified set of content roots.

	Directories will have their contents added. Two paths that contain the same content will be merged.
	Files are added by their base name (foo/bar.txt -> bar.txt)

	to compress a single folder without making a subdiretory:

	parc create myarchive.pzarc --chdir my-path . 


`,
	Run:     createMain,
	Example: "parc create myarchive.pzarc a/ foo",
}

var BuffSize *uint64
var UseBrotli *bool
var NoCompress *bool
var ExcludePaths *[]string

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().String("comment", "", "Add comment to archive")
	createCmd.Flags().String("prefix", "", "Archive prefix")
	BuffSize = createCmd.Flags().Uint64("buff-size", 5000, "Number of blocks to read into memory at once (default 5000, 2GB)")
	createCmd.Flags().String("chdir", ".", "Search this path to find relative paths")
	NoCompress = createCmd.Flags().Bool("no-compress", false, "Disable compression")
	UseBrotli = createCmd.Flags().Bool("brotli", false, "use Brotli compression vs. ZStandard")
}
