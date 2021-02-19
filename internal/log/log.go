package log

import (
	"fmt"
	"os"
)

const (
	ask     = "?>"
	info    = "->"
	fatal   = "!!"
	running = ">>"
)

func Ask(v ...interface{}) {
	Log(ask, v...)
	fmt.Scanln()
}

func Info(v ...interface{}) {
	Log(info, v...)
}

func Fatal(v ...interface{}) {
	Log(fatal, v...)
	os.Exit(1)
}

func Running(v ...interface{}) {
	Log(running, v...)
}

func Askf(f string, v ...interface{}) {
	Log(ask, fmt.Sprintf(f, v...))
	fmt.Scanln()
}

func Infof(f string, v ...interface{}) {
	Log(info, fmt.Sprintf(f, v...))
}

func Fatalf(f string, v ...interface{}) {
	Log(fatal, fmt.Sprintf(f, v...))
	os.Exit(1)
}

func Runningf(f string, v ...interface{}) {
	Log(running, fmt.Sprintf(f, v...))
}

func Log(p string, v ...interface{}) {
	s := fmt.Sprintln(v...)
	fmt.Fprintln(os.Stderr, p, s[:len(s)-1])
}
