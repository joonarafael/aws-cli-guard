package main

import (
	"fmt"
	"os"

	"aws-cli-guard/internal/executor"
)

func main() {
	real, err := executor.FindRealAWS()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := executor.ExecRealAWS(real, os.Args, executor.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
