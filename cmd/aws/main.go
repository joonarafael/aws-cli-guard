// Command aws is the AWS CLI Guard wrapper entrypoint.
package main

import (
	"aws-cli-guard/internal/executor"
	"aws-cli-guard/internal/gate"
	"aws-cli-guard/internal/output"
	"os"
)

func main() {
	env := executor.Environ()

	realAWSPath, err := executor.FindRealAWS()
	if err != nil {
		output.Cancel(os.Stderr, executor.CancellationReason(err))
		os.Exit(output.CancelStatus)
	}

	guard, ok := executor.PolicyGate(realAWSPath, env, os.Stderr)
	if !ok {
		os.Exit(output.CancelStatus)
	}

	if guard && !gate.Decide(os.Args, env, os.Stderr) {
		os.Exit(output.CancelStatus)
	}

	err = executor.ExecRealAWS(realAWSPath, os.Args, env)
	if err != nil {
		output.Cancel(os.Stderr, executor.CancellationReason(err))
		os.Exit(output.CancelStatus)
	}
}
