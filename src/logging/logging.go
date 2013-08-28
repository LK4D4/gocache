// logging package

package logging

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var Config = struct {
	verbosity int
}{
	verbosity: 4,
}

func SetVerbosity(v int) {
	Config.verbosity = v
}

const (
	CRITICAL int = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

var LevelsParams = []struct {
	prefix string
	out    interface {
		io.Writer
	}
}{
	{"[CRITICAL]", os.Stdout},
	{"[ERROR]", os.Stdout},
	{"[WARNING]", os.Stdout},
	{"[NOTICE]", os.Stdout},
	{"[INFO]", os.Stdout},
	{"[DEBUG]", os.Stdout},
}

type Level struct {
	out    io.Writer
	prefix string
}

type Logger struct {
	sync.Mutex
	levels [](*Level)
}

func InitLogger() *Logger {
	logger := new(Logger)
	logger.levels = make([](*Level), len(LevelsParams))
	for levelIndex, params := range LevelsParams {
		logger.levels[levelIndex] = &Level{params.out, params.prefix}
	}

	return logger
}

var DefaultLogger *Logger = InitLogger()

func (lvl *Level) log(format string, v ...interface{}) {
	var buf bytes.Buffer
	buf.WriteString(time.Now().Format(time.RFC3339))
	buf.WriteString(" ")

	buf.WriteString(lvl.prefix)
	buf.WriteString(" ")
	buf.WriteString(fmt.Sprintf(format, v...))
	buf.WriteString("\n")
	_, err := buf.WriteTo(lvl.out)
	if err != nil {
		fmt.Printf("Writing bubffer error: %v", err)
	}
}

func (l *Logger) log(level int, format string, v ...interface{}) {
	if level > Config.verbosity {
		return
	}
	l.Lock()
	defer l.Unlock()
	l.levels[level].log(format, v...)
}

func Crit(format string, v ...interface{}) {
	DefaultLogger.log(CRITICAL, format, v...)
}

func Err(format string, v ...interface{}) {
	DefaultLogger.log(ERROR, format, v...)
}

func Warn(format string, v ...interface{}) {
	DefaultLogger.log(WARNING, format, v...)
}

func Notice(format string, v ...interface{}) {
	DefaultLogger.log(NOTICE, format, v...)
}

func Info(format string, v ...interface{}) {
	DefaultLogger.log(INFO, format, v...)
}

func Debug(format string, v ...interface{}) {
	DefaultLogger.log(DEBUG, format, v...)
}
