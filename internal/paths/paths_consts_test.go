package paths_test

const (
	envHomeUser    = "HOME=/home/user"
	envXDGConfig   = "XDG_CONFIG_HOME=/xdg/config"
	envXDGState    = "XDG_STATE_HOME=/xdg/state"
	envXDGCache    = "XDG_CACHE_HOME=/xdg/cache"
	envConfigDir   = "AWS_CLI_GUARD_CONFIG_DIR=/custom/config"
	envStateDir    = "AWS_CLI_GUARD_STATE_DIR=/state/override"
	envConfigEmpty = "AWS_CLI_GUARD_CONFIG_DIR="
	envConfigRel   = "AWS_CLI_GUARD_CONFIG_DIR=rel/config"
	envConfigOvr   = "AWS_CLI_GUARD_CONFIG_DIR=/override"
)
