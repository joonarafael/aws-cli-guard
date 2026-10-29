package invocation

func successInvocation(
	command, operation, profile string,
	hasProfile, printsAccessKeys bool,
) Invocation {
	return Invocation{
		Command:          command,
		Operation:        operation,
		Profile:          profile,
		HasProfile:       hasProfile,
		OK:               true,
		Reason:           "",
		PrintsAccessKeys: printsAccessKeys,
	}
}

func failedInvocation(reason string) Invocation {
	return Invocation{
		Command:          "",
		Operation:        "",
		Profile:          "",
		HasProfile:       false,
		OK:               false,
		Reason:           reason,
		PrintsAccessKeys: false,
	}
}
