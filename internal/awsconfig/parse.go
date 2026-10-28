package awsconfig

import (
	"strings"
	"unicode/utf8"
)

const defaultSectionName = "DEFAULT"

type iniDocument struct {
	defaults   *iniSection
	sections   map[string]*iniSection
	sectionOrd []string
}

type iniSection struct {
	options map[string]string
}

type iniParser struct {
	path         string
	doc          iniDocument
	current      *iniSection
	lastKey      string
	seenSections map[string]int
}

func parseINI(path string, data []byte) (iniDocument, error) {
	if !utf8.Valid(data) {
		return iniDocument{}, newError(KindParse, path, 0, "", errInvalidUTF8)
	}

	parser := iniParser{
		path: path,
		doc: iniDocument{
			defaults:   &iniSection{options: map[string]string{}},
			sections:   map[string]*iniSection{},
			sectionOrd: []string{},
		},
		current:      nil,
		lastKey:      "",
		seenSections: map[string]int{},
	}

	lines := splitLines(string(data))
	for lineNum, line := range lines {
		err := parser.parseLine(lineNum+1, line)
		if err != nil {
			return iniDocument{}, err
		}
	}

	return parser.doc, nil
}

func (p *iniParser) parseLine(lineNum int, line string) error {
	if isCommentLine(line) {
		return nil
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil
	}

	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return p.parseContinuation(lineNum, line)
	}

	if trimmed[0] == '[' {
		return p.parseSection(lineNum, trimmed)
	}

	return p.parseOption(lineNum, trimmed)
}

func (p *iniParser) parseContinuation(lineNum int, line string) error {
	if p.current == nil || p.lastKey == "" {
		return newError(KindParse, p.path, lineNum, "", errContinuationWithoutOpt)
	}

	cont := strings.TrimSpace(line)
	p.current.options[p.lastKey] = p.current.options[p.lastKey] + "\n" + cont

	return nil
}

func (p *iniParser) parseSection(lineNum int, trimmed string) error {
	name, ok := parseSectionHeader(trimmed)
	if !ok {
		return newError(KindParse, p.path, lineNum, "", errInvalidSectionHeader)
	}

	if _, dup := p.seenSections[name]; dup {
		return newError(KindParse, p.path, lineNum, name, errDuplicateSection)
	}

	p.seenSections[name] = lineNum

	if name == defaultSectionName {
		p.current = p.doc.defaults
	} else {
		sec := &iniSection{options: map[string]string{}}
		p.doc.sections[name] = sec
		p.doc.sectionOrd = append(p.doc.sectionOrd, name)
		p.current = sec
	}

	p.lastKey = ""

	return nil
}

func (p *iniParser) parseOption(lineNum int, trimmed string) error {
	key, val, ok := parseOption(trimmed)
	if !ok {
		return newError(KindParse, p.path, lineNum, "", errInvalidOption)
	}

	if p.current == nil {
		return newError(KindParse, p.path, lineNum, key, errOptionOutsideSection)
	}

	key = strings.ToLower(key)
	if _, dup := p.current.options[key]; dup {
		return newError(KindParse, p.path, lineNum, key, errDuplicateKey)
	}

	p.current.options[key] = val
	p.lastKey = key

	return nil
}

func splitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	return strings.Split(text, "\n")
}

func isCommentLine(line string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return false
	}

	return trimmed[0] == '#' || trimmed[0] == ';'
}

func parseSectionHeader(trimmed string) (string, bool) {
	if len(trimmed) < 3 || trimmed[0] != '[' || trimmed[len(trimmed)-1] != ']' {
		return "", false
	}

	inner := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	if inner == "" {
		return "", false
	}

	return inner, true
}

func parseOption(line string) (string, string, bool) {
	sep := strings.IndexAny(line, "=:")
	if sep < 0 {
		return "", "", false
	}

	key := strings.TrimSpace(line[:sep])
	val := strings.TrimSpace(line[sep+1:])

	if key == "" {
		return "", "", false
	}

	return key, val, true
}

func applyDefaults(doc iniDocument) iniDocument {
	for _, name := range doc.sectionOrd {
		sec := doc.sections[name]
		for key, val := range doc.defaults.options {
			if _, set := sec.options[key]; !set {
				sec.options[key] = val
			}
		}
	}

	return doc
}
