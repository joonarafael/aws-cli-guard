package agent

import "aws-cli-guard/internal/environ"

const recognitionAgent recognition = 1

// builtinMarkerNames lists command-scoped environment variable names for built-in
// agent integrations. Production rows are recorded with the product and pinned
// version in ADR 0022. Test builds may append one extra name.
func builtinMarkerNames() []string {
	prod := productionBuiltinMarkerNames()
	extra := extraBuiltinMarkerNames()
	names := make([]string, 0, len(prod)+len(extra))
	names = append(names, prod...)
	names = append(names, extra...)

	return names
}

// ApplyMarkers updates ctx when enabled and a built-in marker name is set to a
// non-empty value in env. It never reads os.Environ.
func ApplyMarkers(ctx Context, env []string, enabled bool) Context {
	if !enabled {
		return ctx
	}

	if matchMarkers(env, builtinMarkerNames()) {
		ctx.rec = recognitionAgent
	}

	return ctx
}

func matchMarkers(env []string, names []string) bool {
	for _, name := range names {
		value, ok := environ.Lookup(name, env)
		if ok && value != "" {
			return true
		}
	}

	return false
}
