// Command aws-guard is the AWS CLI Guard control-plane entrypoint.
package main

import (
	"aws-cli-guard/internal/cli"
	"os"
)

func main() {
	os.Exit(cli.Run(os.Args, os.Environ(), os.Stdout, os.Stderr))
}
