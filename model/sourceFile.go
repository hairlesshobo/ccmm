package model

import "time"

type SourceFile struct {
	FileName    string
	SourcePath  string
	Size        uint64
	MediaType   string
	SourceName  string
	CaptureDate time.Time
}
