package printline

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
)

func format(d defaultLogger, msg interface{}, logType LogType) []byte {
	if d.formatter == nil {
		d.formatter = globalFormatter
	}
	return d.formatter.Format(logType, msg, d.fields)

}

func log(d defaultLogger, msg interface{}, logType LogType) {
	if !shouldLog(d, logType) {
		return
	}
	d.mutex.Lock()
	_, err := os.Stdout.Write(format(d, msg, logType))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log %v\n", err)
	}
	d.mutex.Unlock()
}

func shouldLog(d defaultLogger, logType LogType) bool {
	if logType == LogTypeInfo && (d.level > GlobalLogLevelMax || d.level < GlobalLogLevelMin) {
		return false
	}

	return true
}

func stackTrace() string {
	b := make([]byte, 8192)
	b = b[:runtime.Stack(b, false)]
	return string(b)
}

func getGID() []byte {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	return b
}

func serializeFields(fields Fields) string {
	var fieldString bytes.Buffer
	for k, v := range fields {
		fieldString.WriteString(fmt.Sprintf("%s=%s ", k, v))
	}
	return fieldString.String()
}
