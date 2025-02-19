
# Reposcanner

**Reposcanner** is a Go module dependency resolver designed to analyze and build a dependency tree for Go projects. It reads `go.mod` files from repositories (supporting both GitHub and Googlesource), recursively resolves dependencies, and caches results to avoid duplicate work and circular dependency issues.

## Overview

The core functionality of Reposcanner is implemented in the `resolver` package. It uses a `Detector` (from the internal `detector` package) to fetch a repository’s metadata and download its `go.mod` file. The resolver then parses the file, extracts dependencies, and recursively resolves each dependency. Key features include:

* **Recursive Dependency Resolution:** Traverses the dependency tree by reading and parsing each module’s `go.mod` file.
* **Caching:** Uses an in-memory cache to avoid resolving the same module multiple times.
* **Cycle Prevention:** Implements a `Visited` map (or, optionally, in-place backtracking) to detect and break circular dependencies.
* **Multi-Source Support:** Supports repositories from GitHub (using raw.githubusercontent.com) and Googlesource (with base64 decoding).

## Design Decisions

### Resolver State Management

Two approaches were considered for managing state during recursive dependency resolution:

* **New Resolver Instances:** Each recursive call creates a new `Resolver` instance with its own state (repo URL and version). These instances share common maps (`Cache` and `Visited`) to avoid duplicate work. This method isolates state between recursive calls and simplifies reasoning.
* **In-Place Backtracking:** Alternatively, the resolver’s state can be updated in place (i.e., saving the current state before a recursive call and restoring it afterward). This method minimizes memory usage but adds complexity in ensuring the state is properly restored in every execution path.

In our implementation, we use new resolver instances for clarity and simplicity, as the memory cost is minimal.

### Dependency Resolution

* **Fetching go.mod:** The resolver uses a detector to download the `go.mod` file. For GitHub, it constructs a raw file URL and downloads it directly. For Googlesource, it retrieves a base64-encoded file, decodes it, and writes it locally.
* **Parsing and Fallback:** The downloaded `go.mod` is parsed using Go’s `modfile` package. If parsing fails or the module declaration is missing, a basic artifact is created using fallback values (e.g., the repository URL).
* **Cache and Visited:** Resolved modules are stored in a cache using a key in the format `depPath@depVersion`. The `Visited` map helps prevent infinite recursion when a module depends on itself (or circular dependencies exist).

## Usage

To run the dependency resolver:

1. Set up your project’s entry point (e.g., in `main.go`) to create a new resolver with the target repository and version.
2. Call the resolver’s `Resolver()` method to get the root artifact, which contains the dependency tree.
3. Optionally, output the dependency tree as JSON for further analysis.

Example:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/low4ey/reposcanner/pkg/resolver"
)

func main() {
	// Example repository and version.
	repoURL := "github.com/etcd-io/etcd"
	version := "v3.6.0-rc.0"

	// Create a new Resolver.
	res := resolver.NewResolver(repoURL, version)
	artifact := res.Resolver()

	// Marshal the dependency tree to JSON.
	jsonData, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		panic(err)
	}

	// Write the JSON output to a file.
	if err := os.WriteFile("dependencies.json", jsonData, 0644); err != nil {
		panic(err)
	}

	fmt.Println("Dependency tree generated as dependencies.json")
}
```

## Project Structure

* **`pkg/resolver/resolve.go`**

  Implements the recursive dependency resolution logic. Handles fetching, parsing, caching, and cycle prevention.
* **`internal/detector/`**

  Contains logic for detecting and fetching the `go.mod` file from the given repository.
* **`pkg/models/`**

  Defines the data structures, such as the `Artifact` that represents a module and its dependencies.
* **`utils/`**

  Contains utility functions such as `ResolveUrl` (which maps module paths to repository URLs) and URL status checks.

## Notes

* **Error Handling & Fallbacks:**

  The resolver is designed to gracefully fall back to basic artifacts if fetching or parsing fails, ensuring that the dependency tree is as complete as possible.
* **State Management Alternatives:**

  While new resolver instances are used for clarity, an alternative implementation could update state in place (backtracking) to potentially reduce memory usage. However, the trade-off in code complexity is usually not justified given the minimal cost of small resolver instances.
