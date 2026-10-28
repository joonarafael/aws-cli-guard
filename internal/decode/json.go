package decode

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
)

const jsonUnknownFieldMatchLen = 2

var jsonUnknownFieldRE = regexp.MustCompile(`^json: unknown field "([^"]+)"$`)

func decodeJSON(data []byte, dest any, supported []int) error {
	if len(data) == 0 || isOnlyWhitespace(data) {
		return newDocError(KindSyntax, ErrSyntax, "", 0, 0, 0, false, supported)
	}

	scanRes, scanErr := scanJSONDocument(data, supported)
	if scanErr != nil {
		return scanErr
	}

	verErr := validateJSONScanVersion(scanRes, supported)
	if verErr != nil {
		return verErr
	}

	strictDec := json.NewDecoder(bytes.NewReader(data))
	strictDec.DisallowUnknownFields()

	err := strictDec.Decode(dest)
	if err != nil {
		return mapJSONStrictError(data, err, supported, scanRes)
	}

	return nil
}

func validateJSONScanVersion(scanRes jsonScanResult, supported []int) error {
	if scanRes.hasDuplicate {
		return newDocError(
			KindDuplicateKey,
			ErrDuplicateKey,
			scanRes.duplicateKey,
			scanRes.duplicatePos.line,
			scanRes.duplicatePos.column,
			0,
			false,
			supported,
		)
	}

	if !scanRes.hasVersion {
		return newDocError(
			KindMissingVersion,
			ErrMissingVersion,
			versionFieldKey,
			0,
			0,
			0,
			false,
			supported,
		)
	}

	version, verOK := classifyJSONVersionText(scanRes.versionText)
	if !verOK {
		return newDocError(
			KindVersionType,
			ErrVersionType,
			versionFieldKey,
			scanRes.versionLine,
			scanRes.versionCol,
			0,
			false,
			supported,
		)
	}

	if !versionSupported(version, supported) {
		return newDocError(
			KindUnsupportedVersion,
			ErrUnsupportedVersion,
			versionFieldKey,
			scanRes.versionLine,
			scanRes.versionCol,
			version,
			true,
			supported,
		)
	}

	return nil
}

func mapJSONStrictError(data []byte, err error, supported []int, scan jsonScanResult) error {
	if typeErr, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
		key := typeErr.Field
		pos := offsetToLineColumn(data, typeErr.Offset)

		if key == versionFieldKey {
			return newDocError(
				KindVersionType,
				ErrVersionType,
				key,
				pos.line,
				pos.column,
				0,
				false,
				supported,
			)
		}

		return newDocError(KindType, ErrType, key, pos.line, pos.column, 0, false, supported)
	}

	msg := err.Error()

	matches := jsonUnknownFieldRE.FindStringSubmatch(msg)
	if len(matches) == jsonUnknownFieldMatchLen {
		key := matches[1]

		line, col := 0, 0
		if pos, found := scan.keyPositions[key]; found {
			line, col = pos.line, pos.column
		}

		return newDocError(KindUnknownField, ErrUnknownField, key, line, col, 0, false, supported)
	}

	if syn, ok := errors.AsType[*json.SyntaxError](err); ok {
		pos := offsetToLineColumn(data, syn.Offset)

		return newDocError(KindSyntax, ErrSyntax, "", pos.line, pos.column, 0, false, supported)
	}

	return newDocError(KindSyntax, ErrSyntax, "", 0, 0, 0, false, supported)
}
