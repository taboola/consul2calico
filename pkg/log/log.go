package log

import (
	"github.com/sirupsen/logrus"
	"github.com/taboola/consul2calico/pkg/utils"
	"time"
)

var logLevel = utils.GetEnv("LOG_LEVEL", "DEBUG")

func init() {
	//Set format for logs
	formatter := &Formatter{
		TimestampFormat: time.RFC3339,
		LogFormat:       " %time% %lvl% [%thread%] %category% [%context%] - %msg% \n",
	}

	logrus.SetFormatter(formatter)

	//Set loglevel based on ENV
	logLvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.Errorf("LOG LEVEL %v is not a valid log level , ERROR %v \n", logLvl.String(), err)
	} else {
		logrus.SetLevel(logLvl)
	}

}
