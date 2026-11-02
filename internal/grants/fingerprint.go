package grants

import (
	"aws-cli-guard/internal/awsconfig"
	"errors"
	"fmt"
	"strconv"
)

// ErrEmptyProfile is returned when Fingerprint is called with an empty profile name.
var ErrEmptyProfile = errors.New("aws-cli-guard: grants: empty profile name")

// Equal reports whether two credential-source fingerprints match.
func (s Source) Equal(other Source) bool {
	if !sourceScalarsEqual(s, other) {
		return false
	}

	leftChain := s.Chain
	rightChain := other.Chain

	if len(leftChain) != len(rightChain) {
		return false
	}

	for i := range leftChain {
		if !leftChain[i].equal(rightChain[i]) {
			return false
		}
	}

	return true
}

func sourceScalarsEqual(left, other Source) bool {
	return left.AccountID == other.AccountID &&
		left.RoleARN == other.RoleARN &&
		left.SSOAccountID == other.SSOAccountID &&
		left.SSORoleName == other.SSORoleName &&
		left.CredentialProcess == other.CredentialProcess &&
		left.SourceProfile == other.SourceProfile &&
		left.CredentialSource == other.CredentialSource &&
		left.SSOSession == other.SSOSession &&
		left.SSOStartURL == other.SSOStartURL &&
		left.SSORegion == other.SSORegion &&
		left.ExternalID == other.ExternalID
}

func (h ChainHop) equal(other ChainHop) bool {
	return h.Profile == other.Profile &&
		h.AccountID == other.AccountID &&
		h.RoleARN == other.RoleARN &&
		h.SSOAccountID == other.SSOAccountID &&
		h.SSORoleName == other.SSORoleName &&
		h.CredentialProcess == other.CredentialProcess &&
		h.SourceProfile == other.SourceProfile &&
		h.CredentialSource == other.CredentialSource &&
		h.SSOSession == other.SSOSession &&
		h.SSOStartURL == other.SSOStartURL &&
		h.SSORegion == other.SSORegion &&
		h.ExternalID == other.ExternalID
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

	if _, ok := cfg.Profiles[profile]; !ok {
		return emptySource(), nil
	}

	resolved, err := awsconfig.ResolveProfile(cfg, profile)
	if err != nil {
		return emptySource(), mapResolveError(err)
	}

	return sourceFromResolved(resolved)
}

func mapResolveError(err error) error {
	var cfgErr *awsconfig.Error
	if errors.As(err, &cfgErr) && cfgErr.Kind == awsconfig.KindResolve {
		return newError(KindResolve, "", cfgErr.Key, 0, 0, err)
	}

	return fmt.Errorf("aws-cli-guard: grants: load aws config: %w", err)
}

func emptySource() Source {
	return Source{
		AccountID:         "",
		RoleARN:           "",
		SSOAccountID:      "",
		SSORoleName:       "",
		CredentialProcess: false,
		SourceProfile:     "",
		CredentialSource:  "",
		SSOSession:        "",
		SSOStartURL:       "",
		SSORegion:         "",
		ExternalID:        "",
		Chain:             nil,
	}
}

func sourceFromResolved(res awsconfig.ResolvedSource) (Source, error) {
	out := Source{
		AccountID:         res.AccountID,
		RoleARN:           res.RoleARN,
		SSOAccountID:      res.SSOAccountID,
		SSORoleName:       res.SSORoleName,
		CredentialProcess: res.CredentialProcess,
		SourceProfile:     res.SourceProfile,
		CredentialSource:  res.CredentialSource,
		SSOSession:        res.SSOSession,
		SSOStartURL:       res.SSOStartURL,
		SSORegion:         res.SSORegion,
		ExternalID:        res.ExternalID,
		Chain:             make([]ChainHop, len(res.Chain)),
	}

	err := validateResolvedStrings(out)
	if err != nil {
		return Source{}, err
	}

	for hopIndex, hop := range res.Chain {
		out.Chain[hopIndex] = ChainHop{
			Profile:           hop.Profile,
			AccountID:         hop.AccountID,
			RoleARN:           hop.RoleARN,
			SSOAccountID:      hop.SSOAccountID,
			SSORoleName:       hop.SSORoleName,
			CredentialProcess: hop.CredentialProcess,
			SourceProfile:     hop.SourceProfile,
			CredentialSource:  hop.CredentialSource,
			SSOSession:        hop.SSOSession,
			SSOStartURL:       hop.SSOStartURL,
			SSORegion:         hop.SSORegion,
			ExternalID:        hop.ExternalID,
		}

		err = validateHopFields(out.Chain[hopIndex], "source.chain."+strconv.Itoa(hopIndex))
		if err != nil {
			return Source{}, err
		}
	}

	return out, nil
}

func validateResolvedStrings(out Source) error {
	prefix := "source"

	fields := []struct {
		val string
		key string
	}{
		{out.AccountID, sourceKeyAccountID},
		{out.RoleARN, sourceKeyRoleARN},
		{out.SSOAccountID, sourceKeySSOAccountID},
		{out.SSORoleName, sourceKeySSORoleName},
		{out.SourceProfile, sourceKeySourceProfile},
		{out.CredentialSource, sourceKeyCredentialSource},
		{out.SSOSession, sourceKeySSOSession},
		{out.SSOStartURL, sourceKeySSOStartURL},
		{out.SSORegion, sourceKeySSORegion},
		{out.ExternalID, sourceKeyExternalID},
	}

	for _, item := range fields {
		err := validateNoControl(item.val, prefix+"."+item.key)
		if err != nil {
			return err
		}
	}

	return nil
}
