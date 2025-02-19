package models

type Artifact struct {
	Name         string      `json:"name"`
	Version      string      `json:"version"`
	Dependencies []*Artifact `json:"dependencies"`
}
