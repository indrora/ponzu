package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	rootCmd.PersistentPreRunE = setupLogging
	rootCmd.PersistentPostRun = closeLogging
	loggerUsesJSON = rootCmd.PersistentFlags().BoolP("log-json", "J", false, "Use JSON logging")
	loggerLevel = rootCmd.PersistentFlags().StringP("log-level", "L", "info", "Log level (debug|info|warn|error|dpanic|panic|fatal)")
}

var loggerLevel *string
var loggerUsesJSON *bool
var GlobalLogger *zap.Logger

type CobraSyncWriter struct {
	command *cobra.Command
}

func NewCobraWriter(command *cobra.Command) *CobraSyncWriter {
	return &CobraSyncWriter{command: command}
}

func (w *CobraSyncWriter) Write(p []byte) (int, error) {
	return w.command.OutOrStderr().Write(p)
}
func (w *CobraSyncWriter) Sync() error {
	/* noop*/
	return nil
}

func setupLogging(cmd *cobra.Command, args []string) error {

	atom := zap.NewAtomicLevel()

	mLevel, err := zapcore.ParseLevel(*loggerLevel)
	if err != nil {
		return err
	}
	atom.SetLevel(mLevel)

	var encoder zapcore.Encoder

	if *loggerUsesJSON {

		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoderCfg := zap.NewDevelopmentEncoderConfig()
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	GlobalLogger = zap.New(zapcore.NewCore(
		encoder,
		zapcore.Lock(NewCobraWriter(cmd)),
		atom,
	))
	GlobalLogger.Debug("Logging begins", zap.Strings("args", args))
	return nil
}

func closeLogging(cmd *cobra.Command, args []string) {
	GlobalLogger.Sync()
}
