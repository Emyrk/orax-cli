package log

import (
	"github.com/sirupsen/logrus"
)

type Log struct {
	*logrus.Entry
}

func New(pkg string) Log {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true,
		FullTimestamp: true}

	return Log{Entry: log.WithField("pkg", pkg)}
}
