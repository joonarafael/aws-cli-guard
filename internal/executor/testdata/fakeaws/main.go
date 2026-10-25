package main

import (
	"encoding/json"
	"os"
	"strconv"
)

func main() {
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
