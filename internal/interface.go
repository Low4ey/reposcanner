package internal

type DependecyResolver interface {
	FetchDependecy() error
	GetModeFile() string
	GetRepoURL() string
	GetVersion() string
	ValidateUrl() (bool, error)
}
