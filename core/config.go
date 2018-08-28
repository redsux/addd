package addd

import (
	"os"
	"strings"
	"github.com/apsdehal/go-logger"
)

var (
	Log *logger.Logger
)

func init() {
	var err error
	Log, err = logger.New("aaaa", 1, os.Stdout)
	if err != nil {
		panic(err)
	}
	Log.SetFormat("%{time:2006-01-02 15:04:05} %{level} ▶ %{message}")
}

func SetLoglevel(level string) {
	if strings.EqualFold(level, "DEBUG") {
		Log.SetLogLevel(logger.DebugLevel)
		Log.SetFormat("%{time:2006-01-02 15:04:05} %{file}:%{line} ▶ %{level} : %{message}")
	} else if strings.EqualFold(level, "CRITICAL") {
		Log.SetLogLevel(logger.ErrorLevel)
	} else if strings.EqualFold(level, "WARNING") {
		Log.SetLogLevel(logger.WarningLevel)
	} else {
		Log.SetLogLevel(logger.NoticeLevel)
	}
}