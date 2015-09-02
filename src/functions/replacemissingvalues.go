package functions

import (
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/utils"
	"math"
	"reflect"
)

type ReplaceMissingValues struct {
	modesAndMeans          []float64
	input, output          data.Instances
	firstTime, ignoreClass bool
	classIndex             int
}

func NewReplacingMissingValues() ReplaceMissingValues {
	var rmv ReplaceMissingValues
	rmv.modesAndMeans = nil
	rmv.firstTime = true
	return rmv
}

// Execute the filter
func (m *ReplaceMissingValues) Exec(instances data.Instances) {
	m.SetInputFormat(instances)
	for _, instance := range instances.Instances() {
		if m.modesAndMeans == nil {
			m.bufferInput(instance)
		} else {
			m.convertInstance(instance)
		}
	}

}

// Convert a single instance over.
func (m *ReplaceMissingValues) convertInstance(instance data.Instance) data.Instance {
	inst := data.NewInstance()
	//Instances for the moment are always SparseInstances
	vals := make([]float64, len(instance.RealValues()))
	indices := make([]int, len(instance.RealValues()))
	num := 0
	for j := 0; j < len(instance.RealValues()); j++ {
		if instance.IsMissingValue(j) && (m.input.ClassIndex() != instance.Index(j)) && 
		(m.input.Attribute(j).IsNominal() || m.input.Attribute(j).IsNumeric()) { /*inst.attributeSparse(j).isNominal() */
			if m.modesAndMeans[instance.Index(j)] != 0 {
				vals[num] = m.modesAndMeans[instance.Index(j)]
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
		inst.SparseInstance(instance.Weight(), vals, indices,len(m.output.Attributes()),m.output.Attributes())
	}
	m.output.Add(inst)
	return inst
}
func (m *ReplaceMissingValues) BatchFinished() {
	if m.modesAndMeans == nil {
		//Compute modes and means
		sumOfWeights := m.input.SumOfWeights()
		counts := make([][]float64, m.input.NumAttributes())
		for i, att := range m.input.Attributes() {
			if att.IsNominal() {
				counts[i] = make([]float64, att.NumValues())
				if len(counts[i]) > 0 {
					counts[i][0] = sumOfWeights
				}
			}
		}
		sums := make([]float64, m.input.NumAttributes())
		for i := 0; i < len(sums); i++ {
			sums[i] = sumOfWeights
		}

		results := make([]float64, m.input.NumAttributes())
		for _, inst := range m.input.Instances() {
			for i := 0; i < len(inst.RealValues()); i++ {
				if !inst.IsMissingValue(i) {
					value := inst.ValueSparse(i)
					if m.input.Attribute(i).IsNominal() { //inst.attributeSparse(i).isNominal()
						if len(counts[inst.Index(i)]) > 0 {
							counts[inst.Index(i)][int(value)] += inst.Weight()
							counts[inst.Index(i)][0] -= inst.Weight()
						}
					} else if m.input.Attribute(i).IsNumeric() {
						results[inst.Index(i)] += inst.Weight() * inst.ValueSparse(i)
					}
				} else {
					if m.input.Attribute(i).IsNominal() {
						if len(counts[inst.Index(i)]) > 0 {
							counts[inst.Index(i)][0] -= inst.Weight()
						}
					} else if m.input.Attribute(i).IsNumeric() {
						sums[inst.Index(i)] -= inst.Weight()
					}
				}
			}
		}
		m.modesAndMeans = make([]float64, m.input.NumAttributes())
		for i, att := range m.input.Attributes() {
			if att.IsNominal() {
				if len(counts[i]) == 0 {
					m.modesAndMeans[i] = math.NaN()
				} else {
					m.modesAndMeans[i] = float64(utils.MaxIndex(counts[i]))
				}
			} else if att.IsNumeric() {
				if utils.Gr(sums[i], 0) {
					m.modesAndMeans[i] = results[i] / sums[i]
				}
			}
		}
		
		//Convert pending input instances
		for _,inst := range m.input.Instances() {
			m.convertInstance(inst)
		}
	}
}

// Adds the supplied input instance to the inputformat dataset for
// later processing
func (m *ReplaceMissingValues) bufferInput(inst data.Instance) {
	m.input.Add(inst)
}

// Sets the format of the input instances.
func (m *ReplaceMissingValues) SetInputFormat(insts data.Instances) {
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
	m.modesAndMeans = nil
	m.SetOutputFormat(insts)
}

// Sets the format of output instances
func (m *ReplaceMissingValues) SetOutputFormat(insts data.Instances) {
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

// Returns the output
func (m *ReplaceMissingValues) Output() data.Instances {
	return m.output
}

// This method does the function of calling in weka convertInstance(Instance) and then output()
// due in this implementation does not exists the m_OutputQueue value
func (m *ReplaceMissingValues) ConvertAndReturn(instance data.Instance) data.Instance {
	return m.convertInstance(instance)
}
