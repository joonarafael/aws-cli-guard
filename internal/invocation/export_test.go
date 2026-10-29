package invocation_test

import (
	"aws-cli-guard/internal/invocation"
	"testing"
)

func TestParsePrintsAccessKeys(t *testing.T) {
	t.Parallel()

	for _, testCase := range printsAccessKeysCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := invocation.Parse(testCase.args)
			if got.OK != testCase.wantOK {
				t.Fatalf("Parse() OK = %v, want %v", got.OK, testCase.wantOK)
			}

			if got.PrintsAccessKeys != testCase.wantPrints {
				t.Fatalf(
					"PrintsAccessKeys = %v, want %v",
					got.PrintsAccessKeys,
					testCase.wantPrints,
				)
			}
		})
	}
}

type printsAccessKeysCase struct {
	name       string
	args       []string
	wantOK     bool
	wantPrints bool
}

func printsAccessKeysCases() []printsAccessKeysCase {
	return append(
		append(exportCredentialsPrintsCases(), configureGetSecretPrintsCases()...),
		append(
			append(
				append(configureGetNonSecretPrintsCases(), configureGetAmbiguousPrintsCases()...),
				configureGetEdgePrintsCases()...,
			),
			otherInvocationPrintsCases()...,
		)...,
	)
}

func exportCredentialsPrintsCases() []printsAccessKeysCase {
	return []printsAccessKeysCase{
		{
			name:       "export credentials",
			args:       []string{awsBin, configureCmd, operationExportCreds},
			wantOK:     true,
			wantPrints: true,
		},
		{
			name: "export credentials with format flag",
			args: []string{
				awsBin,
				configureCmd,
				operationExportCreds,
				"--format",
				"json",
			},
			wantOK:     true,
			wantPrints: true,
		},
		{
			name: "profile before export credentials",
			args: []string{
				awsBin,
				profileOpt,
				profileNameProd,
				configureCmd,
				operationExportCreds,
			},
			wantOK:     true,
			wantPrints: true,
		},
	}
}

func configureGetSecretPrintsCases() []printsAccessKeysCase {
	return []printsAccessKeysCase{
		{
			name:       "get access key id",
			args:       []string{awsBin, configureCmd, operationGet, settingAccessKeyID},
			wantOK:     true,
			wantPrints: true,
		},
		{
			name:       "get secret access key",
			args:       []string{awsBin, configureCmd, operationGet, settingSecretKey},
			wantOK:     true,
			wantPrints: true,
		},
		{
			name:       "get session token",
			args:       []string{awsBin, configureCmd, operationGet, settingSessionToken},
			wantOK:     true,
			wantPrints: true,
		},
		{
			name: "get prefixed secret",
			args: []string{
				awsBin,
				configureCmd,
				operationGet,
				"prod." + settingAccessKeyID,
			},
			wantOK:     true,
			wantPrints: true,
		},
	}
}

func configureGetNonSecretPrintsCases() []printsAccessKeysCase {
	return []printsAccessKeysCase{
		{
			name:       "get region",
			args:       []string{awsBin, configureCmd, operationGet, settingRegion},
			wantOK:     true,
			wantPrints: false,
		},
		{
			name:       "get prefixed non secret",
			args:       []string{awsBin, configureCmd, operationGet, "prod." + settingRegion},
			wantOK:     true,
			wantPrints: false,
		},
		{
			name: "get nested non secret",
			args: []string{
				awsBin,
				configureCmd,
				operationGet,
				"prod.s3.max_concurrent_requests",
			},
			wantOK:     true,
			wantPrints: false,
		},
		{
			name:       "get wrong case secret name",
			args:       []string{awsBin, configureCmd, operationGet, "AWS_ACCESS_KEY_ID"},
			wantOK:     true,
			wantPrints: false,
		},
	}
}

func configureGetAmbiguousPrintsCases() []printsAccessKeysCase {
	return []printsAccessKeysCase{
		{
			name:       "get missing setting",
			args:       []string{awsBin, configureCmd, operationGet},
			wantOK:     true,
			wantPrints: true,
		},
		{
			name: "get two trailing positionals",
			args: []string{
				awsBin,
				configureCmd,
				operationGet,
				settingRegion,
				"extra",
			},
			wantOK:     true,
			wantPrints: true,
		},
		{
			name: "double dash before configure still classifies get",
			args: []string{
				awsBin,
				"--",
				configureCmd,
				operationGet,
				settingAccessKeyID,
			},
			wantOK:     true,
			wantPrints: true,
		},
		{
			name: "get with double dash before setting is ambiguous",
			args: []string{
				awsBin,
				configureCmd,
				operationGet,
				"--",
				settingAccessKeyID,
			},
			wantOK:     true,
			wantPrints: true,
		},
	}
}

func configureGetEdgePrintsCases() []printsAccessKeysCase {
	return []printsAccessKeysCase{
		{
			name: "profile after get secret",
			args: []string{
				awsBin,
				configureCmd,
				operationGet,
				settingAccessKeyID,
				profileOpt,
				profileNameProd,
			},
			wantOK:     true,
			wantPrints: true,
		},
		{
			name: "failed global option before secret get",
			args: []string{
				awsBin,
				"--cli-read-timeout=1.5",
				configureCmd,
				operationGet,
				settingAccessKeyID,
			},
			wantOK:     false,
			wantPrints: false,
		},
	}
}

func otherInvocationPrintsCases() []printsAccessKeysCase {
	return []printsAccessKeysCase{
		{
			name:       "list profiles",
			args:       []string{awsBin, configureCmd, "list-profiles"},
			wantOK:     true,
			wantPrints: false,
		},
		{
			name: "configure set",
			args: []string{
				awsBin,
				configureCmd,
				"set",
				settingRegion,
				"us-east-1",
			},
			wantOK:     true,
			wantPrints: false,
		},
		{
			name:       "sso login",
			args:       []string{awsBin, "sso", "login"},
			wantOK:     true,
			wantPrints: false,
		},
		{
			name:       "ec2 describe",
			args:       []string{awsBin, serviceEC2, opDescribe},
			wantOK:     true,
			wantPrints: false,
		},
	}
}
