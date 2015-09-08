package smo

import (
	datas "github.com/project-mac/src/data"
)

// This is something like multi filter classifier
type Classifier struct {
	SMO
	debug             bool
	filteredInstances datas.Instances
}

func NewClassifier() Classifier {
	var c Classifier
	c.SMO = NewSMO()
	return c
}

func (m *Classifier) BuildClassifier(data datas.Instances) {
	if &m.SMO == nil {
		panic("No base classifier set!")
	}
	data.DeleteWithMissingClass()
	//The data is already filtered
	m.filteredInstances =  data.StringFreeStructure()
	m.SMO.BuildClassifier(data) 
}

func (m *Classifier) DistributionForInstance(instance datas.Instance, classIndex, numClasses int) []float64 {
	return m.SMO.DistributionForInstance(instance, numClasses)
}
	

