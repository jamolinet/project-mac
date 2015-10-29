package functions

import (
	"github.com/project-mac/src/data"
	"reflect"
)

type RemoveUseless struct {
	//The filter used to remove attributes
	RemoveFilter Remove
	//The type of attribute to delete
	MaxVariancePercentage float64
	Input_, Output_         data.Instances
	NotNil bool
}

func NewRemoveUseless() RemoveUseless {
	var ru RemoveUseless
	ru.MaxVariancePercentage = 99.0
	ru.NotNil = true
	return ru
}

func (m *RemoveUseless) Input(instances data.Instances) {
	if m.RemoveFilter.IsNotNil() {
		m.RemoveFilter.Exec(instances)
		m.Output_ = m.RemoveFilter.OutputFormat
	}
	for _, instance := range instances.Instances() {
		m.Input_.Add(instance)
	}
}

// Execute the filter
func (m *RemoveUseless) Exec(instances data.Instances) {
	if !m.RemoveFilter.IsNotNil() {

		//establish attributes to remove from first batch
		toFilter := m.Input_
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
				if variancePercent > m.MaxVariancePercentage {
					attsToDelete[numToDelete] = i
					numToDelete++
				}
			}
		}

		finalAttsToDelete := make([]int, numToDelete)
		copy(attsToDelete, finalAttsToDelete[:numToDelete+1])
		m.RemoveFilter = NewRemove()
		m.RemoveFilter.SetSelectedColumns(finalAttsToDelete)
		m.RemoveFilter.SetInvertSelection(false)
		m.RemoveFilter.SetInputFormat(toFilter)
		m.RemoveFilter.Exec(toFilter)
		
		outputDataset := m.RemoveFilter.OutputFormat
		//restore old relation name to hide attribute filter stamp
		outputDataset.SetDatasetName(toFilter.DatasetName())
		m.SetOutputFormat(outputDataset)
		for _,inst := range outputDataset.Instances() {
			m.Output_.Add(inst)
		}
	}
}

func (m *RemoveUseless) SetInputFormat(insts data.Instances) {
	m.Input_ = data.NewInstances()
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

	m.Input_.SetDatasetName(insts.DatasetName())
	m.Input_.SetAttributes(atts)
	//m.SetOutputFormat(insts)
}

// Sets the format of Output_ instances
func (m *RemoveUseless) SetOutputFormat(insts data.Instances) {
	m.Output_ = data.NewInstances()
	//Strings free structure, "cleanses" string types (i.e. doesn't contain references to the
	//strings seen in the past)
	newAtts := make([]data.Attribute, 0)
	for i, att := range m.Input_.Attributes() {
		if att.Type() == data.STRING {
			at := data.NewAttribute()
			at.SetName(att.Name())
			at.SetIndex(i)
			newAtts = append(newAtts, at)
		}
	}
	atts := m.Input_.Attributes()
	if len(newAtts) != 0 {
		for _, att := range newAtts {
			atts[att.Index()] = att
		}
	}
	//Rename the relation
	dataSetName := insts.DatasetName() + "-" + reflect.TypeOf(insts).String()
	m.Output_.SetDatasetName(dataSetName)
	m.Output_.SetClassIndex(m.Input_.ClassIndex())
	m.Output_.SetAttributes(atts)
}

func (r *RemoveUseless) Output() data.Instances {
	return r.Output_
}

func (r *RemoveUseless) ConvertAndReturn(instance data.Instance) data.Instance {
	return r.RemoveFilter.ConvertAndReturn(instance)
}
