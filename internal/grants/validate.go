package grants

import (
	"strconv"
	"time"
	"unicode"
)

func validateDocument(path string, doc document) (File, error) {
	if doc.Grants == nil {
		return File{}, newError(KindInvalid, "", "grants", 0, 0, nil)
	}

	out := File{
		Path:    path,
		Version: doc.Version,
		Mode:    0,
		Grants:  make([]Grant, 0, len(*doc.Grants)),
	}

	seen := make(map[string]struct{}, len(*doc.Grants))

	for i, item := range *doc.Grants {
		keyPrefix := "grants." + strconv.Itoa(i)

		grantOut, err := validateGrantItem(item, keyPrefix)
		if err != nil {
			return File{}, stampInvalidError(err, path, doc.Version)
		}

		if _, dup := seen[grantOut.Profile]; dup {
			return File{}, stampInvalidError(
				newError(KindInvalid, "", keyPrefix+".profile", 0, 0, nil),
				path,
				doc.Version,
			)
		}

		seen[grantOut.Profile] = struct{}{}
		out.Grants = append(out.Grants, grantOut)
	}

	return out, nil
}

func validateGrantItem(item grantDocument, keyPrefix string) (Grant, error) {
	if item.Profile == nil {
		return Grant{}, newError(KindInvalid, "", keyPrefix+".profile", 0, 0, nil)
	}

	profileErr := validateProfileName(*item.Profile, keyPrefix+".profile")
	if profileErr != nil {
		return Grant{}, profileErr
	}

	if item.Expires == nil {
		return Grant{}, newError(KindInvalid, "", keyPrefix+".expires", 0, 0, nil)
	}

	expires, expErr := parseExpires(*item.Expires, keyPrefix+".expires")
	if expErr != nil {
		return Grant{}, expErr
	}

	if item.Source == nil {
		return Grant{}, newError(KindInvalid, "", keyPrefix+".source", 0, 0, nil)
	}

	source, srcErr := validateSource(*item.Source, keyPrefix+".source")
	if srcErr != nil {
		return Grant{}, srcErr
	}

	return Grant{
		Profile: *item.Profile,
		Expires: expires,
		Source:  source,
	}, nil
}

func validateSource(src sourceDocument, keyPrefix string) (Source, error) {
	out := Source{}

	if src.AccountID == nil {
		return Source{}, newError(KindInvalid, "", keyPrefix+".account_id", 0, 0, nil)
	}

	err := validateNoControl(*src.AccountID, keyPrefix+".account_id")
	if err != nil {
		return Source{}, err
	}

	out.AccountID = *src.AccountID

	if src.RoleARN == nil {
		return Source{}, newError(KindInvalid, "", keyPrefix+".role_arn", 0, 0, nil)
	}

	err = validateNoControl(*src.RoleARN, keyPrefix+".role_arn")
	if err != nil {
		return Source{}, err
	}

	out.RoleARN = *src.RoleARN

	if src.SSOAccountID == nil {
		return Source{}, newError(KindInvalid, "", keyPrefix+".sso_account_id", 0, 0, nil)
	}

	err = validateNoControl(*src.SSOAccountID, keyPrefix+".sso_account_id")
	if err != nil {
		return Source{}, err
	}

	out.SSOAccountID = *src.SSOAccountID

	if src.SSORoleName == nil {
		return Source{}, newError(KindInvalid, "", keyPrefix+".sso_role_name", 0, 0, nil)
	}

	err = validateNoControl(*src.SSORoleName, keyPrefix+".sso_role_name")
	if err != nil {
		return Source{}, err
	}

	out.SSORoleName = *src.SSORoleName

	if src.CredentialProcess == nil {
		return Source{}, newError(KindInvalid, "", keyPrefix+".credential_process", 0, 0, nil)
	}

	out.CredentialProcess = *src.CredentialProcess

	return out, nil
}

func validateProfileName(name, key string) error {
	if name == "" {
		return newError(KindInvalid, "", key, 0, 0, nil)
	}

	for _, r := range name {
		if unicode.IsControl(r) {
			return newError(KindInvalid, "", key, 0, 0, nil)
		}
	}

	return nil
}

func validateNoControl(value, key string) error {
	for _, r := range value {
		if unicode.IsControl(r) {
			return newError(KindInvalid, "", key, 0, 0, nil)
		}
	}

	return nil
}

func parseExpires(raw, key string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}, newError(KindInvalid, "", key, 0, 0, nil)
	}

	trunc := parsed.UTC().Truncate(time.Second)

	return trunc, nil
}

func validateGrantList(path string, list []Grant) error {
	if list == nil {
		list = []Grant{}
	}

	seen := make(map[string]struct{}, len(list))

	for i, grant := range list {
		keyPrefix := "grants." + strconv.Itoa(i)

		nameErr := validateProfileName(grant.Profile, keyPrefix+".profile")
		if nameErr != nil {
			return stampInvalidError(nameErr, path, supportedVersion)
		}

		if _, dup := seen[grant.Profile]; dup {
			return stampInvalidError(
				newError(KindInvalid, "", keyPrefix+".profile", 0, 0, nil),
				path,
				supportedVersion,
			)
		}

		seen[grant.Profile] = struct{}{}

		if grant.Expires.IsZero() {
			return stampInvalidError(
				newError(KindInvalid, "", keyPrefix+".expires", 0, 0, nil),
				path,
				supportedVersion,
			)
		}

		srcErr := validateSourceFields(grant.Source, keyPrefix+".source")
		if srcErr != nil {
			return stampInvalidError(srcErr, path, supportedVersion)
		}
	}

	return nil
}

func validateSourceFields(src Source, keyPrefix string) error {
	err := validateNoControl(src.AccountID, keyPrefix+".account_id")
	if err != nil {
		return err
	}

	err = validateNoControl(src.RoleARN, keyPrefix+".role_arn")
	if err != nil {
		return err
	}

	err = validateNoControl(src.SSOAccountID, keyPrefix+".sso_account_id")
	if err != nil {
		return err
	}

	err = validateNoControl(src.SSORoleName, keyPrefix+".sso_role_name")
	if err != nil {
		return err
	}

	return nil
}
