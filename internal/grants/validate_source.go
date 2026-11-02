package grants

import (
	"strconv"
)

func validateSource(src sourceDocument, keyPrefix string, fileVersion int) (Source, error) {
	out, err := validateSourceBase(src, keyPrefix)
	if err != nil {
		return Source{}, err
	}

	if fileVersion == 1 {
		if sourceHasV2Fields(src) {
			key := firstPresentV2Key(src)

			return Source{}, newError(KindInvalid, "", key, 0, 0, nil)
		}

		return out, nil
	}

	return validateSourceV2(src, keyPrefix, out)
}

func validateSourceBase(src sourceDocument, keyPrefix string) (Source, error) {
	out := Source{}

	if src.AccountID == nil {
		return Source{}, newError(KindInvalid, "", keyPrefix+"."+sourceKeyAccountID, 0, 0, nil)
	}

	err := validateNoControl(*src.AccountID, keyPrefix+"."+sourceKeyAccountID)
	if err != nil {
		return Source{}, err
	}

	out.AccountID = *src.AccountID

	if src.RoleARN == nil {
		return Source{}, newError(KindInvalid, "", keyPrefix+"."+sourceKeyRoleARN, 0, 0, nil)
	}

	err = validateNoControl(*src.RoleARN, keyPrefix+"."+sourceKeyRoleARN)
	if err != nil {
		return Source{}, err
	}

	out.RoleARN = *src.RoleARN

	if src.SSOAccountID == nil {
		return Source{}, newError(KindInvalid, "", keyPrefix+"."+sourceKeySSOAccountID, 0, 0, nil)
	}

	err = validateNoControl(*src.SSOAccountID, keyPrefix+"."+sourceKeySSOAccountID)
	if err != nil {
		return Source{}, err
	}

	out.SSOAccountID = *src.SSOAccountID

	if src.SSORoleName == nil {
		return Source{}, newError(KindInvalid, "", keyPrefix+"."+sourceKeySSORoleName, 0, 0, nil)
	}

	err = validateNoControl(*src.SSORoleName, keyPrefix+"."+sourceKeySSORoleName)
	if err != nil {
		return Source{}, err
	}

	out.SSORoleName = *src.SSORoleName

	if src.CredentialProcess == nil {
		return Source{}, newError(
			KindInvalid,
			"",
			keyPrefix+"."+sourceKeyCredentialProcess,
			0,
			0,
			nil,
		)
	}

	out.CredentialProcess = *src.CredentialProcess

	return out, nil
}

func sourceHasV2Fields(src sourceDocument) bool {
	return src.SourceProfile != nil ||
		src.CredentialSource != nil ||
		src.SSOSession != nil ||
		src.SSOStartURL != nil ||
		src.SSORegion != nil ||
		src.ExternalID != nil ||
		src.Chain != nil
}

func firstPresentV2Key(src sourceDocument) string {
	switch {
	case src.SourceProfile != nil:
		return "source." + sourceKeySourceProfile
	case src.CredentialSource != nil:
		return "source." + sourceKeyCredentialSource
	case src.SSOSession != nil:
		return "source." + sourceKeySSOSession
	case src.SSOStartURL != nil:
		return "source." + sourceKeySSOStartURL
	case src.SSORegion != nil:
		return "source." + sourceKeySSORegion
	case src.ExternalID != nil:
		return "source." + sourceKeyExternalID
	case src.Chain != nil:
		return "source." + sourceKeyChain
	default:
		return "source"
	}
}

func validateSourceV2(src sourceDocument, keyPrefix string, out Source) (Source, error) {
	strings := []struct {
		ptr *string
		key string
	}{
		{src.SourceProfile, sourceKeySourceProfile},
		{src.CredentialSource, sourceKeyCredentialSource},
		{src.SSOSession, sourceKeySSOSession},
		{src.SSOStartURL, sourceKeySSOStartURL},
		{src.SSORegion, sourceKeySSORegion},
		{src.ExternalID, sourceKeyExternalID},
	}

	for _, item := range strings {
		if item.ptr == nil {
			return Source{}, newError(KindInvalid, "", keyPrefix+"."+item.key, 0, 0, nil)
		}

		err := validateNoControl(*item.ptr, keyPrefix+"."+item.key)
		if err != nil {
			return Source{}, err
		}
	}

	out.SourceProfile = *src.SourceProfile
	out.CredentialSource = *src.CredentialSource
	out.SSOSession = *src.SSOSession
	out.SSOStartURL = *src.SSOStartURL
	out.SSORegion = *src.SSORegion
	out.ExternalID = *src.ExternalID

	if src.Chain == nil {
		return Source{}, newError(KindInvalid, "", keyPrefix+"."+sourceKeyChain, 0, 0, nil)
	}

	out.Chain = make([]ChainHop, 0, len(*src.Chain))

	for i, hop := range *src.Chain {
		hopOut, hopErr := validateHop(hop, keyPrefix+".chain."+strconv.Itoa(i))
		if hopErr != nil {
			return Source{}, hopErr
		}

		out.Chain = append(out.Chain, hopOut)
	}

	return out, nil
}
