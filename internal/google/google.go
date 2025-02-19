package google

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/low4ey/reposcanner/internal"
)

type GoogleDependencyResolver struct {
	repoURL     string
	version     string
	repoModFile string
}

func NewGoogleDependencyResolver(repoUrl, version, repoModFile string) internal.DependecyResolver {
	return &GoogleDependencyResolver{
		repoURL:     repoUrl,
		version:     version,
		repoModFile: repoModFile,
	}
}

func (g *GoogleDependencyResolver) FetchDependecy() error {
	g.convertGoogleURLToRaw()
	isReachable, err := g.ValidateUrl()
	if err != nil {
		return fmt.Errorf("failed to check URL: %v", err)
	}
	var output []byte
	if isReachable {
		output, err = exec.Command("curl", "-sL", g.repoURL).Output()
		if err != nil {
			return fmt.Errorf("error fetching file: %v", err)
		}
	}
	// Decode the base64 output.
	var decoded []byte
	decoded, err = base64.StdEncoding.DecodeString(string(output))
	if err != nil {
		return fmt.Errorf("error decoding base64: %v", err)
	}
	// Write the decoded content to repoModFile.
	err = os.WriteFile(g.repoModFile, decoded, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}
	return nil
}

func (g *GoogleDependencyResolver) GetModeFile() string {
	return g.repoModFile
}

func (g *GoogleDependencyResolver) GetRepoURL() string {
	return g.repoURL
}

func (g *GoogleDependencyResolver) GetVersion() string {
	return g.version
}

func (g *GoogleDependencyResolver) ValidateUrl() (bool, error) {
	resp, err := http.Head(g.repoURL)
	if err != nil {
		g.convertBranchUrl()
		resp, err = http.Head(g.repoURL)
		if err != nil {
			return false, fmt.Errorf("failed to check URL: %w", err)
		}
		return resp.StatusCode == http.StatusOK, nil
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

func (g *GoogleDependencyResolver) convertBranchUrl() {
	g.repoURL = fmt.Sprintf("%s/+/refs/heads/%s/go.mod", g.repoURL, "master")
}

func (g *GoogleDependencyResolver) convertGoogleURLToRaw() {
	g.repoURL = fmt.Sprintf("https://%s/+/refs/tags/%s/go.mod?format=TEXT", g.repoURL, g.version)
}
