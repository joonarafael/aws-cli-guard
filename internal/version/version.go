package version

// Version is the release stamp. Untagged builds use "dev". Release builds set it with:
// -ldflags '-X aws-cli-guard/internal/version.Version=X.Y.Z'.
//
//nolint:gochecknoglobals // release stamp variable for -ldflags.
var Version = "dev"
