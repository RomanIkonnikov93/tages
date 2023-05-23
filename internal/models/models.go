package models

const (
	UploadDownloadParallelCount = 10
	FilesInfoParallelCount      = 100
)

// Record types.
const (
	Download = "download"
	Upload   = "upload"
	Info     = "info"
)

type Record struct {
	RequestType string
	FileName    string
	Created     string
	Updated     string
	File        []byte
}
