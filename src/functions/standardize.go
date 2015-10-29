package functions

import (
	"github.com/cosn/collections/queue"
	"github.com/project-mac/src/data"
	"math"
	"reflect"
	//"fmt"
)

type Standardize struct {
	//The Means
	//The variances
	Means, StdDevs []float64
	Input_, Output_  data.Instances
	IgnoreClass    bool
	ClassIndex_     int
	OutputQueue    queue.Q
	NotNil_ bool
}

func NewStandardize() Standardize {
	var s Standardize
	s.NotNil_ = true
	return s
}

func NewStandardizePtr() *Standardize {
	
	return &Standardize{NotNil_: true}
}

// Input an instance for filtering
func (m *Standardize) Input(instance data.Instance) {
	m.OutputQueue.Init()
	if m.Means == nil {
		m.BufferInput(instance)
	} else {
		m.ConvertInstance(instance)
	}
}

func (m *Standardize) NotNil() bool {
	return m.NotNil_
}

// Adds the supplied Input_ instance to the inputformat dataset for
// later processing
func (m *Standardize) BufferInput(inst data.Instance) {
	m.Input_.Add(inst)
}

func (m *Standardize) SetInputFormat(insts data.Instances) {
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
	m.Input_.SetClassIndex(insts.ClassIndex())
	if m.IgnoreClass {
		m.ClassIndex_ = m.Input_.ClassIndex()
		m.Input_.SetClassIndex(-1)
	}
	m.Input_.SetDatasetName(insts.DatasetName())
	m.Input_.SetAttributes(atts)
	m.Means, m.StdDevs = nil, nil
	m.SetOutputFormat(insts)
}

func (m *Standardize) BatchFinished() {
	if m.Means == nil {
		Input_ := m.Input_
		//Compute minimuns and maximuns
		m.Means = make([]float64, Input_.NumAttributes())
		m.StdDevs = make([]float64, Input_.NumAttributes())

		for i := 0; i < Input_.NumAttributes(); i++ {
			if Input_.Attribute(i).IsNumeric() && Input_.ClassIndex() != i {
				m.Means[i] = Input_.MeanOrMode(i)
				m.StdDevs[i] = math.Sqrt(Input_.Variance(i))
			}
		}

		//Convert pending Input_ instances
		for _, inst := range Input_.Instances() {
			m.ConvertInstance(inst)
		}
	}
}

func (m *Standardize) SetOutputFormat(insts data.Instances) {
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
	m.OutputQueue.Init()
}

//Convert a single instance over. The converted instance is added to the
//     * end of the Output_ queue.

func (m *Standardize) ConvertInstance(instance data.Instance) {
	inst := data.NewInstance()
	//It's always a sparse instance
	newVals := make([]float64, instance.NumAttributes())
	newIndices := make([]int, instance.NumAttributes())
	vals := instance.ToFloat64Slice()
	ind := 0
	for j := 0; j < instance.NumAttributes();j++ {
		var value float64
		if m.Input_.Attributes()[j].IsNumeric() && (!math.IsNaN(vals[j]) && m.Input_.ClassIndex() != j) {

			// Just subtract the mean if the standard deviation is zero
			if m.StdDevs[j] > 0 {
				value = (vals[j] - m.Means[j]) / m.StdDevs[j]
			} else {
				value = vals[j] - m.Means[j]
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
	inst = data.NewSparseInstanceWithIndexes(instance.Weight(), tempVals, tempInd, m.Input_.Attributes())
	m.OutputQueue.Push(inst)
	m.Output_.Add(inst)
}

func (m *Standardize) Exec(instances data.Instances) {
	for _, instance := range instances.Instances() {
		m.Input(instance)
	}
	m.BatchFinished()
}

func (m *Standardize) OutputAll() data.Instances {
	return m.Output_
}

func (m *Standardize) Output() data.Instance {
	if !m.OutputQueue.IsEmpty() {
		if result, ok := m.OutputQueue.Pop().(data.Instance); ok {
			return result
		}
	}
	return data.NewInstance()
}
