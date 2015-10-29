package functions

import (
	"fmt"
	"github.com/cosn/collections/queue"
	"github.com/project-mac/src/data"
	"math"
	"reflect"
)

type Normalize struct {
	/** The minimum values for numeric attributes. */
	minArray []float64
	/** The maximum values for numeric attributes. */
	maxArray []float64
	/** The translation of the output range. */
	translation float64
	/** The scaling factor of the output range. */
	scale         float64
	input, output data.Instances
	classIndex    int
	ignoreClass   bool
	outputQueue   queue.Q
	notNil bool
}

func NewNormalize() Normalize {
	var n Normalize
	n.scale = 1.0
	n.translation = 0
	n.notNil = true
	return n
}

func NewNormalizePtr() *Normalize {
	var n Normalize
	n.scale = 1.0
	n.translation = 0
	n.notNil =true
	return &n
}

func (m *Normalize) Exec(instances data.Instances) {
	fmt.Println(len(instances.Instances()))
	for _,instance := range instances.Instances() {
		//println("normalizing1")
		m.Input(instance)
	}
//	for _, in := range m.input.Instances() {
//		fmt.Println(in.RealValues(), in.NumAttributesTest())
//		fmt.Println(in.Indices())
//	}
	m.BatchFinished()
}

func (m *Normalize) Input(instance data.Instance) {
	m.outputQueue.Init()
	if m.minArray == nil {
		m.BufferInput(instance)
		//println("normalizing")
	} else {
		m.ConvertInstance(instance)
	}
}

func (m *Normalize) NotNil() bool {
	return m.notNil
}

// Adds the supplied input instance to the inputformat dataset for
// later processing
func (m *Normalize) BufferInput(inst data.Instance) {
	m.input.Add(inst)
}

func (m *Normalize) SetInputFormat(insts data.Instances) {
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
	m.minArray, m.maxArray = nil, nil
	m.SetOutputFormat(insts)
	//fmt.Println(len(m.input.Instances()), len(m.input.Attributes()))
	//fmt.Println(len(m.output.Instances()),len(m.output.Attributes()))
}

func (m *Normalize) SetOutputFormat(insts data.Instances) {
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

func (m *Normalize) BatchFinished() {
	if m.minArray == nil {
		input := m.input
		//Compute minimuns and maximuns
		m.minArray = make([]float64, input.NumAttributes())
		m.maxArray = make([]float64, input.NumAttributes())
		for i := 0; i < input.NumAttributes(); i++ {
			m.minArray[i] = math.NaN()
		}
		//fmt.Println(len(m.maxArray),"m.maxArray")
		for j := 0; j < input.NumInstances(); j++ {
			value := input.Instance(j).Float64Slice(input.NumAttributes())
			//fmt.Println(value, "value", input.ClassIndex())
			//fmt.Println(input.NumAttributes(),"num`")
			for i := 0; i < input.NumAttributes(); i++ {
				if input.Attribute(i).IsNumeric() && input.ClassIndex() != i {
					if !math.IsNaN(value[i]) {
						if math.IsNaN(m.minArray[i]) {
							m.minArray[i], m.maxArray[i] = value[i], value[i]
						} else {
							if value[i] < m.minArray[i] {
								m.minArray[i] = value[i]
							}
							if value[i] > m.maxArray[i] {
								m.maxArray[i] = value[i]
							}
						}
					}
				}
			}
		}

		//Convert pending input instances
		for _, inst := range input.Instances() {
			m.ConvertInstance(inst)
		}
	}
}

//Convert a single instance over. The converted instance is added to the
//     * end of the output queue.

func (m *Normalize) ConvertInstance(instance data.Instance) {
	inst := data.NewInstance()
	//It's always a sparse instance
	newVals := make([]float64, instance.NumAttributesSparse())
	newIndices := make([]int, instance.NumAttributesSparse())
	vals := instance.Float64Slice(m.input.NumAttributes())
 // fmt.Println(vals, m.minArray, m.maxArray) //ok so far
	ind := 0
	for j, att := range m.input.Attributes() {
		var value float64
		if att.IsNumeric() && !math.IsNaN(vals[j]) && m.input.ClassIndex() != j {
			if math.IsNaN(m.minArray[j]) || m.maxArray[j] == m.minArray[j] {
				value = 0
			} else {
				//fmt.Println("je je je im in", vals[j],m.minArray[j],m.maxArray[j],m.scale,m.translation)
				value = (vals[j]-m.minArray[j])/(m.maxArray[j]-m.minArray[j])*m.scale + m.translation
				//fmt.Println(value)
				if math.IsNaN(value) {
					panic(fmt.Sprint("A NaN value was generated while normalizing ", att.Name()))
				}
			}
			if value != 0 {
				newVals[ind] = value
				newIndices[ind] = j
				ind++
			}
			//fmt.Println(value)
		} else {
			value = vals[j]
			if value != 0.0 {
				newVals[ind] = value
				newIndices[ind] = j
				ind++
			}
			//fmt.Println(value)
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

func (m *Normalize) OutputAll() data.Instances {
	return m.output
}

func (m *Normalize) Output() data.Instance {
	if !m.outputQueue.IsEmpty() {
		if result, ok := m.outputQueue.Pop().(data.Instance); ok {
			return result
		}
	}
	return data.NewInstance()
}
