package client

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	logTimestampFormat = "2006-01-02 15:04:05"
)

// NewFileLogger creates logger with file output.
func NewFileLogger(path string) *logrus.Logger {
	return NewLogger(path, logrus.InfoLevel)
}

// NewLogger creates logger with custom level.
func NewLogger(path string, level logrus.Level) *logrus.Logger {

	logger := logrus.New()

	logger.SetLevel(level)

	logger.SetReportCaller(false)

	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: logTimestampFormat,
		ForceColors:     false,
		DisableColors:   true,
	}

	writers := []io.Writer{
		os.Stdout,
	}

	if path != "" {

		file, err := os.OpenFile(
			path,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY,
			0666,
		)

		if err == nil {

			writers = append(
				writers,
				file,
			)

		} else {

			logger.WithError(err).
				Warn("failed to open log file")
		}
	}

	logger.SetOutput(
		io.MultiWriter(
			writers...,
		),
	)

	return logger
}
