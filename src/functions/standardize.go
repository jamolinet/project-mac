package functions

import (
	"github.com/cosn/collections/queue"
	"github.com/project-mac/src/data"
	"math"
	"reflect"
)

type Standardize struct {
	//The means
	//The variances
	means, stdDevs []float64
	input, output  data.Instances
	ignoreClass    bool
	classIndex     int
	outputQueue    queue.Q
}

func NewStandardize() Standardize {
	var s Standardize
	return s
}

func NewStandardizePtr() *Standardize {
	return new(Standardize)
}

// Input an instance for filtering
func (m *Standardize) Input(instance data.Instance) {
	if m.means == nil {
		m.bufferInput(instance)
	} else {
		m.ConvertInstance(instance)
	}
}

// Adds the supplied input instance to the inputformat dataset for
// later processing
func (m *Standardize) bufferInput(inst data.Instance) {
	m.input.Add(inst)
}

func (m *Standardize) SetInputFormat(insts data.Instances) {
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
	m.means, m.stdDevs = nil, nil
	m.SetOutputFormat(insts)
}

func (m *Standardize) BatchFinished() {
	if m.means == nil {
		input := m.input
		//Compute minimuns and maximuns
		m.means = make([]float64, input.NumAttributes())
		m.stdDevs = make([]float64, input.NumAttributes())

		for i := 0; i < input.NumAttributes(); i++ {
			if input.Attribute(i).IsNumeric() && input.ClassIndex() != i {
				m.means[i] = input.MeanOrMode(i)
				m.stdDevs[i] = math.Sqrt(input.Variance(i))
			}
		}

		//Convert pending input instances
		for _, inst := range input.Instances() {
			m.ConvertInstance(inst)
		}
	}
}

func (m *Standardize) SetOutputFormat(insts data.Instances) {
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
	m.outputQueue.Init()
}

//Convert a single instance over. The converted instance is added to the
//     * end of the output queue.

func (m *Standardize) ConvertInstance(instance data.Instance) {
	inst := data.NewInstance()
	//It's always a sparse instance
	newVals := make([]float64, instance.NumAttributes())
	newIndices := make([]int, instance.NumAttributes())
	vals := instance.RealValues()
	ind := 0
	for j, att := range m.input.Attributes() {
		var value float64
		if att.IsNumeric() && math.IsNaN(vals[j]) && m.input.ClassIndex() != j {

			// Just subtract the mean if the standard deviation is zero
			if m.stdDevs[j] > 0 {
				value = (vals[j] - m.means[j]) / m.stdDevs[j]
			} else {
				value = vals[j] - m.means[j]
			}
			if value != 0 {
				newVals[ind] = value
				newIndices[ind] = j
				ind++
			}
		} else {
			value = vals[j]
			if value != 0.0 {
				newVals[ind] = value
				newIndices[ind] = j
				ind++
			}
		}
	}
	tempVals := make([]float64, ind)
	tempInd := make([]int, ind)
	copy(tempVals, newVals)
	copy(tempInd, newIndices)
	inst = data.NewSparseInstanceWithIndexes(instance.Weight(), tempVals, tempInd, m.input.Attributes())
	m.outputQueue.Push(inst)
	m.output.Add(inst)
}

func (m *Standardize) Exec(instances data.Instances) {
	for _, instance := range instances.Instances() {
		m.Input(instance)
	}
	m.BatchFinished()
}

func (m *Standardize) OutputAll() data.Instances {
	return m.output
}

func (m *Standardize) Output() data.Instance {
	if !m.outputQueue.IsEmpty() {
		if result, ok := m.outputQueue.Pop().(data.Instance); ok {
			return result
		}
	}
	return data.NewInstance()
}
