package common

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	instance *logrus.Logger
	once     sync.Once
)

func GetLog() *logrus.Logger {

	once.Do(func() {
		instance = logrus.New()
		instance.Formatter = &logrus.TextFormatter{ForceColors: true,
			FullTimestamp: true}
	})

	return instance
}
