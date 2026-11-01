package invocation_test

import (
	"aws-cli-guard/internal/invocation"
	"testing"
)

func TestInvocationIsLogin(t *testing.T) {
	t.Parallel()

	for _, testCase := range isLoginCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := invocation.Parse(testCase.args)
			if got.IsLogin() != testCase.wantLogin {
				t.Fatalf("IsLogin() = %v, want %v (OK=%v cmd=%q op=%q)",
					got.IsLogin(), testCase.wantLogin, got.OK, got.Command, got.Operation)
			}
		})
	}
}

type isLoginCase struct {
	name      string
	args      []string
	wantLogin bool
}

func isLoginCases() []isLoginCase {
	return []isLoginCase{
		{
			name:      "sso login",
			args:      []string{awsBin, ssoCmd, loginOp},
			wantLogin: true,
		},
		{
			name:      "sso login with sso-session",
			args:      []string{awsBin, ssoCmd, loginOp, "--sso-session", "dev"},
			wantLogin: true,
		},
		{
			name:      "profile before sso login",
			args:      []string{awsBin, profileOpt, "default", ssoCmd, loginOp},
			wantLogin: true,
		},
		{
			name:      "profile between sso and login",
			args:      []string{awsBin, ssoCmd, profileOpt, profileNameProd, loginOp},
			wantLogin: true,
		},
		{
			name:      "top-level login",
			args:      []string{awsBin, loginOp},
			wantLogin: false,
		},
		{
			name:      "sso logout",
			args:      []string{awsBin, ssoCmd, logoutOp},
			wantLogin: false,
		},
		{
			name:      "wrong case SSO",
			args:      []string{awsBin, "SSO", loginOp},
			wantLogin: false,
		},
		{
			name:      "empty profile fails parse",
			args:      []string{awsBin, "--profile=", ssoCmd, loginOp},
			wantLogin: false,
		},
	}
}
