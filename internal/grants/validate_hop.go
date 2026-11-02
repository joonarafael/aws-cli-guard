package grants

import "strconv"

func validateHop(hop hopDocument, keyPrefix string) (ChainHop, error) {
	if hop.Profile == nil {
		return ChainHop{}, newError(KindInvalid, "", keyPrefix+".profile", 0, 0, nil)
	}

	profileErr := validateProfileName(*hop.Profile, keyPrefix+".profile")
	if profileErr != nil {
		return ChainHop{}, profileErr
	}

	out := emptyChainHop(*hop.Profile)

	if hop.CredentialProcess == nil {
		return ChainHop{}, newError(
			KindInvalid,
			"",
			keyPrefix+"."+sourceKeyCredentialProcess,
			0,
			0,
			nil,
		)
	}

	out.CredentialProcess = *hop.CredentialProcess

	fields := []struct {
		ptr  *string
		key  string
		dest *string
	}{
		{hop.AccountID, sourceKeyAccountID, &out.AccountID},
		{hop.RoleARN, sourceKeyRoleARN, &out.RoleARN},
		{hop.SSOAccountID, sourceKeySSOAccountID, &out.SSOAccountID},
		{hop.SSORoleName, sourceKeySSORoleName, &out.SSORoleName},
		{hop.SourceProfile, sourceKeySourceProfile, &out.SourceProfile},
		{hop.CredentialSource, sourceKeyCredentialSource, &out.CredentialSource},
		{hop.SSOSession, sourceKeySSOSession, &out.SSOSession},
		{hop.SSOStartURL, sourceKeySSOStartURL, &out.SSOStartURL},
		{hop.SSORegion, sourceKeySSORegion, &out.SSORegion},
		{hop.ExternalID, sourceKeyExternalID, &out.ExternalID},
	}

	for _, field := range fields {
		if field.ptr == nil {
			return ChainHop{}, newError(KindInvalid, "", keyPrefix+"."+field.key, 0, 0, nil)
		}

		err := validateNoControl(*field.ptr, keyPrefix+"."+field.key)
		if err != nil {
			return ChainHop{}, err
		}

		*field.dest = *field.ptr
	}

	return out, nil
}

func validateSourceFields(src Source, keyPrefix string) error {
	err := validateNoControl(src.AccountID, keyPrefix+"."+sourceKeyAccountID)
	if err != nil {
		return err
	}

	err = validateNoControl(src.RoleARN, keyPrefix+"."+sourceKeyRoleARN)
	if err != nil {
		return err
	}

	err = validateNoControl(src.SSOAccountID, keyPrefix+"."+sourceKeySSOAccountID)
	if err != nil {
		return err
	}

	err = validateNoControl(src.SSORoleName, keyPrefix+"."+sourceKeySSORoleName)
	if err != nil {
		return err
	}

	v2Strings := []struct {
		val string
		key string
	}{
		{src.SourceProfile, sourceKeySourceProfile},
		{src.CredentialSource, sourceKeyCredentialSource},
		{src.SSOSession, sourceKeySSOSession},
		{src.SSOStartURL, sourceKeySSOStartURL},
		{src.SSORegion, sourceKeySSORegion},
		{src.ExternalID, sourceKeyExternalID},
	}

	for _, item := range v2Strings {
		err = validateNoControl(item.val, keyPrefix+"."+item.key)
		if err != nil {
			return err
		}
	}

	for i, hop := range src.Chain {
		err = validateHopFields(hop, keyPrefix+".chain."+strconv.Itoa(i))
		if err != nil {
			return err
		}
	}

	return nil
}

func validateHopFields(hop ChainHop, keyPrefix string) error {
	nameErr := validateProfileName(hop.Profile, keyPrefix+".profile")
	if nameErr != nil {
		return nameErr
	}

	fields := []struct {
		val string
		key string
	}{
		{hop.AccountID, sourceKeyAccountID},
		{hop.RoleARN, sourceKeyRoleARN},
		{hop.SSOAccountID, sourceKeySSOAccountID},
		{hop.SSORoleName, sourceKeySSORoleName},
		{hop.SourceProfile, sourceKeySourceProfile},
		{hop.CredentialSource, sourceKeyCredentialSource},
		{hop.SSOSession, sourceKeySSOSession},
		{hop.SSOStartURL, sourceKeySSOStartURL},
		{hop.SSORegion, sourceKeySSORegion},
		{hop.ExternalID, sourceKeyExternalID},
	}

	for _, item := range fields {
		err := validateNoControl(item.val, keyPrefix+"."+item.key)
		if err != nil {
			return err
		}
	}

	return nil
}
