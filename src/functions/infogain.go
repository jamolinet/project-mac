package functions

import (
	"data"
)

type InfoGain struct {
	missingMerge, binarize bool
	infoGains []float64
}

func (ig *InfoGain) NewInfoGain() {
	ig.missingMerge = true
	ig.binarize = false
	ig.infoGains = make([]float64,0)
}

func(ig *InfoGain) buildEvaluator(instances *data.Instances) {
	
}

func(ig *InfoGain) SetMissingMerge(mm bool) {
	ig.missingMerge = mm
}

func(ig *InfoGain) SetBinarize( binarize bool) {
	ig.binarize = binarize
}

func(ig *InfoGain) InfoGain() []float64 {
	return ig.infoGains;
}