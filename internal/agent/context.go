package agent

import "os"

// Stdin classifies standard input for caller context.
type Stdin int

const (
	// StdinUndetermined means stdin could not be classified.
	StdinUndetermined Stdin = iota
	// StdinInteractive means stdin is a terminal.
	StdinInteractive
	// StdinNonInteractive means stdin is open but not a terminal.
	StdinNonInteractive
)

type recognition int

const recognitionUnknown recognition = 0

// Context holds caller-context signals observed for one invocation.
type Context struct {
	stdin Stdin
	rec   recognition
}

// Observe classifies stdin without reading from it.
func Observe(stdin *os.File) Context {
	return ContextWithStdin(classifyStdin(stdin))
}

// ContextWithStdin builds a context with unknown agent recognition.
func ContextWithStdin(state Stdin) Context {
	return Context{stdin: state, rec: recognitionUnknown}
}

// FollowsProfilePolicy reports whether normal profile rules apply.
//
// Interactive, non-interactive, and undetermined stdin all follow profile
// policy. An out-of-range stdin class does not.
func (c Context) FollowsProfilePolicy() bool {
	switch c.stdin {
	case StdinUndetermined, StdinInteractive, StdinNonInteractive:
		return true
	default:
		return false
	}
}

// Recognized reports whether an enabled detector positively matched this caller.
func (c Context) Recognized() bool {
	return c.rec != recognitionUnknown
}
