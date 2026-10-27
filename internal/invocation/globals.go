package invocation

import "slices"

const profileOptionName = "profile"

type optionKind int

const (
	optFlag optionKind = iota
	optString
	optChoice
	optInt
)

type globalOption struct {
	name    string
	kind    optionKind
	choices []string
}

func globalOptions() []globalOption {
	return []globalOption{
		flagOption("debug"),
		stringOption("endpoint-url"),
		flagOption("no-verify-ssl"),
		flagOption("no-paginate"),
		{
			name:    "output",
			kind:    optChoice,
			choices: []string{"json", "text", "table", "yaml", "yaml-stream"},
		},
		stringOption("query"),
		stringOption(profileOptionName),
		stringOption("region"),
		flagOption("version"),
		{
			name:    "color",
			kind:    optChoice,
			choices: []string{"on", "off", "auto"},
		},
		flagOption("no-sign-request"),
		stringOption("ca-bundle"),
		intOption("cli-read-timeout"),
		intOption("cli-connect-timeout"),
		{
			name:    "cli-binary-format",
			kind:    optChoice,
			choices: []string{"base64", "raw-in-base64-out"},
		},
		flagOption("no-cli-pager"),
		flagOption("cli-auto-prompt"),
		flagOption("no-cli-auto-prompt"),
	}
}

func flagOption(name string) globalOption {
	return globalOption{name: name, kind: optFlag, choices: nil}
}

func stringOption(name string) globalOption {
	return globalOption{name: name, kind: optString, choices: nil}
}

func intOption(name string) globalOption {
	return globalOption{name: name, kind: optInt, choices: nil}
}

type nameMatch struct {
	option globalOption
	exact  bool
}

func matchGlobalOption(name string) (nameMatch, matchStatus) {
	opts := globalOptions()

	var (
		exact         *globalOption
		prefixMatches []globalOption
	)

	for _, opt := range opts {
		if opt.name == name {
			copyOpt := opt
			exact = &copyOpt

			break
		}
	}

	if exact != nil {
		return nameMatch{option: *exact, exact: true}, matchOK
	}

	for _, opt := range opts {
		if len(name) > 0 && len(name) <= len(opt.name) && opt.name[:len(name)] == name {
			prefixMatches = append(prefixMatches, opt)
		}
	}

	empty := nameMatch{option: globalOption{name: "", kind: optFlag, choices: nil}, exact: false}

	switch len(prefixMatches) {
	case 0:
		return empty, matchUnknown
	case 1:
		return nameMatch{option: prefixMatches[0], exact: false}, matchOK
	default:
		return empty, matchAmbiguous
	}
}

type matchStatus int

const (
	matchOK matchStatus = iota
	matchUnknown
	matchAmbiguous
)

func choiceAllowed(choices []string, value string) bool {
	return slices.Contains(choices, value)
}
