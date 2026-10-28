package decode

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

var errScanStop = errors.New("aws-cli-guard: decode: scan stop")

type jsonKeyPos struct {
	line   int
	column int
}

type jsonScanResult struct {
	duplicateKey string
	duplicatePos jsonKeyPos
	hasDuplicate bool
	versionText  string
	versionLine  int
	versionCol   int
	hasVersion   bool
	keyPositions map[string]jsonKeyPos
}

func scanJSONDocument(data []byte, supported []int) (jsonScanResult, *Error) {
	var res jsonScanResult

	res.keyPositions = make(map[string]jsonKeyPos)

	if len(data) == 0 || isOnlyWhitespace(data) {
		return res, newDocError(KindSyntax, ErrSyntax, "", 0, 0, 0, false, supported)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	scanner := &jsonScanner{
		data:      data,
		dec:       dec,
		path:      nil,
		keyStack:  nil,
		supported: supported,
		result:    &res,
	}
	scanner.pushObject()

	tok, err := dec.Token()
	if err != nil {
		return res, mapJSONTokenError(data, err, supported)
	}

	delim, ok := tok.(json.Delim)
	if !ok || delim != json.Delim('{') {
		return res, newDocError(KindSyntax, ErrSyntax, "", 0, 0, 0, false, supported)
	}

	readErr := scanner.readObjectContents()
	if readErr != nil {
		if scanner.result.hasDuplicate {
			return res, nil
		}

		if docErr, ok := errors.AsType[*Error](readErr); ok {
			return res, docErr
		}

		return res, mapJSONTokenError(data, readErr, supported)
	}

	return res, nil
}

type jsonScanner struct {
	data      []byte
	dec       *json.Decoder
	path      []string
	keyStack  []map[string]jsonKeyPos
	supported []int
	result    *jsonScanResult
}

func (s *jsonScanner) pushObject() {
	s.keyStack = append(s.keyStack, make(map[string]jsonKeyPos))
}

func (s *jsonScanner) popObject() {
	if len(s.keyStack) > 0 {
		s.keyStack = s.keyStack[:len(s.keyStack)-1]
	}
}

func (s *jsonScanner) fullKey(name string) string {
	if len(s.path) == 0 {
		return name
	}

	parts := append(append([]string(nil), s.path...), name)

	return strings.Join(parts, ".")
}

func (s *jsonScanner) noteKey(name string, pos jsonKeyPos) bool {
	keys := s.keyStack[len(s.keyStack)-1]
	if prev, seen := keys[name]; seen {
		s.result.hasDuplicate = true
		s.result.duplicateKey = s.fullKey(name)
		s.result.duplicatePos = prev

		return true
	}

	keys[name] = pos
	full := s.fullKey(name)
	s.result.keyPositions[full] = pos

	return false
}

func (s *jsonScanner) readObjectContents() error {
	for s.dec.More() {
		keyTok, err := s.dec.Token()
		if err != nil {
			return mapJSONTokenError(s.data, err, s.supported)
		}

		key, ok := keyTok.(string)
		if !ok {
			return newDocError(KindSyntax, ErrSyntax, "", 0, 0, 0, false, s.supported)
		}

		pos := offsetToLineColumn(s.data, s.dec.InputOffset())
		if s.noteKey(key, pos) {
			return errScanStop
		}

		if len(s.path) == 0 && key == versionFieldKey {
			err := s.readVersionValue()
			if err != nil {
				return err
			}

			continue
		}

		s.path = append(s.path, key)

		err = s.readValue()
		if err != nil {
			if errors.Is(err, errScanStop) {
				return errScanStop
			}

			return err
		}

		s.path = s.path[:len(s.path)-1]
	}

	_, err := s.dec.Token()
	if err != nil {
		return mapJSONTokenError(s.data, err, s.supported)
	}

	return nil
}

func (s *jsonScanner) readVersionValue() error {
	tok, err := s.dec.Token()
	if err != nil {
		return mapJSONTokenError(s.data, err, s.supported)
	}

	s.result.hasVersion = true
	valPos := offsetToLineColumn(s.data, s.dec.InputOffset())
	s.result.versionLine = valPos.line
	s.result.versionCol = valPos.column

	switch val := tok.(type) {
	case json.Number:
		s.result.versionText = val.String()
	default:
		s.result.versionText = ""
	}

	return nil
}

func (s *jsonScanner) readValue() error {
	tok, err := s.dec.Token()
	if err != nil {
		return mapJSONTokenError(s.data, err, s.supported)
	}

	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}

	switch delim {
	case '{':
		s.pushObject()
		defer s.popObject()

		return s.readObjectContents()
	case '[':
		return s.readArrayContents()
	default:
		return newDocError(KindSyntax, ErrSyntax, "", 0, 0, 0, false, s.supported)
	}
}

func (s *jsonScanner) readArrayContents() error {
	index := 0

	for s.dec.More() {
		s.path = append(s.path, strconv.Itoa(index))

		err := s.readValue()
		if err != nil {
			if errors.Is(err, errScanStop) {
				return errScanStop
			}

			return err
		}

		s.path = s.path[:len(s.path)-1]
		index++
	}

	_, err := s.dec.Token()
	if err != nil {
		return mapJSONTokenError(s.data, err, s.supported)
	}

	return nil
}
