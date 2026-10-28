package awsconfig

import (
	"errors"
	"io"
	"os"
)

const maxFileSize = 1 << 20

// Load reads and merges the AWS config and shared credentials files from env.
func Load(env []string) (Config, error) {
	configPath, err := ConfigPath(env)
	if err != nil {
		return Config{}, err
	}

	credPath, err := CredentialsPath(env)
	if err != nil {
		return Config{}, err
	}

	cfg := emptyConfig(configPath, credPath)

	configData, err := readFileOptional(configPath)
	if err != nil {
		return Config{}, err
	}

	credData, err := readFileOptional(credPath)
	if err != nil {
		return Config{}, err
	}

	profiles, ssoSessions, services, err := buildFromConfig(configPath, configData)
	if err != nil {
		return Config{}, err
	}

	credProfiles, err := buildFromCredentials(credPath, credData)
	if err != nil {
		return Config{}, err
	}

	cfg.Profiles = mergeProfiles(profiles, credProfiles)
	cfg.SSOSessions = ssoSessions
	cfg.Services = services

	return cfg, nil
}

func readFileOptional(path string) ([]byte, error) {
	//nolint:gosec // G304: path is the resolved AWS config or credentials file location.
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, newError(KindRead, path, 0, "", err)
	}

	defer func() {
		_ = file.Close()
	}()

	info, err := file.Stat()
	if err != nil {
		return nil, newError(KindRead, path, 0, "", err)
	}

	if !info.Mode().IsRegular() {
		return nil, newError(KindRead, path, 0, "", ErrNotRegular)
	}

	if info.Size() > maxFileSize {
		return nil, newError(KindRead, path, 0, "", ErrTooLarge)
	}

	data, err := io.ReadAll(io.LimitReader(file, maxFileSize+1))
	if err != nil {
		return nil, newError(KindRead, path, 0, "", err)
	}

	if int64(len(data)) > maxFileSize {
		return nil, newError(KindRead, path, 0, "", ErrTooLarge)
	}

	return data, nil
}
