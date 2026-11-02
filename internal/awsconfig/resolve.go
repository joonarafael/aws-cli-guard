package awsconfig

const maxChainDepth = 8

// ResolvedHop is one upstream profile in a credential chain, nearest parent first.
type ResolvedHop struct {
	Profile           string
	AccountID         string
	RoleARN           string
	SSOAccountID      string
	SSORoleName       string
	CredentialProcess bool
	SourceProfile     string
	CredentialSource  string
	SSOSession        string
	SSOStartURL       string
	SSORegion         string
	ExternalID        string
}

// ResolvedSource is the credential-source snapshot for one named profile.
type ResolvedSource struct {
	AccountID         string
	RoleARN           string
	SSOAccountID      string
	SSORoleName       string
	CredentialProcess bool
	SourceProfile     string
	CredentialSource  string
	SSOSession        string
	SSOStartURL       string
	SSORegion         string
	ExternalID        string
	Chain             []ResolvedHop
}

// ResolveProfile builds the credential-source snapshot for profile in cfg.
func ResolveProfile(cfg Config, profile string) (ResolvedSource, error) {
	sec, ok := cfg.Profiles[profile]
	if !ok {
		return emptyResolvedSource(), nil
	}

	outer, err := sectionSnapshot(cfg, sec)
	if err != nil {
		return emptyResolvedSource(), err
	}

	chain, err := walkChain(cfg, profile, sec)
	if err != nil {
		return emptyResolvedSource(), err
	}

	outer.Chain = chain

	return outer, nil
}

func walkChain(cfg Config, startProfile string, startSec Section) ([]ResolvedHop, error) {
	sourceProfile := setting(startSec, "source_profile")
	credentialSource := setting(startSec, "credential_source")

	if sourceProfile != "" && credentialSource != "" {
		return nil, newResolveError("source_profile")
	}

	if credentialSource != "" || sourceProfile == "" {
		return nil, nil
	}

	visited := map[string]struct{}{startProfile: {}}
	chain := make([]ResolvedHop, 0, 1)
	current := sourceProfile

	for {
		if _, seen := visited[current]; seen {
			return nil, newResolveError("cycle")
		}

		visited[current] = struct{}{}

		sec, ok := cfg.Profiles[current]
		if !ok {
			return nil, newResolveError("source_profile")
		}

		hop, err := sectionSnapshot(cfg, sec)
		if err != nil {
			return nil, err
		}

		chain = append(chain, hopToResolvedHop(current, hop))

		if hop.CredentialSource != "" || hop.SourceProfile == "" {
			return chain, nil
		}

		if len(chain) >= maxChainDepth {
			return nil, newResolveError("depth")
		}

		current = hop.SourceProfile
	}
}

func sectionSnapshot(cfg Config, sec Section) (ResolvedSource, error) {
	sourceProfile := setting(sec, "source_profile")
	credentialSource := setting(sec, "credential_source")

	if sourceProfile != "" && credentialSource != "" {
		return emptyResolvedSource(), newResolveError("source_profile")
	}

	ssoSession := setting(sec, "sso_session")

	ssoStartURL, ssoRegion, ssoErr := resolveSSORegion(cfg, sec, ssoSession)
	if ssoErr != nil {
		return emptyResolvedSource(), ssoErr
	}

	out := emptyResolvedSource()
	out.AccountID = setting(sec, "account_id")
	out.RoleARN = setting(sec, "role_arn")
	out.SSOAccountID = setting(sec, "sso_account_id")
	out.SSORoleName = setting(sec, "sso_role_name")
	out.CredentialProcess = sec.HasCredentialProcess
	out.SourceProfile = sourceProfile
	out.CredentialSource = credentialSource
	out.SSOSession = ssoSession
	out.SSOStartURL = ssoStartURL
	out.SSORegion = ssoRegion
	out.ExternalID = setting(sec, "external_id")

	return out, nil
}

func resolveSSORegion(cfg Config, sec Section, ssoSession string) (string, string, error) {
	if ssoSession == "" {
		return setting(sec, "sso_start_url"), setting(sec, "sso_region"), nil
	}

	sessionSec, ok := cfg.SSOSessions[ssoSession]
	if !ok {
		return "", "", newResolveError("sso_session")
	}

	return setting(sessionSec, "sso_start_url"), setting(sessionSec, "sso_region"), nil
}

func setting(sec Section, key string) string {
	val, ok := sec.Settings[key]
	if !ok {
		return ""
	}

	return val
}

func hopToResolvedHop(profile string, snap ResolvedSource) ResolvedHop {
	return ResolvedHop{
		Profile:           profile,
		AccountID:         snap.AccountID,
		RoleARN:           snap.RoleARN,
		SSOAccountID:      snap.SSOAccountID,
		SSORoleName:       snap.SSORoleName,
		CredentialProcess: snap.CredentialProcess,
		SourceProfile:     snap.SourceProfile,
		CredentialSource:  snap.CredentialSource,
		SSOSession:        snap.SSOSession,
		SSOStartURL:       snap.SSOStartURL,
		SSORegion:         snap.SSORegion,
		ExternalID:        snap.ExternalID,
	}
}

func newResolveError(key string) *Error {
	return newError(KindResolve, "", 0, key, errCredentialUnresolved)
}
