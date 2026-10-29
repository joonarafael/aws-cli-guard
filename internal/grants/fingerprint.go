package grants

import (
	"aws-cli-guard/internal/awsconfig"
	"errors"
	"fmt"
)

// ErrEmptyProfile is returned when Fingerprint is called with an empty profile name.
var ErrEmptyProfile = errors.New("aws-cli-guard: grants: empty profile name")

// Equal reports whether two credential-source fingerprints match.
func (s Source) Equal(other Source) bool {
	return s.AccountID == other.AccountID &&
		s.RoleARN == other.RoleARN &&
		s.SSOAccountID == other.SSOAccountID &&
		s.SSORoleName == other.SSORoleName &&
		s.CredentialProcess == other.CredentialProcess
}

// Fingerprint reads the credential-source fingerprint for profile from AWS config.
func Fingerprint(env []string, profile string) (Source, error) {
	if profile == "" {
		return emptySource(), ErrEmptyProfile
	}

	cfg, err := awsconfig.Load(env)
	if err != nil {
		return emptySource(), fmt.Errorf("aws-cli-guard: grants: load aws config: %w", err)
	}

	sec, ok := cfg.Profiles[profile]
	if !ok {
		return emptySource(), nil
	}

	return sourceFromSection(sec)
}

func emptySource() Source {
	return Source{
		AccountID:         "",
		RoleARN:           "",
		SSOAccountID:      "",
		SSORoleName:       "",
		CredentialProcess: false,
	}
}

func sourceFromSection(sec awsconfig.Section) (Source, error) {
	out := Source{
		AccountID:         "",
		RoleARN:           "",
		SSOAccountID:      "",
		SSORoleName:       "",
		CredentialProcess: sec.HasCredentialProcess,
	}

	var err error

	out.AccountID, err = copySetting(sec, "account_id", "source.account_id")
	if err != nil {
		return Source{}, err
	}

	out.RoleARN, err = copySetting(sec, "role_arn", "source.role_arn")
	if err != nil {
		return Source{}, err
	}

	out.SSOAccountID, err = copySetting(sec, "sso_account_id", "source.sso_account_id")
	if err != nil {
		return Source{}, err
	}

	out.SSORoleName, err = copySetting(sec, "sso_role_name", "source.sso_role_name")
	if err != nil {
		return Source{}, err
	}

	return out, nil
}

func copySetting(sec awsconfig.Section, setting, key string) (string, error) {
	val, ok := sec.Settings[setting]
	if !ok {
		return "", nil
	}

	err := validateNoControl(val, key)
	if err != nil {
		return "", err
	}

	return val, nil
}
