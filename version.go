package main

// VersionInfo represents version information
type VersionInfo struct {
	Service   string `json:"service"`
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
}
