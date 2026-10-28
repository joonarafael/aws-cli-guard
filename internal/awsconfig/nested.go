package awsconfig

import (
	"strings"
)

func materializeSection(path string, sec iniSection) (Section, error) {
	out := emptySection()

	for key, val := range sec.options {
		if strings.HasPrefix(val, "\n") {
			nested, err := parseNestedBlock(path, key, val)
			if err != nil {
				return Section{}, err
			}

			redactNestedIntoSection(&out, key, nested)

			continue
		}

		redactSettingIntoSection(&out, key, val)
	}

	return out, nil
}

func redactSettingIntoSection(sec *Section, key, val string) {
	switch key {
	case "aws_access_key_id":
		sec.HasAccessKeyID = true
	case "aws_secret_access_key":
		sec.HasSecretAccessKey = true
	case "aws_session_token", "aws_security_token":
		sec.HasSessionToken = true
	case "credential_process":
		sec.HasCredentialProcess = true
	default:
		sec.Settings[key] = val
	}
}

func redactNestedIntoSection(sec *Section, block string, nested map[string]string) {
	clean := map[string]string{}

	for key, val := range nested {
		switch key {
		case "aws_access_key_id":
			sec.HasAccessKeyID = true
		case "aws_secret_access_key":
			sec.HasSecretAccessKey = true
		case "aws_session_token", "aws_security_token":
			sec.HasSessionToken = true
		case "credential_process":
			sec.HasCredentialProcess = true
		default:
			clean[key] = val
		}
	}

	if len(clean) > 0 {
		sec.Nested[block] = clean
	}
}

func parseNestedBlock(path, key, val string) (map[string]string, error) {
	out := map[string]string{}

	for line := range strings.SplitSeq(val, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		eqIndex := strings.IndexByte(line, '=')
		if eqIndex < 0 {
			return nil, newError(KindParse, path, 0, key, errNestedLineWithoutEquals)
		}

		nKey := strings.TrimSpace(line[:eqIndex])
		nVal := strings.TrimSpace(line[eqIndex+1:])

		if nKey == "" {
			return nil, newError(KindParse, path, 0, key, errEmptyNestedKey)
		}

		nKey = strings.ToLower(nKey)
		if _, dup := out[nKey]; dup {
			return nil, newError(KindParse, path, 0, nKey, errDuplicateNestedKey)
		}

		out[nKey] = nVal
	}

	return out, nil
}
