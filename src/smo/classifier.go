package smo

import (
	"fmt"
	datas "github.com/project-mac/src/data"
	"github.com/project-mac/src/functions"
)

// This is something like multi filter classifier
type Classifier struct {
	*SMO
	debug             bool
	filteredInstances datas.Instances
	STWV              functions.StringToWordVector
	//ranker functions.Ranker
	//infogain functions.InfoGain
	AS functions.AttributeSelection
}

func NewClassifier(smo *SMO) Classifier {
	var c Classifier
	c.SMO = smo
	return c
}

func (m *Classifier) SetStringToWordVector(STWV functions.StringToWordVector) {
	m.STWV = STWV
}

//func (m *Classifier) SetRanker(ranker functions.Ranker) {
//	m.ranker = ranker
//}
//
//func (m *Classifier) SetInfoGain(infogain functions.InfoGain) {
//	m.infogain = infogain
//}

func (m *Classifier) SetAS(AS functions.AttributeSelection) {
	m.AS = AS
}

func (m *Classifier) BuildClassifier(data datas.Instances) {
	if &m.SMO == nil {
		panic("No base classifier set!")
	}
	data.DeleteWithMissingClass()
	//The data is already filtered
	m.STWV.SetInputFormatAndOutputFormat(data)
	data = m.STWV.Exec()
	m.AS.StartSelection(data)
	data = m.AS.Output()
	fmt.Print()
	m.filteredInstances = data.StringFreeStructure()
	m.SMO.BuildClassifier(data)
}

func (m *Classifier) DistributionForInstance(instance datas.Instance, classIndex, numClasses int) []float64 {

	instance = m.STWV.Input(instance)
	instance = m.AS.ConvertInstance(instance)
	fmt.Print()

	return m.SMO.DistributionForInstance(instance, numClasses)
}

func (m *Classifier) ClassifyInstances(instance datas.Instance, data *datas.Instances) float64 {

	dist := m.DistributionForInstance(instance, 0, data.NumClasses())
	if dist == nil {
		panic("Nil distribution predicted")
	}
	switch data.Attribute(data.ClassIndex()).Type() {
	case datas.NOMINAL:
		max := 0.0
		maxIndex := 0
		for i, d := range dist {
			if d > max {
				maxIndex = i
				max = d
			}
		}
		if max > 0 {
			return float64(maxIndex)
		} else {
			return 0
		}
	case datas.NUMERIC:
		return dist[0]
	default:
		return 0
	}

}
