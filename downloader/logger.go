package downloader

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Logger *logrus.Logger

func init() {
	Logger = &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.DebugLevel,
		Formatter: &logrus.TextFormatter{
			ForceColors:               true,
			EnvironmentOverrideColors: true,
			DisableQuote:              true,
			DisableLevelTruncation:    true,
			FullTimestamp:             true,
			TimestampFormat:           "15:04:05",
		},
	}
}
