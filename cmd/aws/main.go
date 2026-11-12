// Command aws is the AWS CLI Guard wrapper entrypoint.
package main

import (
	"aws-cli-guard/internal/executor"
	"aws-cli-guard/internal/gate"
	"aws-cli-guard/internal/output"
	"aws-cli-guard/internal/paths"
	"aws-cli-guard/internal/settings"
	"aws-cli-guard/internal/update"
	"aws-cli-guard/internal/verbose"
	"os"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == update.CheckArgv {
		update.RunUpdateCheckFromEnv(os.Environ())
		os.Exit(0)
	}

	env := executor.Environ()

	envLevel, err := verbose.EnvLevel(env)
	if err != nil {
		output.Cancel(os.Stderr, verbose.ReasonInvalid)
		os.Exit(output.CancelStatus)
	}

	realAWSPath, err := executor.FindRealAWS()
	if err != nil {
		if envLevel == settings.LevelVerbose {
			verbose.Write(
				os.Stderr,
				envLevel,
				verbose.Event{ //nolint:exhaustruct_v5 // discovery refuse block.
					Home:    homeLabelForLog(env),
					Outcome: "refuse",
				},
			)
		}

		output.Cancel(os.Stderr, executor.CancellationReason(err))
		os.Exit(output.CancelStatus)
	}

	guard, ok := executor.PolicyGate(realAWSPath, env, os.Stderr)
	if !ok {
		os.Exit(output.CancelStatus)
	}

	if guard && !gate.Decide(os.Args, env, realAWSPath, os.Stderr) {
		os.Exit(output.CancelStatus)
	}

	err = executor.ExecRealAWS(realAWSPath, os.Args, env)
	if err != nil {
		output.Cancel(os.Stderr, executor.CancellationReason(err))
		os.Exit(output.CancelStatus)
	}
}

func homeLabelForLog(env []string) string {
	label, ok := paths.HomeOrigin(env)
	if !ok {
		return ""
	}

	return label
}
