package invocation

const (
	commandSSO     = "sso"
	operationLogin = "login"
)

// IsLogin reports whether inv is a recognized login command for this build.
func (inv Invocation) IsLogin() bool {
	return inv.OK && inv.Command == commandSSO && inv.Operation == operationLogin
}
