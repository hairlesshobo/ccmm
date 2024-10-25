package model

import "time"

type SourceFile struct {
	FileName    string
	SourcePath  string
	Size        int64
	MediaType   string
	SourceName  string
	CaptureDate time.Time
	FileModTime time.Time
}
