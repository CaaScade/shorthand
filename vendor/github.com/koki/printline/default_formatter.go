package printline

import (
	"bytes"
	"fmt"
	"time"
)

const defaultTimestampFormat = time.RFC3339

type defaultFormatter struct{}

func (d defaultFormatter) Format(logType LogType, msg interface{}, fields Fields) []byte {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%-26s", time.Now().Format(defaultTimestampFormat)))
	out.WriteString(fmt.Sprintf("%-12s", fmt.Sprintf("%s[%04s] ", logType, getGID())))
	out.WriteString(fmt.Sprintf("%-70s ", msg))
	out.WriteString(fmt.Sprintf("%s \n", serializeFields(fields)))
	if logType == "FATAL" {
		out.WriteString(fmt.Sprintf("%s \n", stackTrace()))
	}
	return out.Bytes()
}
