package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/low4ey/reposcanner/pkg/resolver"
)

func main() {
	repoUrl := flag.String("repoUrl", "github.com/etcd-io/etcd", "Repository URL")
	version := flag.String("version", "v3.6.0-rc.0", "Version")
	flag.Parse()
	if *repoUrl == "" {
		fmt.Println("Please provide a repository URL")
		return
	}
	if *version == "" {
		fmt.Println("Please provide a version")
		return
	}

	// Process dependencies
	dependencyResolver := resolver.NewResolver(*repoUrl, *version)
	root := dependencyResolver.Resolver()

	// Generate JSON output
	jsonData, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("dependencies.json", jsonData, 0644); err != nil {
		panic(err)
	}

	fmt.Println("Dependency tree generated as dependencies.json")

}
