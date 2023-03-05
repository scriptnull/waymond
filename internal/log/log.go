// log package is trying to make all logs from waymond to be logged via a single interface
package log

import (
	"fmt"
	"strings"
)

type Logger struct {
	prefix string
}

func New(prefix string) Logger {
	return Logger{
		prefix,
	}
}

func (l *Logger) logf(format string, args ...any) {
	var logdata = []string{
		l.prefix, "::", format,
	}
	fmt.Printf(strings.Join(logdata, " "), args...)
}

func (l *Logger) log(args ...any) {
	logargs := []any{
		l.prefix,
		"::",
	}
	logargs = append(logargs, args...)
	fmt.Println(logargs...)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.logf(format, args...)
}

func (l *Logger) Verbosef(format string, args ...any) {
	l.logf(format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.logf(format, args...)
}

func (l *Logger) Debug(args ...any) {
	l.log(args...)
}

func (l *Logger) Verbose(args ...any) {
	l.log(args...)
}

func (l *Logger) Error(args ...any) {
	l.log(args...)
}
