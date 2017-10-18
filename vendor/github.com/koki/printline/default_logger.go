package printline

import (
	"os"
	"sync"
)

// defaultLogger is the default implementation of the
// Logger interface
type defaultLogger struct {
	// the current log level
	level int

	// the fields in the structured log
	fields Fields

	// mutex to ensure thread safety
	mutex sync.Mutex

	// formatter formats the message
	formatter Formatter
}

func (d defaultLogger) V(level int) Logger {
	return defaultLogger{
		level:  level,
		fields: d.fields,
	}
}

func (d defaultLogger) WithFormatter(formatter Formatter) Logger {
	return defaultLogger{
		level:     d.level,
		fields:    d.fields,
		formatter: formatter,
	}
}

func (d defaultLogger) WithFields(fields Fields) Logger {
	return defaultLogger{
		level:  d.level,
		fields: fields,
	}
}

func (d defaultLogger) Info(msg interface{}) {
	log(d, msg, LogTypeInfo)
}

func (d defaultLogger) Error(msg interface{}) {
	log(d, msg, LogTypeError)
}

func (d defaultLogger) Fatal(msg interface{}) {
	log(d, msg, LogTypeFatal)
	os.Exit(1)
}
