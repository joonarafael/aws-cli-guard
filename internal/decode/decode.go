package decode

import (
	"reflect"
	"slices"
	"strings"
)

const versionFieldKey = "version"

// Format selects TOML or JSON decoding.
type Format int

const (
	// TOML decodes a TOML document.
	TOML Format = iota + 1
	// JSON decodes a JSON document.
	JSON
)

// Decode strictly decodes data into dest. dest must be a non-nil pointer to a struct
// with exactly one exported int field tagged version for the chosen format.
func Decode(format Format, data []byte, dest any, supported []int) error {
	if format != TOML && format != JSON {
		return ErrFormat
	}

	if len(supported) == 0 {
		return ErrSupported
	}

	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Pointer || destVal.IsNil() {
		return ErrDest
	}

	elem := destVal.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrDest
	}

	tagName := versionTagName(format)
	if !hasVersionIntField(elem.Type(), tagName) {
		return ErrVersionField
	}

	switch format {
	case TOML:
		return decodeTOML(data, dest, supported)
	case JSON:
		return decodeJSON(data, dest, supported)
	default:
		return ErrFormat
	}
}

func versionTagName(format Format) string {
	if format == JSON {
		return "json"
	}

	return "toml"
}

func hasVersionIntField(structType reflect.Type, tag string) bool {
	found := false

	for field := range structType.Fields() {
		if !field.IsExported() {
			continue
		}

		if field.Type.Kind() != reflect.Int {
			continue
		}

		tagVal, ok := field.Tag.Lookup(tag)
		if !ok {
			continue
		}

		name, _, _ := strings.Cut(tagVal, ",")
		if name != versionFieldKey {
			continue
		}

		if found {
			return false
		}

		found = true
	}

	return found
}

func versionSupported(version int, supported []int) bool {
	return slices.Contains(supported, version)
}

func isOnlyWhitespace(data []byte) bool {
	for _, b := range data {
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			return false
		}
	}

	return true
}
