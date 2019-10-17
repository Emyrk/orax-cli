package common

import (
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

var (
	instance *logrus.Logger
	once     sync.Once
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
		fmt.Fprintf(os.Stderr, "Invalid value for --color: [%s] \n", c)
		os.Exit(1)
	}
}
