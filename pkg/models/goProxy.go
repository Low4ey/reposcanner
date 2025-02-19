package models

type Origin struct {
	VCS  string `json:"VCS"`
	URL  string `json:"URL"`
	Hash string `json:"Hash"`
	Ref  string `json:"Ref"`
}

// ModuleInfo represents the module metadata
type ModuleInfo struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
	Origin  Origin `json:"Origin"`
}
