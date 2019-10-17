package common

import (
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

var (
	instance  *logrus.Logger
	once      sync.Once
	successCr = color.New(color.FgGreen)
	errorC    = color.New(color.FgRed)
)

func GetLog() *logrus.Logger {

	once.Do(func() {
		instance = logrus.New()
		formatter := &logrus.TextFormatter{FullTimestamp: true}
		instance.Formatter = formatter
	})

	return instance
}

func SetLogColor(c string) {

	switch c {
	case "auto":
		instance.Formatter = &logrus.TextFormatter{FullTimestamp: true}
		break
	case "on":
		instance.Formatter = &logrus.TextFormatter{FullTimestamp: true, ForceColors: true}
		color.NoColor = false
		break
	case "off":
		instance.Formatter = &logrus.TextFormatter{FullTimestamp: true, DisableColors: true}
		color.NoColor = true
		break
	default:
		PrintError("Invalid value for --color: [%s] \n", c)
		os.Exit(1)
	}
}

func PrintSuccess(format string, a ...interface{}) {
	successCr.Fprintf(os.Stdout, format, a...)
}

func PrintError(format string, a ...interface{}) {
	errorC.Fprintf(os.Stderr, format, a...)
}
