package gate_test

func gateSSOLoginArgs(extra ...string) []string {
	base := make([]string, 0, 3+len(extra))
	base = append(base, gateProgName, gateSSOCommand, gateLoginOperation)

	return append(base, extra...)
}
