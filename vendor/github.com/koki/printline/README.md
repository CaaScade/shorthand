# Printline

Named after the beloved task of print line debugging in go, this debugger also doubles down as a logger as decided at runtime. 

In order to use this library, import this package

`import "github.com/koki/printline"`

then you can use it to log lines like so,

```
printline.Info("log message")
printline.Error("error message")
printline.Fatal("fatal message")
```

As implicitly shown above, there are 3 log levels

```
Info
Error
Fatal
```

Info can be further divided into sub-levels. This pattern is based on setting verbosity levels in popular projects like Kubernetes.

Sub-levels only apply to Info level logs. 

For eg. In order to log at lvl 2 and lvl 3

```
printline.V(2).Info("lvl2 log message")
printline.V(3).Info("lvl3 log message")
```

During runtime, the log level window can be chosen

For eg. In order to only view lvl 3 logs

```
PRINTLINE_LOG_VERBOSITY_LEVEL=3:3 ./your_program_with_printline
```

In order to view a window of levels (levels 2 and 3)

```
PRINTLINE_LOG_VERBOSITY_LEVEL=2:3 ./your_program_with_printline
```

The value of `PRINTLINE_LOG_VERBOSITY_LEVEL` has to be two unsigned integers separated by a `:`. This is a inclusive representation of the window. i.e. Both values on either side of `:` are considered a part of the window.


This library also provides primitives for error merging. There are times when multiple levels of errors from multiple stack frames need to be given back to the user in a readable format. 

In order to use this feature, replace `error` objects in your project with `printline.StackError`

For more information about usage, refer to test.go

### Other features
-------------------


Prints Stack traces on Fatal logs (log level will be configurable in the future)

### Contributing
------------------

We welcome contributions, issues, bug reports and documentations. Please make a PR! 
