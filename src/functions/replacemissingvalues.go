package functions

import (
	"github.com/cosn/collections/queue"
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/utils"
	"math"
	"reflect"
	"fmt"
)

type ReplaceMissingValues struct {
	ModesAndMeans_          []float64
	Input_, Output_          data.Instances
	FirstTime, IgnoreClass bool
	ClassIndex             int
	OutputQueue queue.Q
	NotNil bool
}

func NewReplacingMissingValues() ReplaceMissingValues {
	var rmv ReplaceMissingValues
	rmv.ModesAndMeans_ = nil
	rmv.FirstTime = true
	rmv.OutputQueue.Init()
	rmv.NotNil = true
	return rmv
}

// Execute the filter
func (m *ReplaceMissingValues) Exec(instances data.Instances) {
	m.SetInputFormat(instances)
	for _, instance := range instances.Instances() {
		if m.ModesAndMeans_ == nil {
			m.BufferInput(instance)
		} else {
			m.ConvertInstance(instance)
		}
	}

}

// Input an instance for filtering
func (m *ReplaceMissingValues) Input(instance data.Instance) {
	m.OutputQueue.Init()
	if m.ModesAndMeans_ == nil {
		m.BufferInput(instance)
	} else {
		m.ConvertInstance(instance)
	}
}

func (m ReplaceMissingValues) ModesAndMeans() []float64 {
	return m.ModesAndMeans_
}

// Convert a single instance over.
func (m *ReplaceMissingValues) ConvertInstance(instance data.Instance) data.Instance {
	fmt.Print()
	inst := data.NewInstance()
	//Instances for the moment are always SparseInstances
	vals := make([]float64, len(instance.RealValues()))
	indices := make([]int, len(instance.RealValues()))
	num := 0
	for j := 0; j < len(instance.RealValues()); j++ {
		if instance.IsMissingSparse(j) && (m.Input_.ClassIndex() != instance.Index(j)) &&
			(m.Input_.Attribute(j).IsNominal() || m.Input_.Attribute(j).IsNumeric()) { /*inst.attributeSparse(j).isNominal() */
			if m.ModesAndMeans_[instance.Index(j)] != 0 {
				vals[num] = m.ModesAndMeans_[instance.Index(j)]
				indices[num] = instance.Index(j)
				num++
			}
		} else {
			vals[num] = instance.ValueSparse(j)
			indices[num] = instance.Index(j)
			num++
		}
	}
	if num == len(instance.RealValues()) {
		inst.SparseInstance(instance.Weight(), vals, indices, len(m.Output_.Attributes()), m.Output_.Attributes())
	}
	m.Output_.Add(inst)
	m.OutputQueue.Push(inst)
	return inst
}
func (m *ReplaceMissingValues) BatchFinished() {
	if m.ModesAndMeans_ == nil {
		//fmt.Println("je je j jeanjaja")
		//Compute modes and means
		sumOfWeights := m.Input_.SumOfWeights()
		counts := make([][]float64, m.Input_.NumAttributes())
		for i, att := range m.Input_.Attributes() {
			if att.IsNominal() {
				counts[i] = make([]float64, att.NumValues())
				if len(counts[i]) > 0 {
					counts[i][0] = sumOfWeights
				}
			}
		}
		sums := make([]float64, m.Input_.NumAttributes())
		for i := 0; i < len(sums); i++ {
			sums[i] = sumOfWeights
		}

		results := make([]float64, m.Input_.NumAttributes())
		for _, inst := range m.Input_.Instances() {
			for i := 0; i < len(inst.RealValues()); i++ {
				if !inst.IsMissingValue(i) {
					value := inst.ValueSparse(i)
					if m.Input_.Attribute(i).IsNominal() { //inst.attributeSparse(i).isNominal()
						if len(counts[inst.Index(i)]) > 0 {
							counts[inst.Index(i)][int(value)] += inst.Weight()
							counts[inst.Index(i)][0] -= inst.Weight()
						}
					} else if m.Input_.Attribute(i).IsNumeric() {
						results[inst.Index(i)] += inst.Weight() * inst.ValueSparse(i)
					}
				} else {
					if m.Input_.Attribute(i).IsNominal() {
						if len(counts[inst.Index(i)]) > 0 {
							counts[inst.Index(i)][0] -= inst.Weight()
						}
					} else if m.Input_.Attribute(i).IsNumeric() {
						sums[inst.Index(i)] -= inst.Weight()
					}
				}
			}
		}
		m.ModesAndMeans_ = make([]float64, m.Input_.NumAttributes())
		for i, att := range m.Input_.Attributes() {
			if att.IsNominal() {
				if len(counts[i]) == 0 {
					m.ModesAndMeans_[i] = math.NaN()
				} else {
					m.ModesAndMeans_[i] = float64(utils.MaxIndex(counts[i]))
				}
			} else if att.IsNumeric() {
				if utils.Gr(sums[i], 0) {
					m.ModesAndMeans_[i] = results[i] / sums[i]
				}
			}
		}

		//Convert pending Input_ instances
		for _, inst := range m.Input_.Instances() {
			m.ConvertInstance(inst)
		}
	}
}

// Adds the supplied Input_ instance to the inputformat dataset for
// later processing
func (m *ReplaceMissingValues) BufferInput(inst data.Instance) {
	m.Input_.Add(inst)
}

// Sets the format of the Input_ instances.
func (m *ReplaceMissingValues) SetInputFormat(insts data.Instances) {
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
		m.ClassIndex = m.Input_.ClassIndex()
		m.Input_.SetClassIndex(-1)
	}
	m.Input_.SetDatasetName(insts.DatasetName())
	m.Input_.SetAttributes(atts)
	m.ModesAndMeans_ = nil
	m.SetOutputFormat(insts)
}

// Sets the format of Output_ instances
func (m *ReplaceMissingValues) SetOutputFormat(insts data.Instances) {
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

// Returns the Output_
func (m *ReplaceMissingValues) OutputAll() data.Instances {
	return m.Output_
}

func (m *ReplaceMissingValues) Output() data.Instance {
	if !m.OutputQueue.IsEmpty() {
		if result, ok := m.OutputQueue.Pop().(data.Instance); ok {
			return result
		}
	}
	return data.NewInstance()
}

// This method does the function of calling in weka ConvertInstance(Instance) and then Output_()
// due in this implementation does not exists the m_OutputQueue value
func (m *ReplaceMissingValues) ConvertAndReturn(instance data.Instance) data.Instance {
	return m.ConvertInstance(instance)
}
