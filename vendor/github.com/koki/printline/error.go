package printline

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

type StackedError struct {
	err  []error
	meta []string
}

func NewError() StackedError {
	return StackedError{
		err:  []error{},
		meta: []string{},
	}
}

func (e StackedError) Error() string {
	var out bytes.Buffer
	out.WriteString("\n")
	for i := range e.err {
		out.WriteString(fmt.Sprintf("%s:\n", e.meta[len(e.err)-1-i]))
		out.WriteString(fmt.Sprintf("\t%s\n", e.err[len(e.err)-1-i].Error()))
	}
	return out.String()
}

func (e StackedError) StackError(err error) StackedError {
	e.err = append(e.err, err)
	e.meta = append(e.meta, stackInfo())
	return e
}

func stackInfo() string {
	if _, file, line, ok := runtime.Caller(3); ok {
		return strings.Join([]string{file, strconv.Itoa(line)}, ":")
	}
	return ""
}
