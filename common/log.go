package common

import (
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

var (
	instance *logrus.Logger
	once     sync.Once
	successC = color.New(color.FgGreen)
	errorC   = color.New(color.FgRed)
)

func GetLog() *logrus.Logger {

	once.Do(func() {
		instance = logrus.New()
	})

	return instance
}

func SetLogConfig(c string) {
	switch c {
	case "auto":
		instance.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
		break
	case "on":
		instance.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, ForceColors: true})
		color.NoColor = false
		break
	case "off":
		instance.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, DisableColors: true})
		color.NoColor = true
		break
	default:
		PrintError("Invalid value for --color: [%s] \n", c)
		os.Exit(1)
	}

	instance.AddHook(NewStdDemuxerHook(instance))
}

func PrintSuccess(format string, a ...interface{}) {
	successC.Fprintf(os.Stdout, format, a...)
}

func PrintError(format string, a ...interface{}) {
	errorC.Fprintf(os.Stderr, format, a...)
}
