package logger

import (
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	once     sync.Once
	logger   zerolog.Logger
	initDone bool
)

func Init(level zerolog.Level) {
	once.Do(func() {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC822,
			FormatCaller: func(i interface{}) string {
				if s, ok := i.(string); ok {
					if index := strings.LastIndex(s, "/"); index != -1 {

						return s[index+1:]
					}
					return s
				}
				return ""
			},
		}
		logger = zerolog.New(output).
			Level(level).
			With().
			Timestamp().
			Logger()

		initDone = true
	})
}

func Get() *zerolog.Logger {
	if !initDone {
		Init(zerolog.InfoLevel)
	}
	return &logger
}

func WithCaller() *zerolog.Logger {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return Get()
	}

	var funcName string
	if fun := runtime.FuncForPC(pc); fun != nil {
		funcName = filepath.Base(fun.Name())
	}
	relFile := file

	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, file); err == nil {
			relFile = rel
		}
	}

	lg := Get().With().
		Str("func", funcName).
		Str("file", relFile).
		Int("line", line).
		Logger()
	return &lg
}

func Debug() *zerolog.Event {
	return WithCaller().Debug()
}

func Info() *zerolog.Event {
	return WithCaller().Info()
}

func Warn() *zerolog.Event {
	return WithCaller().Warn()
}

func Error() *zerolog.Event {
	return WithCaller().Error()
}

func Fatal() *zerolog.Event {
	return WithCaller().Fatal()
}

func Panic() *zerolog.Event {
	return WithCaller().Panic()
}
