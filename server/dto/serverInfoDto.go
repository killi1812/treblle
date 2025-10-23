package dto

type ServerInfoDto struct {
	Build          string `json:"build"`
	Version        string `json:"version"`
	CommitHash     string `json:"commitHash"`
	BuildTimestamp string `json:"buildTimestamp"`
}
