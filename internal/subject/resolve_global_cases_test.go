package subject_test

import "aws-cli-guard/internal/subject"

func globalOptionProfileBasicsCases() []resolveCase {
	return []resolveCase{
		{
			name: "profile abbreviation",
			args: []string{awsBin, "--prof", profileNameProd, "s3", "ls"},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "double dash ends option recognition",
			args: []string{awsBin, "s3", "ls", "--", profileOpt, profileNameProd},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameDefault, Reason: ""},
		},
		{
			name: "profile name negative integer",
			args: []string{awsBin, profileOpt, "-1", "s3", "ls"},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: "-1", Reason: ""},
		},
		{
			name: "profile equals looks like option",
			args: []string{awsBin, "--profile=" + regionOpt, "s3", "ls"},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: regionOpt, Reason: ""},
		},
	}
}

func globalOptionProfileWithGlobalsCases() []resolveCase {
	return append(globalOptionProfileOutputCases(), globalOptionProfileMiscGlobalCases()...)
}

func globalOptionProfileOutputCases() []resolveCase {
	return []resolveCase{
		{
			name: "output before profile",
			args: []string{awsBin, outputOpt, outputJSON, profileOpt, profileNameProd, "s3", "ls"},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "output equals before profile",
			args: []string{
				awsBin,
				outputOpt + "=" + outputJSON,
				profileOpt,
				profileNameProd,
				"s3",
				"ls",
			},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "yaml stream output before profile",
			args: []string{
				awsBin,
				outputOpt,
				"yaml-stream",
				profileOpt,
				profileNameProd,
				"s3",
				"ls",
			},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
	}
}

func globalOptionProfileMiscGlobalCases() []resolveCase {
	return []resolveCase{
		{
			name: "region before service and profile",
			args: []string{awsBin, regionOpt, "us-east-1", "s3", "ls", profileOpt, profileNameProd},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "debug before profile",
			args: []string{awsBin, "--debug", profileOpt, profileNameProd, "s3", "ls"},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "color before profile",
			args: []string{awsBin, "--color", "on", profileOpt, profileNameProd, "s3", "ls"},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
		{
			name: "endpoint url before profile",
			args: []string{
				awsBin,
				"--endpoint-url",
				"https://example.com",
				profileOpt,
				profileNameProd,
				"s3",
				"ls",
			},
			env:  nil,
			want: subject.Subject{Kind: subject.NamedProfile, Name: profileNameProd, Reason: ""},
		},
	}
}

func globalOptionFailurePrimaryCases() []resolveCase {
	return []resolveCase{
		{
			name: "output consumes profile token",
			args: []string{awsBin, outputOpt, profileOpt, "json", "s3", "ls"},
			env:  nil,
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: reasonGlobalOptionMissingValue,
			},
		},
		{
			name: "query missing value before profile",
			args: []string{awsBin, "--query", profileOpt, "s3", "ls"},
			env:  nil,
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: reasonGlobalOptionMissingValue,
			},
		},
		{
			name: "color missing value before profile",
			args: []string{awsBin, "--color", profileOpt, profileNameProd, "s3", "ls"},
			env:  nil,
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: reasonGlobalOptionMissingValue,
			},
		},
		{
			name: "ambiguous global option",
			args: []string{awsBin, "--cli", "s3", "ls"},
			env:  nil,
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: reasonAmbiguousGlobalOption,
			},
		},
	}
}

func globalOptionFailureSecondaryCases() []resolveCase {
	return []resolveCase{
		{
			name: "flag with unexpected value",
			args: []string{awsBin, "--debug=true", "s3", "ls"},
			env:  nil,
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: reasonGlobalOptionUnexpected,
			},
		},
		{
			name: "invalid output choice",
			args: []string{awsBin, outputOpt, "JSON", "s3", "ls"},
			env:  nil,
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: reasonGlobalOptionInvalidValue,
			},
		},
		{
			name: "missing global option beats environment keys",
			args: []string{awsBin, outputOpt, profileOpt, "json", "s3", "ls"},
			env: []string{
				"AWS_ACCESS_KEY_ID=" + fixtureAccessKey,
				"AWS_SECRET_ACCESS_KEY=" + fixtureSecretKey,
			},
			want: subject.Subject{
				Kind:   subject.Unresolved,
				Name:   "",
				Reason: reasonGlobalOptionMissingValue,
			},
		},
		{
			name: "valid profile loses to environment keys",
			args: []string{awsBin, profileOpt, profileNameProd, "s3", "ls"},
			env: []string{
				"AWS_ACCESS_KEY_ID=" + fixtureAccessKey,
				"AWS_SECRET_ACCESS_KEY=" + fixtureSecretKey,
			},
			want: subject.Subject{Kind: subject.Environment, Name: "", Reason: ""},
		},
	}
}

func globalOptionProfileCases() []resolveCase {
	cases := globalOptionProfileBasicsCases()
	cases = append(cases, globalOptionProfileWithGlobalsCases()...)

	return cases
}

func globalOptionResolveCases() []resolveCase {
	cases := globalOptionProfileCases()
	cases = append(cases, globalOptionFailurePrimaryCases()...)
	cases = append(cases, globalOptionFailureSecondaryCases()...)

	return cases
}
