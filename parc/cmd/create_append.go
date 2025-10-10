package cmd

import (
	"os"

	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/writer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func appendToArchive(localFilePath string, localLog *zap.Logger, writer *writer.ArchiveWriter, archiveFilePath string) error {
	cFileStat, err := os.Lstat(localFilePath)
	if err != nil {
		localLog.Fatal("could not stat file path", zap.Error(err))
	}

	localLog.Debug("got file stat", zap.Any("stat", cFileStat))

	mask := os.ModeDir | os.ModeSymlink

	switch mode := cFileStat.Mode(); mode & mask {
	case os.ModeDir:
		localLog.Info("Append Directory")
		writer.AppendDirectory(archiveFilePath, cFileStat)
	case os.ModeSymlink:
		localLog.Info("Append symlink")
		linkinfo, err := os.Readlink(localFilePath)
		if err != nil {
			return errors.Wrap(err, "failed to read symlink")
		} else {
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
			return errors.Wrap(err, "Failed to append file to archive")
		}
	}
	return nil
}
