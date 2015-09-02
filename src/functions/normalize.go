package functions

import (
	"github.com/project-mac/src/data"
)

type Normalize struct {
	/** The minimum values for numeric attributes. */
	minArray []float64
	/** The maximum values for numeric attributes. */
	maxArray []float64
	/** The translation of the output range. */
	translation float64
	/** The scaling factor of the output range. */
	scale         float64
	input, output data.Instances
	classIndex int
	ignoreClass bool
}

func NewNormalize() Normalize {
	var n Normalize
	n.scale = 1.0
	n.translation = 0
	return n
}

func (m *Normalize) SetInputFormat(insts data.Instances) {
	m.input = data.NewInstances()
	newAtts := make([]data.Attribute, 0)
	for i, att := range insts.Attributes() {
		if att.Type() == data.STRING {
			at := data.NewAttribute()
			at.SetName(att.Name())
			at.SetIndex(i)
			newAtts = append(newAtts, at)
		}
	}
	atts := insts.Attributes()
	if len(newAtts) != 0 {
		for _, att := range newAtts {
			atts[att.Index()] = att
		}
	}
	m.input.SetClassIndex(insts.ClassIndex())
	if m.ignoreClass {
		m.classIndex = m.input.ClassIndex()
		m.input.SetClassIndex(-1)
	}
	m.input.SetDatasetName(insts.DatasetName())
	m.input.SetAttributes(atts)
	m.minArray, m.maxArray = nil, nil
	m.SetOutputFormat(insts)
}
