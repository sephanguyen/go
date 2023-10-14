package log

import (
	"log"
	"os"
	"runtime"
)

var (
	colorReset  = "\033[0m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"

	infoLogger *log.Logger
	warnLogger *log.Logger
)

func init() {
	if runtime.GOOS == "windows" {
		colorReset = ""
		colorYellow = ""
		colorBlue = ""
	}

	infoLogger = log.New(os.Stdout, colorBlue, log.LstdFlags)
	warnLogger = log.New(os.Stderr, colorYellow, log.LstdFlags)
}

// Info prints the input using log.Logger.Printf to stdout using blue color scheme.
func Info(format string, v ...any) {
	infoLogger.Printf(format+colorReset, v...)
}

// Warn prints the input using log.Logger.Printf to stderr using yellow color scheme.
func Warn(format string, v ...any) {
	warnLogger.Printf(format+colorReset, v...)
}
