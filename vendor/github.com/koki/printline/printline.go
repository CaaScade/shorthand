package printline

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var GlobalLogLevelMax int = 1<<32 - 1 // MaxUint32
var GlobalLogLevelMin int = 1

func init() {
	if min := os.Getenv(PRINTLINE_LOG_MIN_VERBOSITY); min != "" {
		minInt, err := strconv.ParseInt(min, 10, 32)
		if err != nil {
			panic(fmt.Sprintf("Invalid format: %s=%s", PRINTLINE_LOG_MIN_VERBOSITY, min))
		}
		GlobalLogLevelMin = int(minInt)
	}

	if max := os.Getenv(PRINTLINE_LOG_MAX_VERBOSITY); max != "" {
		maxInt, err := strconv.ParseInt(max, 10, 32)
		if err != nil {
			panic(fmt.Sprintf("Invalid format: %s=%s", PRINTLINE_LOG_MAX_VERBOSITY, max))
		}
		GlobalLogLevelMax = int(maxInt)

	}

	if logRange := os.Getenv(PRINTLINE_LOG_VERBOSITY_LEVEL); logRange != "" {
		vals := strings.Split(logRange, ":")
		if len(vals) == 1 {
			panic(fmt.Sprintf("Invalid format: %s=%s. Expected %s=Min:Max", PRINTLINE_LOG_VERBOSITY_LEVEL, vals[0], PRINTLINE_LOG_VERBOSITY_LEVEL))
		}

		minInt, err := strconv.ParseInt(vals[0], 10, 32)
		if err != nil {
			panic(fmt.Sprintf("Invalid format: %s=%s. Expected Min value to be int", PRINTLINE_LOG_VERBOSITY_LEVEL, vals[0]))
		}
		GlobalLogLevelMin = int(minInt)

		maxInt, err := strconv.ParseInt(vals[1], 10, 32)
		if err != nil {
			panic(fmt.Sprintf("Invalid format: %s=%s. Expected Max value to be int", PRINTLINE_LOG_VERBOSITY_LEVEL, vals[1]))
		}
		GlobalLogLevelMax = int(maxInt)
	}
}

var globalLogger = defaultLogger{
	level: 1,
}

var globalFormatter = defaultFormatter{}

func V(level int) Logger {
	return globalLogger.V(level)
}

func WithFields(fields Fields) Logger {
	return globalLogger.WithFields(fields)
}

func WithFormatter(formatter Formatter) Logger {
	return globalLogger.WithFormatter(formatter)
}

func Info(msg interface{}) {
	globalLogger.Info(msg)
}

func Error(msg interface{}) {
	globalLogger.Error(msg)
}

func Fatal(msg interface{}) {
	globalLogger.Fatal(msg)
}
