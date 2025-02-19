package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/low4ey/reposcanner/internal"
	"github.com/low4ey/reposcanner/internal/github"
	"github.com/low4ey/reposcanner/internal/google"
	"github.com/low4ey/reposcanner/pkg/global"
	"github.com/low4ey/reposcanner/utils"
)

type Detector struct {
	RepoURL     string
	Version     string
	repoModFile string
}

func NewDetector(repoURL, version string) *Detector {
	// Ensure the temporary directory exists.
	repoName := utils.ExtractRepoName(repoURL)
	repoDir := filepath.Join(global.Directory, repoName)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		panic(fmt.Errorf("failed to create repo directory: %v", err))
	}

	// Prepare the local go.mod file path.
	repoModFile := filepath.Join(repoDir, "go.mod")
	return &Detector{
		RepoURL:     repoURL,
		Version:     version,
		repoModFile: repoModFile,
	}
}

func (b *Detector) DetectPackage(url string) internal.DependecyResolver {
	if strings.Contains(url, "github") {
		return github.NewGithubDependencyResolver(b.RepoURL, b.Version, b.repoModFile)
	}
	if strings.Contains(url, "googlesource") {
		return google.NewGoogleDependencyResolver(b.RepoURL, b.Version, b.repoModFile)
	}
	return nil
}
