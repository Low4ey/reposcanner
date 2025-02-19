package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/low4ey/reposcanner/pkg/models"
)

// ExtractRepoName extracts the repository name from a GitHub URL
func ExtractRepoName(url string) string {
	// Trim any trailing slash to avoid an empty string as the last element.
	url = strings.TrimSuffix(url, "/")
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func ResolveUrl(url string) (string, string) {
	proxyURL := fmt.Sprintf("https://proxy.golang.org/%s/@latest", url)

	// Execute curl command to fetch module metadata from the Go proxy.
	cmd := exec.Command("curl", "-sL", proxyURL)
	output, err := cmd.Output()
	if err != nil {
		return "", ""
	}

	// Parse JSON response.
	var moduleInfo models.ModuleInfo
	if err := json.Unmarshal(output, &moduleInfo); err != nil {
		return "", ""
	}

	// Remove the "https://" or "http://" prefix from the origin URL.
	resolvedURL := moduleInfo.Origin.URL
	if strings.HasPrefix(resolvedURL, "https://") {
		resolvedURL = strings.TrimPrefix(resolvedURL, "https://")
	} else if strings.HasPrefix(resolvedURL, "http://") {
		resolvedURL = strings.TrimPrefix(resolvedURL, "http://")
	}

	return resolvedURL, moduleInfo.Version
}
