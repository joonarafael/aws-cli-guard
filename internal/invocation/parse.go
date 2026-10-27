package invocation

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	reasonAmbiguousGlobalOption       = "ambiguous-global-option"
	reasonGlobalOptionUnexpectedValue = "global-option-unexpected-value"
	reasonGlobalOptionMissingValue    = "global-option-missing-value"
	reasonProfileFlagMissingValue     = "profile-flag-missing-value"
	reasonEmptyProfileName            = "empty-profile-name"
	reasonGlobalOptionInvalidValue    = "global-option-invalid-value"
)

var negativeNumberPattern = regexp.MustCompile(`^-\d+$|^-\d*\.\d+$`)

// Invocation is the result of parsing top-level AWS CLI global options from argv.
type Invocation struct {
	Command    string
	Operation  string
	Profile    string
	HasProfile bool
	OK         bool
	Reason     string
}

// Parse interprets args in the AWS CLI 2.9.0 top-level global-option grammar.
// args[0] is the program name and is ignored.
func Parse(args []string) Invocation {
	if len(args) == 0 {
		return successInvocation("", "", "", false)
	}

	state := parseState{}
	positional := state.walkArgs(args)

	return state.finish(positional)
}

func successInvocation(command, operation, profile string, hasProfile bool) Invocation {
	return Invocation{
		Command:    command,
		Operation:  operation,
		Profile:    profile,
		HasProfile: hasProfile,
		OK:         true,
		Reason:     "",
	}
}

func failedInvocation(reason string) Invocation {
	return Invocation{
		Command:    "",
		Operation:  "",
		Profile:    "",
		HasProfile: false,
		OK:         false,
		Reason:     reason,
	}
}

type parseState struct {
	command    string
	operation  string
	profile    string
	hasProfile bool
	failed     bool
	reason     string
}

func (s *parseState) fail(reason string) {
	if s.failed {
		return
	}

	s.failed = true
	s.reason = reason
	s.command = ""
	s.operation = ""
	s.profile = ""
	s.hasProfile = false
}

func (s *parseState) setProfile(name string) {
	if name == "" {
		s.fail(reasonEmptyProfileName)

		return
	}

	s.profile = name
	s.hasProfile = true
}

func (s *parseState) walkArgs(args []string) []string {
	var positional []string

	for argIndex := 1; argIndex < len(args); argIndex++ {
		if s.failed {
			break
		}

		arg := args[argIndex]
		if arg == "--" {
			positional = append(positional, args[argIndex+1:]...)

			break
		}

		if strings.HasPrefix(arg, "--") {
			argIndex = s.consumeLongOption(args, argIndex)

			continue
		}

		if isOptionToken(arg) {
			continue
		}

		positional = append(positional, arg)
	}

	return positional
}

func (s *parseState) consumeLongOption(args []string, argIndex int) int {
	arg := args[argIndex]
	body := arg[2:]
	name, explicitValue, hasEquals := strings.Cut(body, "=")

	matched, status := matchGlobalOption(name)
	switch status {
	case matchUnknown:
		return argIndex
	case matchAmbiguous:
		s.fail(reasonAmbiguousGlobalOption)

		return argIndex
	case matchOK:
		return s.applyOption(args, argIndex, matched.option, explicitValue, hasEquals)
	default:
		return argIndex
	}
}

func (s *parseState) applyOption(
	args []string,
	argIndex int,
	opt globalOption,
	explicitValue string,
	hasEquals bool,
) int {
	if opt.kind == optFlag {
		if hasEquals {
			s.fail(reasonGlobalOptionUnexpectedValue)

			return argIndex
		}

		return argIndex
	}

	if hasEquals {
		return s.applyExplicitValue(opt, explicitValue, argIndex)
	}

	return s.applySeparateValue(args, argIndex, opt)
}

func (s *parseState) applyExplicitValue(opt globalOption, value string, argIndex int) int {
	if !s.validateValue(opt, value) {
		return argIndex
	}

	if opt.name == profileOptionName {
		s.setProfile(value)
	}

	return argIndex
}

func (s *parseState) applySeparateValue(args []string, argIndex int, opt globalOption) int {
	if argIndex+1 >= len(args) {
		s.failMissingValue(opt)

		return argIndex
	}

	next := args[argIndex+1]
	if isOptionToken(next) {
		s.failMissingValue(opt)

		return argIndex
	}

	if !s.validateValue(opt, next) {
		return argIndex + 1
	}

	if opt.name == profileOptionName {
		s.setProfile(next)
	}

	return argIndex + 1
}

func (s *parseState) failMissingValue(opt globalOption) {
	if opt.name == profileOptionName {
		s.fail(reasonProfileFlagMissingValue)

		return
	}

	s.fail(reasonGlobalOptionMissingValue)
}

func (s *parseState) validateValue(opt globalOption, value string) bool {
	switch opt.kind {
	case optString:
		if opt.name == profileOptionName && value == "" {
			s.fail(reasonEmptyProfileName)

			return false
		}

		return true
	case optChoice:
		if !choiceAllowed(opt.choices, value) {
			s.fail(reasonGlobalOptionInvalidValue)

			return false
		}

		return true
	case optInt:
		if !validPythonInt(value) {
			s.fail(reasonGlobalOptionInvalidValue)

			return false
		}

		return true
	case optFlag:
		return true
	}

	return true
}

func (s *parseState) finish(positional []string) Invocation {
	if s.failed {
		return failedInvocation(s.reason)
	}

	command := ""
	operation := ""

	if len(positional) > 0 {
		command = positional[0]
	}

	if len(positional) > 1 {
		operation = positional[1]
	}

	return successInvocation(command, operation, s.profile, s.hasProfile)
}

func isOptionToken(arg string) bool {
	if len(arg) <= 1 || arg[0] != '-' {
		return false
	}

	if negativeNumberPattern.MatchString(arg) {
		return false
	}

	return true
}

func validPythonInt(value string) bool {
	if value == "" {
		return false
	}

	_, err := strconv.ParseInt(value, 10, 64)

	return err == nil
}
