// Command aws is the AWS CLI Guard wrapper entrypoint.
package main

import (
	"aws-cli-guard/internal/executor"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/subject"
	"os"
)

func main() {
	realAWSPath, err := executor.FindRealAWS()
	if err != nil {
		output.Cancel(os.Stderr, executor.CancellationReason(err))
		os.Exit(output.CancelStatus)
	}

	env := executor.Environ()
	subject.WarnDefaultProfile(os.Stderr, subject.Resolve(os.Args, env))

	err = executor.ExecRealAWS(realAWSPath, os.Args, env)
	if err != nil {
		output.Cancel(os.Stderr, executor.CancellationReason(err))
		os.Exit(output.CancelStatus)
	}
}
