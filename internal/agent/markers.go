package agent

import "aws-cli-guard/internal/environ"

const recognitionAgent recognition = 1

// builtinMarkerNames lists command-scoped environment variable names for built-in
// agent integrations. Add a name only with the product and pinned version that sets it.
func builtinMarkerNames() []string {
	return extraBuiltinMarkerNames()
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
