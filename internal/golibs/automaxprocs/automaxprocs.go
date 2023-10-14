// Package automaxprocs does exactly the following
//
//	import (
//		_ "go.uber.org/automaxprocs"
//	)
//
// but since we cannot control the log output from that upstream package,
// this package customizes the log output to JSON.
//
// Example usage: simply import the package for its side effect
//
//	import (
//		_ "github.com/manabie-com/backend/internal/golibs/automaxprocs"
//	)
package automaxprocs

import (
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	// l prints to os.Stdout in JSON format
	// this helps avoid confusion because the default printer prints to os.Stderr
	l := log.New(os.Stdout, "", 0)
	printer := func(format string, v ...interface{}) {
		ts := time.Now().Format(time.RFC3339)
		msg := fmt.Sprintf(format, v...)
		l.Print(`{"severity":"info","time":"` + ts + `","msg":"` + msg + `"}`)
	}

	undo, err := maxprocs.Set(maxprocs.Logger(printer))
	if err != nil {
		log.Printf("automaxprocs failed: %s", err)
		undo()
	}
}
