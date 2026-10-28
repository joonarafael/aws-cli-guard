package decode

import (
	"bytes"
	"errors"
	"math"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

func decodeTOML(data []byte, dest any, supported []int) error {
	if len(data) == 0 || isOnlyWhitespace(data) {
		return newDocError(KindSyntax, ErrSyntax, "", 0, 0, 0, false, supported)
	}

	var top map[string]any

	dec := toml.NewDecoder(bytes.NewReader(data))

	err := dec.Decode(&top)
	if err != nil {
		return mapTOMLSyntaxError(err, supported)
	}

	verErr := checkTOMLTopVersion(top, supported)
	if verErr != nil {
		return verErr
	}

	strictDec := toml.NewDecoder(bytes.NewReader(data))
	strictDec.DisallowUnknownFields()

	err = strictDec.Decode(dest)
	if err != nil {
		return mapTOMLStrictError(err, supported)
	}

	return nil
}

func mapTOMLSyntaxError(err error, supported []int) error {
	if decErr, ok := errors.AsType[*toml.DecodeError](err); ok {
		line, col := decErr.Position()
		key := joinKey(decErr.Key())

		msg := decErr.Error()
		if strings.Contains(msg, "already defined") {
			return newDocError(
				KindDuplicateKey,
				ErrDuplicateKey,
				key,
				line,
				col,
				0,
				false,
				supported,
			)
		}

		return newDocError(KindSyntax, ErrSyntax, key, line, col, 0, false, supported)
	}

	return newDocError(KindSyntax, ErrSyntax, "", 0, 0, 0, false, supported)
}

func checkTOMLTopVersion(top map[string]any, supported []int) *Error {
	raw, present := top[versionFieldKey]
	if !present {
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

	version, valid := intFromTOMLValue(raw)
	if !valid {
		return newDocError(
			KindVersionType,
			ErrVersionType,
			versionFieldKey,
			0,
			0,
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
			0,
			0,
			version,
			true,
			supported,
		)
	}

	return nil
}

func intFromTOMLValue(raw any) (int, bool) {
	switch val := raw.(type) {
	case int64:
		if val < math.MinInt || val > math.MaxInt {
			return 0, false
		}

		return int(val), true
	case int:
		return val, true
	default:
		return 0, false
	}
}

func mapTOMLStrictError(err error, supported []int) error {
	if strictErr, ok := errors.AsType[*toml.StrictMissingError](err); ok {
		unwraps := strictErr.Unwrap()
		if len(unwraps) > 0 {
			if first, ok := errors.AsType[*toml.DecodeError](unwraps[0]); ok {
				line, col := first.Position()
				key := joinKey(first.Key())

				return newDocError(
					KindUnknownField,
					ErrUnknownField,
					key,
					line,
					col,
					0,
					false,
					supported,
				)
			}
		}

		return newDocError(KindUnknownField, ErrUnknownField, "", 0, 0, 0, false, supported)
	}

	if decErr, ok := errors.AsType[*toml.DecodeError](err); ok {
		line, col := decErr.Position()
		key := joinKey(decErr.Key())

		if key == versionFieldKey || strings.HasSuffix(key, "."+versionFieldKey) {
			return newDocError(KindVersionType, ErrVersionType, key, line, col, 0, false, supported)
		}

		return newDocError(KindType, ErrType, key, line, col, 0, false, supported)
	}

	return newDocError(KindSyntax, ErrSyntax, "", 0, 0, 0, false, supported)
}
