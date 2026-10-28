package awsconfig

import (
	"aws-cli-guard/internal/environ"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	configFileEnv = "AWS_CONFIG_FILE"
	//nolint:gosec // G101: AWS environment variable name, not a credential value.
	credentialsFileEnv = "AWS_SHARED_CREDENTIALS_FILE"
)

// ConfigPath resolves the absolute AWS config file path from env.
func ConfigPath(env []string) (string, error) {
	raw, err := configPathRaw(env)
	if err != nil {
		return "", err
	}

	return absPath(raw, env)
}

// CredentialsPath resolves the absolute shared credentials file path from env.
func CredentialsPath(env []string) (string, error) {
	raw, err := credentialsPathRaw(env)
	if err != nil {
		return "", err
	}

	return absPath(raw, env)
}

func configPathRaw(env []string) (string, error) {
	if override, ok := environ.Lookup(configFileEnv, env); ok && override != "" {
		return override, nil
	}

	home, err := requireHome(env)
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".aws", "config"), nil
}

func credentialsPathRaw(env []string) (string, error) {
	if override, ok := environ.Lookup(credentialsFileEnv, env); ok && override != "" {
		return override, nil
	}

	home, err := requireHome(env)
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".aws", "credentials"), nil
}

func requireHome(env []string) (string, error) {
	home, ok := environ.Lookup("HOME", env)
	if !ok || home == "" {
		return "", ErrHomeUnset
	}

	return home, nil
}

func absPath(path string, env []string) (string, error) {
	expanded, err := expandPath(path, env)
	if err != nil {
		return "", err
	}

	clean := filepath.Clean(expanded)

	resolved, err := filepath.Abs(clean)
	if err != nil {
		return "", fmt.Errorf("aws-cli-guard: awsconfig: resolve path %q: %w", path, err)
	}

	return resolved, nil
}

func expandPath(path string, env []string) (string, error) {
	path, err := expandTilde(path, env)
	if err != nil {
		return "", err
	}

	return expandVars(path, env), nil
}

func expandTilde(path string, env []string) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}

	if path == "~" {
		home, err := requireHome(env)
		if err != nil {
			return "", err
		}

		return home, nil
	}

	if len(path) >= 2 && path[1] == filepath.Separator {
		home, err := requireHome(env)
		if err != nil {
			return "", err
		}

		return home + path[1:], nil
	}

	return "", ErrTildeUser
}

func expandVars(path string, env []string) string {
	var out strings.Builder

	for index := 0; index < len(path); index++ {
		if path[index] != '$' {
			out.WriteByte(path[index])

			continue
		}

		if index+1 < len(path) && path[index+1] == '$' {
			out.WriteByte('$')

			index++

			continue
		}

		name, consumed := readVarName(path[index:])
		if consumed == 0 {
			out.WriteByte('$')

			continue
		}

		if val, found := environ.Lookup(name, env); found {
			out.WriteString(val)
		} else {
			out.WriteString(path[index : index+consumed])
		}

		index += consumed - 1
	}

	return out.String()
}

const minDollarVarLen = 2

func readVarName(rest string) (string, int) {
	if len(rest) < minDollarVarLen || rest[0] != '$' {
		return "", 0
	}

	if rest[1] == '{' {
		end := strings.IndexByte(rest, '}')
		if end < minDollarVarLen {
			return "", 0
		}

		return rest[2:end], end + 1
	}

	start := 1
	for start < len(rest) {
		char := rest[start]
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || char == '_' {
			start++

			continue
		}

		if start > 1 && char >= '0' && char <= '9' {
			start++

			continue
		}

		break
	}

	if start == 1 {
		return "", 0
	}

	return rest[1:start], start
}
