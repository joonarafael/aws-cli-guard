package policy_test

import "aws-cli-guard/internal/subject"

func namedProfile(name string) subject.Subject {
	return subject.Subject{Kind: subject.NamedProfile, Name: name, Reason: ""}
}

func envSubject(name string) subject.Subject {
	return subject.Subject{Kind: subject.Environment, Name: name, Reason: ""}
}

func unresolvedSubject(name, reason string) subject.Subject {
	return subject.Subject{Kind: subject.Unresolved, Name: name, Reason: reason}
}
