package log

import (
	"fmt"
	"os"
)

const (
	ask     = "?>"
	info    = "->"
	fatal   = "!>"
	running = ">>"
	success = "+>"
)

func Ask(v ...interface{}) {
	Custom(ask, v...)
	fmt.Scanln()
}

func Info(v ...interface{}) {
	Custom(info, v...)
}

func Fatal(v ...interface{}) {
	Custom(fatal, v...)
	os.Exit(1)
}

func Running(v ...interface{}) {
	Custom(running, v...)
}

func Success(v ...interface{}) {
	Custom(success, v...)
}

func Askf(f string, v ...interface{}) {
	Custom(ask, fmt.Sprintf(f, v...))
	fmt.Scanln()
}

func Infof(f string, v ...interface{}) {
	Custom(info, fmt.Sprintf(f, v...))
}

func Fatalf(f string, v ...interface{}) {
	Custom(fatal, fmt.Sprintf(f, v...))
	os.Exit(1)
}

func Runningf(f string, v ...interface{}) {
	Custom(running, fmt.Sprintf(f, v...))
}

func Successf(f string, v ...interface{}) {
	Custom(success, fmt.Sprintf(f, v...))
}

func Custom(p string, v ...interface{}) {
	s := fmt.Sprintln(v...)
	fmt.Fprintln(os.Stderr, p, s[:len(s)-1])
}

// func custom(p, s string) {
// 	fmt.Fprintln(os.Stderr, p, s)
// }
