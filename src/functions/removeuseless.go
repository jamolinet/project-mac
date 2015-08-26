package functions

import (
	"github.com/project-mac/src/data"
	"reflect"
)

type RemoveUseless struct {
	//The filter used to remove attributes
	removeFilter Remove
	//The type of attribute to delete
	maxVariancePercentage float64
	input, output         data.Instances
}

func NewRemoveUseless() RemoveUseless {
	var ru RemoveUseless
	ru.maxVariancePercentage = 99.0
	return ru
}

func (m *RemoveUseless) Input(instances data.Instances) {
	if m.removeFilter.IsNotNil() {
		m.removeFilter.Exec(instances)
		m.output = m.removeFilter.outputFormat
	}
	for _, instance := range instances.Instances() {
		m.input.Add(instance)
	}
}

// Execute the filter
func (m *RemoveUseless) Exec(instances data.Instances) {
	if !m.removeFilter.IsNotNil() {

		//establish attributes to remove from first batch
		toFilter := m.input
		attsToDelete := make([]int, toFilter.NumAttributes())
		numToDelete := 0
		for i := range toFilter.Attributes() {
			if i == toFilter.ClassIndex() {
				continue //skip class
			}
			stats := toFilter.AttributeStats(i)
			if stats.MissingCount == toFilter.NumInstances() {
				attsToDelete[numToDelete] = i
				numToDelete++
			} else if stats.DistinctCount < 2 {
				//remove constant attributes
				attsToDelete[numToDelete] = i
				numToDelete++
			} else if toFilter.Attribute(i).IsNominal() {
				//remove nominal attributes that vary too much
				variancePercent := float64(stats.DistinctCount) / float64(stats.TotalCount-stats.MissingCount) * 100.0
				if variancePercent > m.maxVariancePercentage {
					attsToDelete[numToDelete] = i
					numToDelete++
				}
			}
		}

		finalAttsToDelete := make([]int, numToDelete)
		copy(attsToDelete, finalAttsToDelete[:numToDelete+1])
		m.removeFilter = NewRemove()
		m.removeFilter.SetSelectedColumns(finalAttsToDelete)
		m.removeFilter.SetInvertSelection(false)
		m.removeFilter.SetInputFormat(toFilter)
		m.removeFilter.Exec(toFilter)
		
		outputDataset := m.removeFilter.outputFormat
		//restore old relation name to hide attribute filter stamp
		outputDataset.SetDatasetName(toFilter.DatasetName())
		m.SetOutputFormat(outputDataset)
		for _,inst := range outputDataset.Instances() {
			m.output.Add(inst)
		}
	}
}

func (m *RemoveUseless) SetInputFormat(insts data.Instances) {
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

	m.input.SetDatasetName(insts.DatasetName())
	m.input.SetAttributes(atts)
	//m.SetOutputFormat(insts)
}

// Sets the format of output instances
func (m *RemoveUseless) SetOutputFormat(insts data.Instances) {
	m.output = data.NewInstances()
	//Strings free structure, "cleanses" string types (i.e. doesn't contain references to the
	//strings seen in the past)
	newAtts := make([]data.Attribute, 0)
	for i, att := range m.input.Attributes() {
		if att.Type() == data.STRING {
			at := data.NewAttribute()
			at.SetName(att.Name())
			at.SetIndex(i)
			newAtts = append(newAtts, at)
		}
	}
	atts := m.input.Attributes()
	if len(newAtts) != 0 {
		for _, att := range newAtts {
			atts[att.Index()] = att
		}
	}
	//Rename the relation
	dataSetName := insts.DatasetName() + "-" + reflect.TypeOf(insts).String()
	m.output.SetDatasetName(dataSetName)
	m.output.SetClassIndex(m.input.ClassIndex())
	m.output.SetAttributes(atts)
}
