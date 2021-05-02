package log

import (
	"fmt"
	"os"
)

// TODO colors ?
const (
	info    = "->"
	fatal   = "!!"
	prompt  = "?>"
	running = ">>"
)

var NoPrompt bool

func Info(v interface{}) {
	log(info, v)
}

func Fatal(v interface{}) {
	log(fatal, v)
	os.Exit(1)
}

func Prompt(v interface{}) {
	if !NoPrompt {
		log(prompt, v)
		fmt.Scanln()
	}
}

func Running(v interface{}) {
	log(running, v)
}

func Infof(f string, v ...interface{}) {
	log(info, fmt.Sprintf(f, v...))
}

func Fatalf(f string, v ...interface{}) {
	log(fatal, fmt.Sprintf(f, v...))
	os.Exit(1)
}

func Promptf(f string, v ...interface{}) {
	if !NoPrompt {
		log(prompt, fmt.Sprintf(f, v...))
		fmt.Scanln()
	}
}

func Runningf(f string, v ...interface{}) {
	log(running, fmt.Sprintf(f, v...))
}

func log(p string, v interface{}) {
	fmt.Fprintln(os.Stderr, p, v)
}
