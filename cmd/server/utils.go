package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/zanz1n/blog/internal/utils/errutils"
)

func fatal(err any) {
	exitCode := 1
	if err, ok := err.(error); ok {
		exitCode = errutils.Os(err).OsStatus()
	}

	slog.Error(fmt.Sprint("FATAL: ", err))
	os.Exit(exitCode)
}
