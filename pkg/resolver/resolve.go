package resolver

import (
	"fmt"
	"os"
	"strings"

	"github.com/low4ey/reposcanner/internal/detector"
	"github.com/low4ey/reposcanner/pkg/models"
	"github.com/low4ey/reposcanner/utils"
	"golang.org/x/mod/modfile"
)

// Resolver resolves a module's dependencies by reading its go.mod file.
type Resolver struct {
	detector detector.Detector
	Cache    map[string]*models.Artifact
	Visited  map[string]bool
	repoUrl  string
	version  string
}

// NewResolver creates a new Resolver instance with initial repo URL and version.
func NewResolver(repoUrl, version string) *Resolver {
	return &Resolver{
		detector: *detector.NewDetector(repoUrl, version),
		Cache:    make(map[string]*models.Artifact),
		Visited:  make(map[string]bool),
		repoUrl:  repoUrl,
		version:  version,
	}
}

// Resolver resolves the dependencies for the current repoUrl and version.
func (r *Resolver) Resolver() *models.Artifact {
	// Work on local copies so we don't mutate r.repoUrl and r.version.
	currentRepo := r.repoUrl
	currentVersion := r.version

	// Check if we've already visited this repo in the current chain.
	if r.Visited[currentRepo] {
		return &models.Artifact{
			Name:    currentRepo,
			Version: currentVersion,
		}
	}
	r.Visited[currentRepo] = true
	defer delete(r.Visited, currentRepo)

	// Use the detector to fetch package metadata (which downloads go.mod).
	dependency := r.detector.DetectPackage(currentRepo)
	if dependency == nil {
		// Fallback to currentRepo if we cannot detect the package.
		return &models.Artifact{
			Name:    currentRepo,
			Version: currentVersion,
		}
	}

	// Fetch the dependency's go.mod file.
	err := dependency.FetchDependecy()
	if err != nil {
		fmt.Printf("failed to fetch dependency: %v\n", err)
		return &models.Artifact{
			Name:    dependency.GetRepoURL(),
			Version: dependency.GetVersion(),
		}
	}

	// Read the downloaded go.mod file.
	modData, err := os.ReadFile(dependency.GetModeFile())
	if err != nil {
		fmt.Printf("failed to read go.mod: %v\n", err)
		return &models.Artifact{
			Name:    dependency.GetRepoURL(),
			Version: dependency.GetVersion(),
		}
	}
	// If modData is empty, fallback to using dependency or current repo URL.
	if len(modData) == 0 {
		fallback := dependency.GetRepoURL()
		if fallback == "" {
			fallback = currentRepo
		}
		return &models.Artifact{
			Name:    fallback,
			Version: dependency.GetVersion(),
		}
	}

	// Parse go.mod.
	f, err := modfile.Parse("mod.go", modData, nil)
	if err != nil {
		fmt.Printf("failed to parse go.mod: %v\n", err)
		return &models.Artifact{
			Name:    dependency.GetRepoURL(),
			Version: dependency.GetVersion(),
		}
	}
	// If the module declaration is missing, fallback.
	if f.Module == nil || f.Module.Mod.Path == "" {
		fallback := dependency.GetRepoURL()
		if fallback == "" {
			fallback = currentRepo
		}
		return &models.Artifact{
			Name:    fallback,
			Version: dependency.GetVersion(),
		}
	}

	// Create the root artifact. Ensure the name is not empty.
	rootName := f.Module.Mod.Path
	if rootName == "" {
		rootName = currentRepo
	}
	root := &models.Artifact{
		Name:    rootName,
		Version: dependency.GetVersion(),
	}

	// Process each dependency requirement.
	for _, req := range f.Require {
		depPath := req.Mod.Path
		depVersion := req.Mod.Version

		// Handle replacement if specified.
		for _, rep := range f.Replace {
			if rep.Old.Path == req.Mod.Path {
				depPath = rep.New.Path
				if rep.New.Version != "" {
					depVersion = rep.New.Version
				} else {
					depVersion = dependency.GetVersion()
				}
				break
			}
		}

		// Use canonical key for caching.
		key := depPath + "@" + depVersion
		var depArtifact *models.Artifact
		if cached, exists := r.Cache[key]; exists {
			depArtifact = cached
		} else {
			// Avoid self-dependency.
			if f.Module != nil && f.Module.Mod.Path == depPath {
				depArtifact = &models.Artifact{
					Name:    depPath,
					Version: depVersion,
				}
			} else if strings.Contains(depPath, "github.com") {
				// Create a new resolver instance for GitHub dependency.
				newResolver := NewResolver(depPath, depVersion)
				// Share Cache and Visited maps.
				newResolver.Cache = r.Cache
				newResolver.Visited = r.Visited
				depArtifact = newResolver.Resolver()
			} else if depPath != "" {
				resolvedUrl, reSolvedVersion := utils.ResolveUrl(depPath)
				newResolver := NewResolver(resolvedUrl, reSolvedVersion)
				newResolver.Cache = r.Cache
				newResolver.Visited = r.Visited
				depArtifact = newResolver.Resolver()
			} else {
				depArtifact = &models.Artifact{
					Name:    depPath,
					Version: depVersion,
				}
			}
			// Fallback: if the dependency artifact name is empty, use depPath.
			if depArtifact.Name == "" {
				depArtifact.Name = depPath
			}
			r.Cache[key] = depArtifact
		}
		root.Dependencies = append(root.Dependencies, depArtifact)
	}

	return root
}
