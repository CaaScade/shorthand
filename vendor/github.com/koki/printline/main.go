package printline

/* This project is meant to be a logging library that also
 * doubles down as a debugging tool when needed.
 *
 * Three goals of this project is to provide the following
 * tools to a person debugging a go project
 *
 * 1. StackTrace logging
 * 2. Runtime log level windows
 * 3. StackTraces of nested errors
 */

import (
	"fmt"
)

func test() {
	V(1).Info("This is a Lvl1 Info message")
	V(2).WithFields(
		Fields{
			"field1": "value1",
			"field2": "value2",
		}).Info("This is a structured Lvl 2 Info message with additional information")
	V(3).Info("This is a Lvl3 Info message")
	V(4).Info("This is a Lvl4 Info message")
	V(5).Info("This is a Lvl5 Info message")
	Error("This is a Error message")
	WithFields(
		Fields{
			"field1": "value1",
			"field2": "value2",
		}).Error("This is a structured Error message with additional information")
	Error(stackedErrors())
	stackedLogs() //prints stack trace by default on Fatal log messages
}

func stackedErrors() string {
	return depth1().StackError(fmt.Errorf("caller error")).Error()
}

func depth1() StackedError {
	return depth2().StackError(fmt.Errorf("depth1 error"))
}

func depth2() StackedError {
	return NewError().StackError(fmt.Errorf("depth2 error"))
}

func stackedLogs() {
	stackLvl1()
}

func stackLvl1() {
	stackLvl2()
}

func stackLvl2() {
	WithFields(
		Fields{
			"field1": "value1",
			"field2": "value2",
		}).Fatal("Quitting.")
}
