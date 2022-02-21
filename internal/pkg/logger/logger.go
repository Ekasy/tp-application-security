package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger(log_level string) *logrus.Logger {
	logger := logrus.New()
	lvl, err := logrus.ParseLevel(log_level)
	if err != nil {
		logrus.Fatalf("cannot parse log level: %s", err)
	}
	logger.Writer()
	logger.SetLevel(lvl)
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(
		&logrus.TextFormatter{},
	)

	return logger
}
