// Command aws is the AWS CLI Guard wrapper entrypoint.
package main

import (
	"aws-cli-guard/internal/executor"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/policy"
	"aws-cli-guard/internal/subject"
	"os"
)

func main() {
	env := executor.Environ()

	err := policy.Present(env)
	if err != nil {
		output.Cancel(os.Stderr, policy.PresentReason(err))
		os.Exit(output.CancelStatus)
	}

	realAWSPath, err := executor.FindRealAWS()
	if err != nil {
		output.Cancel(os.Stderr, executor.CancellationReason(err))
		os.Exit(output.CancelStatus)
	}

	subject.WarnDefaultProfile(os.Stderr, subject.Resolve(os.Args, env))

	err = executor.ExecRealAWS(realAWSPath, os.Args, env)
	if err != nil {
		output.Cancel(os.Stderr, executor.CancellationReason(err))
		os.Exit(output.CancelStatus)
	}
}
