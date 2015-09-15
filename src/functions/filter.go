package functions

import (
	"github.com/project-mac/src/data"
)

type Filter interface {
	Exec(data.Instances)
	SetOutputFormat(data.Instances)
	SetInputFormat(data.Instances)
	BatchFinished()
	Input(data.Instance)
	bufferInput(data.Instance)
	ConvertInstance(data.Instance)
	OutputAll() data.Instances
	Output() data.Instance
	NotNil() bool
}