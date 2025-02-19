package github

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/low4ey/reposcanner/internal"
)

type GithubDependencyResolver struct {
	repoURL     string
	version     string
	repoModFile string
	ownerName   string
	repoName    string
}

func NewGithubDependencyResolver(repoURL, version, repoModFile string) internal.DependecyResolver {
	return &GithubDependencyResolver{
		repoURL:     repoURL,
		version:     version,
		repoModFile: repoModFile,
	}
}

func (g *GithubDependencyResolver) FetchDependecy() error {
	g.extractMetadata()
	g.convertGitHubURLToRaw()
	isReachable, err := g.ValidateUrl()
	if err != nil {
		return fmt.Errorf("failed to check URL: %v", err)
	}

	if isReachable {
		err = exec.Command("curl", "-s", "-o", g.repoModFile, g.repoURL).Run()
		if err != nil {
			return fmt.Errorf("failed to download from GitHub: %v", err)
		}
	}
	return nil
}

func (g *GithubDependencyResolver) GetModeFile() string {
	return g.repoModFile
}

func (g *GithubDependencyResolver) GetRepoURL() string {
	return g.repoURL
}

func (g *GithubDependencyResolver) GetVersion() string {
	return g.version
}

func (g *GithubDependencyResolver) ValidateUrl() (bool, error) {
	// First, try a HEAD request on the tag-based raw URL.
	resp, err := http.Head(g.repoURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		branchResp, err := http.Head(g.repoURL)
		if err != nil {
			return false, fmt.Errorf("failed to check branch URL: %w", err)
		}
		defer branchResp.Body.Close()
		return branchResp.StatusCode == http.StatusOK, nil
	}

	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}
func (g *GithubDependencyResolver) convertGitHubURLToRaw() {
	//
	const githubDomain = "github.com/"
	const rawDomain = "raw.githubusercontent.com/"
	// Replace "github.com/" with "raw.githubusercontent.com/"
	rawUrl := strings.Replace(g.repoURL, githubDomain, rawDomain, 1)
	g.repoURL = fmt.Sprintf("https://%s/refs/tags/%s/go.mod", rawUrl, g.version)
}

func (g *GithubDependencyResolver) convertBranchUrl() {
	g.repoURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/refs/heads/master/go.mod", g.ownerName, g.repoName)
}

func (g *GithubDependencyResolver) extractMetadata() {
	url := strings.TrimSuffix(g.repoURL, "/")
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return
	}
	g.repoName = parts[len(parts)-1]
	g.ownerName = parts[len(parts)-2]
}
