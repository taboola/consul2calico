package log

import (
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// Default log format will output [INFO]: 2006-01-02T15:04:05Z07:00 - Log message
	defaultLogFormat       = "[%lvl%]: %time% - %msg%"
	defaultTimestampFormat = time.RFC3339
)

var Category = ""
var Thread = ""
var Context = ""

// Formatter implements logrus.Formatter interface.
type Formatter struct {
	// Timestamp format
	TimestampFormat string
	// Available standard keys: time, msg, lvl
	// Also can include custom fields but limited to strings.
	// All of fields need to be wrapped inside %% i.e %time% %msg%
	LogFormat string
}

// Format building log message.
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	output := f.LogFormat
	if output == "" {
		output = defaultLogFormat
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}

	output = strings.Replace(output, "%time%", entry.Time.Format(timestampFormat), 1)

	output = strings.Replace(output, "%msg%", entry.Message, 1)

	level := strings.ToUpper(entry.Level.String())
	output = strings.Replace(output, "%lvl%", level, 1)

	mapping := map[string]string{}
	for k, val := range entry.Data {
		switch v := val.(type) {
		case string:
			mapping[k] = v
			output = strings.Replace(output, "%"+k+"%", v, 1)
		case int:
			s := strconv.Itoa(v)
			output = strings.Replace(output, "%"+k+"%", s, 1)
		case bool:
			s := strconv.FormatBool(v)
			output = strings.Replace(output, "%"+k+"%", s, 1)
		}
	}

	if mapping["thread"] == "" {
		output = strings.Replace(output, "%thread%", Thread, 1)
	}

	if mapping["category"] == "" {
		output = strings.Replace(output, "%category%", Category, 1)
	}

	if mapping["context"] == "" {
		output = strings.Replace(output, "%context%", Context, 1)
	}

	//removing prefix space, which causes wrong log output
	output = strings.TrimLeft(output, " ")
	return []byte(output), nil
}

func SeValues(thread string, category string, context string) {
	Thread = thread
	Category = category
	Context = context
}

func LoggerFunction(args ...string) *logrus.Entry {
	if len(args) > 2 {
		return logrus.WithField("thread", args[0]).WithField("context", args[1]).WithField("category", args[2])
	}

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("Could not get context info for logger!")
	}

	filename := file[strings.LastIndex(file, "/")+1:] + ":" + strconv.Itoa(line)
	funcname := runtime.FuncForPC(pc).Name()
	fn := funcname[strings.LastIndex(funcname, ".")+1:]
	if len(args) == 1 {
		return logrus.WithField("thread", args[0]).WithField("context", filename).WithField("category", fn)
	}
	return logrus.WithField("context", filename).WithField("category", fn)

}
