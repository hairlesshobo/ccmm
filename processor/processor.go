package processor

import "github.com/hairlesshobo/go-import-media/model"

type Processor interface {
	SetSourceDir(sourceDir string)
	CheckSource() bool
	ProcessSource()
	EnumerateFiles() []model.SourceFile
}
