package log

import (
	"fmt"
	"os"
)

// TODO refactor
// TODO add *f functions

func Info(v ...interface{}) {
	Custom("info", v...)
}

func Fatal(v ...interface{}) {
	Custom("fatal", v...)
	os.Exit(1)
}

func Warning(v ...interface{}) {
	Custom("warning", v...)
}

func Success(v ...interface{}) {
	Custom("success", v...)
}

func Custom(p string, v ...interface{}) {
	m := fmt.Sprintln(v...)
	fmt.Fprintln(os.Stderr, "["+p+"]", m[:len(m)-1])
}
