package main

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
)

const defaultVersionLine = "aws-cli/2.9.0 Python/3.11.0 Linux/test exe/test"

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "--version" {
		runVersionProbe()
		return
	}

	runUserCommand()
}

func runVersionProbe() {
	if countPath := os.Getenv("FAKE_AWS_VERSION_COUNT"); countPath != "" {
		appendVersionCount(countPath)
	}

	line := os.Getenv("FAKE_AWS_VERSION")
	if line == "" {
		line = defaultVersionLine
	} else if line == "omit" {
		line = ""
	}

	if line == "unreadable" {
		line = "not a version"
	}

	if line != "" {
		_, _ = io.WriteString(os.Stdout, line+"\n")
	}

	if stderrLine := os.Getenv("FAKE_AWS_VERSION_STDERR"); stderrLine != "" {
		_, _ = io.WriteString(os.Stderr, stderrLine+"\n")
	}

	os.Exit(0)
}

func appendVersionCount(path string) {
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		os.Exit(96)
	}

	count := 0
	if len(data) > 0 {
		parsed, parseErr := strconv.Atoi(string(data))
		if parseErr == nil {
			count = parsed
		}
	}

	count++

	err = os.WriteFile(path, []byte(strconv.Itoa(count)), 0o600)
	if err != nil {
		os.Exit(95)
	}
}

func runUserCommand() {
	if stdinPath := os.Getenv("FAKE_AWS_STDIN"); stdinPath != "" {
		captureStdin(stdinPath)
	}

	recordPath := os.Getenv("FAKE_AWS_RECORD")
	if recordPath != "" {
		payload, err := json.Marshal(os.Args)
		if err != nil {
			os.Exit(99)
		}

		if err := os.WriteFile(recordPath, payload, 0o600); err != nil {
			os.Exit(98)
		}
	}

	code := 0
	if raw := os.Getenv("FAKE_AWS_EXIT"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			os.Exit(97)
		}

		code = parsed
	}

	os.Exit(code)
}

func captureStdin(path string) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(94)
	}

	err = os.WriteFile(path, data, 0o600)
	if err != nil {
		os.Exit(93)
	}
}
