package subject

// Kind identifies how credentials are supplied for an invocation.
type Kind int

const (
	// NamedProfile is a conventional profile name, including the literal name "default".
	NamedProfile Kind = iota + 1
	// Environment is static credentials from AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.
	Environment
	// Unresolved is a credential source the wrapper cannot classify safely.
	Unresolved
)

// Subject is the effective credential subject for policy evaluation.
type Subject struct {
	Kind   Kind
	Name   string
	Reason string
}
