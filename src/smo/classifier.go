package smo

import (
	"fmt"
	datas "github.com/project-mac/src/data"
	"github.com/project-mac/src/functions"
)

// This is something like multi filter classifier
type Classifier struct {
	SMO
	debug             bool
	filteredInstances datas.Instances
	stwv              functions.StringToWordVector
	//ranker functions.Ranker
	//infogain functions.InfoGain
	as functions.AttributeSelection
}

func NewClassifier(smo SMO) Classifier {
	var c Classifier
	c.SMO = smo
	return c
}

func (m *Classifier) SetStringToWordVector(stwv functions.StringToWordVector) {
	m.stwv = stwv
}

//func (m *Classifier) SetRanker(ranker functions.Ranker) {
//	m.ranker = ranker
//}
//
//func (m *Classifier) SetInfoGain(infogain functions.InfoGain) {
//	m.infogain = infogain
//}

func (m *Classifier) SetAS(as functions.AttributeSelection) {
	m.as = as
}

func (m *Classifier) BuildClassifier(data datas.Instances) {
	if &m.SMO == nil {
		panic("No base classifier set!")
	}
	data.DeleteWithMissingClass()
	//The data is already filtered
	m.stwv.SetInputFormatAndOutputFormat(data)
	data = m.stwv.Exec()
	m.as.StartSelection(data)
	data = m.as.Output()
	fmt.Print()
	//	for _, in := range data.Instances() {
	//		fmt.Println(in.RealValues(), in.NumAttributesTest())
	//		fmt.Println(in.Indices())
	//	}
	m.filteredInstances = data.StringFreeStructure()
	m.SMO.BuildClassifier(data)
}

func (m *Classifier) DistributionForInstance(instance datas.Instance, classIndex, numClasses int) []float64 {
//	
//		fmt.Println(instance.RealValues(),"hhhhhhhhhhhh", instance.NumAttributesTest())
//		fmt.Println(instance.Indices())
	
	instance = m.stwv.Input(instance)
	instance = m.as.ConvertInstance(instance)
//	fmt.Println(instance.RealValues(),"ggggggggggg", instance.NumAttributesTest())
//		fmt.Println(instance.Indices())
	//fmt.Println(instance.RealValues(), "instance.Indices()")
	fmt.Print()
	//ok so far
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
			return instance.MissingValue
		}
	case datas.NUMERIC:
		return dist[0]
	default:
		return instance.MissingValue
	}

}
