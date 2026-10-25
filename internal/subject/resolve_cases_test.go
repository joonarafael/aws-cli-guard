package subject_test

import "aws-cli-guard/internal/subject"

type resolveCase struct {
	name string
	args []string
	env  []string
	want subject.Subject
}

func profileFlagValidCases() []resolveCase {
	return []resolveCase{
		{
			name: "profile flag before service",
			args: []string{awsBin, profileOpt, profileNameProd, "ec2", "describe-instances"},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "profile equals form",
			args: []string{awsBin, "--profile=" + profileNameProd, "s3", "ls"},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "profile flag after service",
			args: []string{awsBin, "ec2", "describe-instances", profileOpt, profileNameProd},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "argv0 ignored for flags",
			args: []string{"./bin/aws", profileOpt, profileNameProd, "sts", "get-caller-identity"},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "repeated identical profile flags",
			args: []string{
				awsBin,
				profileOpt,
				profileNameProd,
				"--profile=" + profileNameProd,
				"s3",
				"ls",
			},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
	}
}

func profileFlagErrorCases() []resolveCase {
	return []resolveCase{
		{
			name: "conflicting profile flags",
			args: []string{awsBin, profileOpt, "a", profileOpt, "b", "s3", "ls"},
			env:  nil,
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: "conflicting-profiles",
			},
		},
		{
			name: "dangling profile flag",
			args: []string{awsBin, "s3", "ls", profileOpt},
			env:  nil,
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: "profile-flag-missing-value",
			},
		},
		{
			name: "empty profile equals",
			args: []string{awsBin, "--profile=", "s3", "ls"},
			env:  nil,
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: "empty-profile-name",
			},
		},
		{
			name: "profile flag value looks like option",
			args: []string{awsBin, profileOpt, "--region", "us-east-1", "s3", "ls"},
			env:  nil,
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: "profile-flag-missing-value",
			},
		},
	}
}

func profileFlagResolveCases() []resolveCase {
	return append(profileFlagValidCases(), profileFlagErrorCases()...)
}

func profileEnvResolveCases() []resolveCase {
	return []resolveCase{
		{
			name: "aws profile env",
			args: []string{awsBin, "s3", "ls"},
			env:  []string{"AWS_PROFILE=staging"},
			want: subject.Subject{Kind: subject.NamedProfile, Name: "staging", Reason: ""},
		},
		{
			name: "flag beats aws profile env",
			args: []string{awsBin, profileOpt, "cli", "s3", "ls"},
			env:  []string{"AWS_PROFILE=env"},
			want: subject.Subject{Kind: subject.NamedProfile, Name: "cli", Reason: ""},
		},
		{
			name: "implicit default profile",
			args: []string{awsBin, "s3", "ls"},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameDefault, Reason: ""},
		},
		{
			name: "default profile name is normal",
			args: []string{awsBin, profileOpt, profileNameDefault, "s3", "ls"},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameDefault, Reason: ""},
		},
	}
}

func profileResolveCases() []resolveCase {
	return append(profileFlagResolveCases(), profileEnvResolveCases()...)
}

func credentialResolveCases() []resolveCase {
	return []resolveCase{
		{
			name: "environment credentials beat named profile",
			args: []string{awsBin, profileOpt, profileNameProd, "s3", "ls"},
			env: []string{
				"AWS_ACCESS_KEY_ID=" + fixtureAccessKey,
				"AWS_SECRET_ACCESS_KEY=" + fixtureSecretKey,
			},
			want: subject.Subject{Kind: subject.Environment, Name: "", Reason: ""},
		},
		{
			name: "secret only leaves named profile",
			args: []string{awsBin, profileOpt, profileNameProd, "s3", "ls"},
			env:  []string{"AWS_SECRET_ACCESS_KEY=" + fixtureSecretKey},
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "session token only leaves named profile",
			args: []string{awsBin, profileOpt, profileNameProd, "s3", "ls"},
			env:  []string{"AWS_SESSION_TOKEN=" + fixtureSession},
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "partial access key without secret",
			args: []string{awsBin, "s3", "ls"},
			env:  []string{"AWS_ACCESS_KEY_ID=" + fixtureAccessKey},
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: "partial-env-credentials",
			},
		},
		{
			name: "broken flag beats environment keys",
			args: []string{awsBin, profileOpt, "a", profileOpt, "b", "s3", "ls"},
			env: []string{
				"AWS_ACCESS_KEY_ID=" + fixtureAccessKey,
				"AWS_SECRET_ACCESS_KEY=" + fixtureSecretKey,
			},
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: "conflicting-profiles",
			},
		},
		{
			name: "env first wins for access key",
			args: []string{awsBin, "s3", "ls"},
			env: []string{
				"AWS_ACCESS_KEY_ID=first",
				"AWS_ACCESS_KEY_ID=" + fixtureAccessKey,
				"AWS_SECRET_ACCESS_KEY=" + fixtureSecretKey,
			},
			want: subject.Subject{Kind: subject.Environment, Name: "", Reason: ""},
		},
	}
}

func ambientResolveCases() []resolveCase {
	return []resolveCase{
		{
			name: "container relative uri ambient",
			args: []string{awsBin, "s3", "ls"},
			env:  []string{"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI=/v2/creds"},
			want: subject.Subject{Kind: subject.Unresolved, Name: "", Reason: reasonAmbient},
		},
		{
			name: "container full uri ambient",
			args: []string{awsBin, "s3", "ls"},
			env:  []string{"AWS_CONTAINER_CREDENTIALS_FULL_URI=http://169.254.170.2/creds"},
			want: subject.Subject{Kind: subject.Unresolved, Name: "", Reason: reasonAmbient},
		},
		{
			name: "web identity token ambient",
			args: []string{awsBin, "s3", "ls"},
			env:  []string{"AWS_WEB_IDENTITY_TOKEN_FILE=/var/run/secrets/token"},
			want: subject.Subject{Kind: subject.Unresolved, Name: "", Reason: reasonAmbient},
		},
	}
}
