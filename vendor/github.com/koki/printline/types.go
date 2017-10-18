package printline

type Logger interface {
	// V returns a logger of a particular log level
	V(level int) Logger

	// WithFields returs a logger with fields set
	WithFields(fields Fields) Logger

	// WithFormatter returns a logger with formatter set
	WithFormatter(formatter Formatter) Logger

	// Prints a Info message
	Info(msg interface{})

	// Prints a Error message
	Error(msg interface{})

	// Prints a Fatal message
	Fatal(msg interface{})
}

// Fields type denotes structured fields that are
// embedded into the log message
type Fields map[string]interface{}

type Formatter interface {
	// Format returns a string with the log entry in
	// a specific format
	Format(logType LogType, msg interface{}, fields Fields) []byte
}

type LogType string

const (
	LogTypeInfo  LogType = "INFO"
	LogTypeError LogType = "ERROR"
	LogTypeFatal LogType = "FATAL"

	PRINTLINE_ENABLE_ERR_STACKTRACE    = "PRINTLINE_ENABLE_STACKTRACE"
	PRINTLINE_ENABLE_INFO_STACKTRACE   = "PRINTLINE_ENABLE_INFO_STACKTRACE"
	PRINTLINE_DISABLE_FATAL_STACKTRACE = "PRINTLINE_DISABLE_FATAL_STACKTRACE"
	PRINTLINE_STACKTRACE_LEVEL         = "PRINTLINE_STACKTRACE_LEVEL"

	PRINTLINE_ENABLE_INFO_BREAKPOINT = "PRINTLINE_ENABLE_INFO_BREAKPOINT"
	PRINTLINE_ENABLE_ERR_BREAKPOINT  = "PRINTLINE_ENABLE_ERR_BREAKPOINT"
	PRINTLINE_INFO_BREAKPOINT_LEVEL  = "PRINTLINE_INFO_BREAKPOINT_LEVEL"

	PRINTLINE_LOG_VERBOSITY_LEVEL = "PRINTLINE_LOG_VERBOSITY_LEVEL"
	PRINTLINE_LOG_MIN_VERBOSITY   = "PRINTLINE_LOG_MIN_VERBOSITY"
	PRINTLINE_LOG_MAX_VERBOSITY   = "PRINTLINE_LOG_MAX_VERBOSITY"
)
