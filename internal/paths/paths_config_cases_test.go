package paths_test

type configFileCase struct {
	name       string
	env        []string
	want       string
	wantErr    bool
	wantSuffix string
}

func configFileDefaultCases() []configFileCase {
	return []configFileCase{
		{
			name:       "xdg appends aws-cli-guard",
			env:        []string{envXDGConfig, envHomeUser},
			want:       rootJoin("xdg", "config", "aws-cli-guard", "config.toml"),
			wantErr:    false,
			wantSuffix: "",
		},
		{
			name:       "home default",
			env:        []string{envHomeUser},
			want:       rootJoin("home", "user", ".config", "aws-cli-guard", "config.toml"),
			wantErr:    false,
			wantSuffix: "",
		},
		{
			name:       "home unset without xdg or override",
			env:        []string{},
			want:       "",
			wantErr:    true,
			wantSuffix: "",
		},
		{
			name:       "home unset xdg ok",
			env:        []string{envXDGConfig},
			want:       rootJoin("xdg", "config", "aws-cli-guard", "config.toml"),
			wantErr:    false,
			wantSuffix: "",
		},
	}
}

func configFileOverrideCases() []configFileCase {
	return []configFileCase{
		{
			name: "override replaces directory",
			env: []string{
				envConfigDir,
				"HOME=/should/not/use",
			},
			want:       rootJoin("custom", "config", "config.toml"),
			wantErr:    false,
			wantSuffix: "",
		},
		{
			name: "empty override falls through to xdg",
			env: []string{
				envConfigEmpty,
				envXDGConfig,
				envHomeUser,
			},
			want:       rootJoin("xdg", "config", "aws-cli-guard", "config.toml"),
			wantErr:    false,
			wantSuffix: "",
		},
		{
			name: "override beats xdg",
			env: []string{
				envConfigOvr,
				envXDGConfig,
				envHomeUser,
			},
			want:       rootJoin("override", "config.toml"),
			wantErr:    false,
			wantSuffix: "",
		},
		{
			name:       "relative override is absolute",
			env:        []string{envConfigRel, envHomeUser},
			want:       "",
			wantErr:    false,
			wantSuffix: "rel/config/config.toml",
		},
	}
}

func configFileCases() []configFileCase {
	return append(configFileDefaultCases(), configFileOverrideCases()...)
}
