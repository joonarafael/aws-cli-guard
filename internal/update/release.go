package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	githubLatestURL    = "https://api.github.com/repos/joonarafael/aws-cli-guard/releases/latest"
	githubDownloadBase = "https://github.com/joonarafael/aws-cli-guard/releases/download"
	requestTimeout     = 30 * time.Second
	maxReleaseBody     = 1 << 20
	githubAPIAccept    = "application/vnd.github+json"
	githubAPIVersion   = "2022-11-28"
)

// ReleaseClient fetches published release metadata and assets.
type ReleaseClient struct {
	HTTP         *http.Client
	LatestURL    string
	DownloadBase string
}

type releaseAPIResponse struct {
	TagName    string `json:"tag_name"` //nolint:tagliatelle // GitHub API wire name.
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// FetchLatest returns X.Y.Z when a usable latest release exists.
//
//nolint:wrapcheck // HTTP and JSON errors are handled by the helper as check failures.
func FetchLatest(client ReleaseClient, stamp string) (string, bool, error) {
	httpClient := client.HTTP
	if httpClient == nil {
		//nolint:exhaustruct_v5 // default HTTP client.
		httpClient = &http.Client{Timeout: requestTimeout}
	}

	latestURL := client.LatestURL
	if latestURL == "" {
		latestURL = githubLatestURL
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, latestURL, nil)
	if err != nil {
		return "", false, err
	}

	req.Header.Set("User-Agent", "aws-cli-guard/"+stamp)
	req.Header.Set("Accept", githubAPIAccept)
	req.Header.Set("X-Github-Api-Version", githubAPIVersion)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", false, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("%w: status %d", errReleaseHTTP, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxReleaseBody))
	if err != nil {
		return "", false, err
	}

	var doc releaseAPIResponse

	err = json.Unmarshal(body, &doc)
	if err != nil {
		return "", false, err
	}

	if doc.Draft || doc.Prerelease {
		return "", false, nil
	}

	version, ok := parseReleaseTag(doc.TagName)
	if !ok {
		return "", false, nil
	}

	return version, true, nil
}

func parseReleaseTag(tag string) (string, bool) {
	after, ok := strings.CutPrefix(tag, "v")
	if !ok || !ReleaseStamp(after) {
		return "", false
	}

	return after, true
}

func releaseDownloadBase(client ReleaseClient, version string) string {
	base := client.DownloadBase
	if base == "" {
		base = githubDownloadBase
	}

	return base + "/v" + version
}

func archiveName(version, goos, goarch string) string {
	return fmt.Sprintf("aws-cli-guard_%s_%s_%s.tar.gz", version, goos, goarch)
}
