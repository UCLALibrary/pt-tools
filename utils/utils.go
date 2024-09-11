package utils

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger creates logger with output of info and debug to file and error to stdout
func Logger(logFile string) *zap.Logger {
	pe := zap.NewDevelopmentEncoderConfig()

	fileEncoder := zapcore.NewJSONEncoder(pe)

	pe.EncodeTime = zapcore.ISO8601TimeEncoder // The encoder can be customized for each output

	// Console encoder (for stdout)
	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	// Create file core
	file, err := os.Create(logFile)
	if err != nil {
		panic(err)
	}

	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(file), zap.DebugLevel)

	// Console core for errors
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.ErrorLevel)

	// Combine the cores
	core := zapcore.NewTee(fileCore, consoleCore)
	// Create a logger with two cores
	logger := zap.New(core, zap.AddCaller())

	return logger
}

// ApplyExitOnHelp exits out of program if --help is flag
func ApplyExitOnHelp(c *cobra.Command, exitCode int) {
	helpFunc := c.HelpFunc()
	c.SetHelpFunc(func(c *cobra.Command, s []string) {
		helpFunc(c, s)
		os.Exit(exitCode)
	})
}
