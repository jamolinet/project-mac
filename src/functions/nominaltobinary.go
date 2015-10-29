package functions

import (
	"fmt"
	"github.com/cosn/collections/queue"
	"github.com/project-mac/src/data"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Converts all nominal attributes into binary Numeric attributes.
// An attribute with k values is transformed into k binary attributes if the class is nominal
// (using the one-attribute-per-value approach). Binary attributes are left binary, if option '-A' is not given.
// If the class is Numeric, you might want to use the supervised version of this filter.

type NominalToBinary struct {
	// Stores which Columns to act on
	Columns            []int
	SelectedAttributes []int
	IsRangeInUse       bool
	// Are the new attributes going to be nominal or Numeric ones?
	Numeric bool
	// Are all values transformed into new attributes?
	TransformAll bool
	// Whether we need to transform at all
	NeedToTransform bool
	Input_, Output_   data.Instances
	FirstTime       bool
	InvertSel       bool
	OutputQueue     queue.Q
	IsNil string
}

func NewNominalToBinary() NominalToBinary {
	var ntb NominalToBinary
	ntb.Numeric = true
	ntb.NeedToTransform, ntb.TransformAll = false, false
	ntb.IsRangeInUse = false
	ntb.FirstTime = true
	ntb.InvertSel = false
	ntb.OutputQueue.Init()
	ntb.IsNil = "no"
	return ntb
}

func NewNominalToBinaryWithInstances(data data.Instances) NominalToBinary {
	var ntb NominalToBinary
	ntb.Numeric = true
	ntb.NeedToTransform, ntb.TransformAll = false, false
	ntb.IsRangeInUse = false
	ntb.FirstTime = true
	ntb.Input_ = data
	ntb.OutputQueue.Init()
	ntb.IsNil = "no"
	return ntb
}

// Execute the filter
func (m *NominalToBinary) Exec(instances data.Instances) error {
	if m.FirstTime {
		m.FirstTime = false
	}
	m.SetInputFormat(instances)

	for _, instance := range m.Input_.Instances() {
		m.ConvertInstance(instance)
	}
	return nil
}

// Input_ an instance for filtering
func (m *NominalToBinary) Input(instance data.Instance) {
	m.ConvertInstance(instance)
}

// Convert an instance over
func (m *NominalToBinary) ConvertInstance(instance data.Instance) data.Instance {
	if !m.NeedToTransform {
		m.Output_.Add(instance)
		return instance
	}

	vals := make([]float64, m.Output_.NumAttributes())
	attSoFar := 0

	for j, att := range m.Input_.Attributes() {
		if !att.IsNominal() || (j == m.Input_.ClassIndex() || !m.IsInRange(j)) {
			vals[attSoFar] = instance.Value(j)
			attSoFar++
		} else {
			if att.NumValues() <= 2 && (!m.TransformAll) {
				vals[attSoFar] = instance.Value(j)
				attSoFar++
			} else {
				if instance.IsMissingValue(j) {
					for k := 0; k < att.NumValues(); k++ {
						vals[attSoFar+k] = instance.Value(j)
					}
				} else {
					for k := 0; k < att.NumValues(); k++ {
						if k == int(instance.Value(j)) {
							vals[attSoFar+k] = 1
						} else {
							vals[attSoFar+k] = 0
						}
					}
				}
				attSoFar += att.NumValues()
			}
		}
	}
	inst := data.NewInstance()
	inst.SetWeight(instance.Weight())
	//Get values and indexes different to zero
	indices := make([]int, 0, len(vals))
	values := make([]float64, 0, len(vals))
	idx := 0
	for i, val := range vals {
		if val != 0 {
			values[idx] = vals[i]
			indices[idx] = i
			idx++
		}
	}
	inst.SetIndices(indices)
	inst.SetRealValues(values)
	for k, i := range indices {
		if m.Output_.Attribute(i).IsNominal() {
			if math.IsNaN(values[k]) {
				inst.AddValues("?")
			} else {
				inst.AddValues(m.Output_.Attributes()[i].Values()[int(values[k])])
			}
		} else if m.Output_.Attributes()[i].IsNominal() && !m.Output_.Attributes()[i].IsString() {
			inst.AddValues(m.Output_.Attributes()[i].Values()[i])
		} else {
			inst.AddValues(m.Output_.Attributes()[i].Name())
		}
	}
	inst.SetNumAttributes(len(values))
	m.Output_.Add(inst)
	m.OutputQueue.Push(inst)
	println("pushed")
	return inst
}

func (m *NominalToBinary) SetInputFormat(data data.Instances) {
	m.Input_ = data
	m.GetSelectedAttributes(len(data.Attributes()))
	m.SetOuputFormat()
}

// Set the Output_ format if the class is nominal
func (m *NominalToBinary) SetOuputFormat() {
	newAtts := make([]data.Attribute, 0)
	var newClassIndex int
	var attributeName string
	outputFormat := data.NewInstances()
	vals := make([]string, 2)

	//Compute new attributes
	m.NeedToTransform = false
	for i, att := range m.Input_.Attributes() {
		if att.IsNominal() && i != m.Input_.ClassIndex() && (att.NumValues() > 2 || m.TransformAll || m.Numeric) {
			m.NeedToTransform = true
			break
		}
	}

	newClassIndex = m.Input_.ClassIndex()
	for j, att := range m.Input_.Attributes() {
		if !att.IsNominal() || j == m.Input_.ClassIndex() || !m.IsInRange(j) {
			newAtts = append(newAtts, att)
		} else {
			if att.NumValues() <= 2 && !m.TransformAll {
				if m.Numeric {
					atemp := data.NewAttribute()
					atemp.SetName(att.Name())
					newAtts = append(newAtts, atemp)
				} else {
					newAtts = append(newAtts, att)
				}
			} else {
				if newClassIndex >= 0 && j < m.Input_.ClassIndex() {
					newClassIndex += att.NumValues() - 1
				}

				//Compute values for new attributes
				for k := 0; k < att.NumValues(); k++ {
					attributeName = att.Name() + "=" + att.Value(k)
					if m.Numeric {
						atemp := data.NewAttribute()
						atemp.SetName(attributeName)
						newAtts = append(newAtts, atemp)
					} else {
						vals[0], vals[1] = "f", "t"
						atemp := data.NewAttribute()
						atemp.SetName(attributeName)
						atemp.SetValues(vals)
						newAtts = append(newAtts, atemp)
					}
				}
			}
		}
	}
	outputFormat.SetAttributes(newAtts)
	outputFormat.SetDatasetName(m.Input_.DatasetName())
	outputFormat.SetClassIndex(newClassIndex)
	m.Output_ = outputFormat
}

func (r *NominalToBinary) IsInRange(f int) bool {
	// omit the class from the evaluation
	//	if r.hasClass && r.classIndex == f {
	//		return true
	//	}
	//
	//	if !r.IsRangeInUse || len(r.notInclude) == 0 {
	//		return false
	//	}

	for _, sel := range r.Columns {
		if sel == f {
			return true
		}
	}

	return false
}

func (ntb *NominalToBinary) SetRange(rang string) {
	if strings.EqualFold(rang, "") {
		panic("The range cannot be empty")
	}
	if strings.EqualFold(rang, "all") {
		ntb.IsRangeInUse = false
		return
	}
	selected := make([]int, 0)
	attrs := strings.Split(rang, ",")
	for _, attr := range attrs {
		if strings.Contains(attr, "-") {
			bounds := strings.Split(attr, "-")
			if len(bounds) > 2 {
				panic("It is only permitted to establish a lower bound and an upper bound")
			}
			lowBound, err1 := strconv.ParseInt(bounds[0], 10, 0)
			upBound, err2 := strconv.ParseInt(bounds[1], 10, 0)
			if err1 != nil || err2 != nil {
				panic(fmt.Errorf("Make sure the bound %s is correctly defined, allow nummber-number only", attr))
			}
			lowBound = lowBound - 1
			upBound = upBound - 1
			for lowBound <= upBound {
				selected = append(selected, int(lowBound))
				lowBound++
			}

		} else {
			index, err := strconv.ParseInt(attr, 10, 0)
			if err != nil {
				panic(fmt.Errorf("Only numbers allow in %s ", attr))
			}
			index = index - 1
			selected = append(selected, int(index))
		}
		sort.Ints(selected)
		ntb.Columns = selected
		ntb.IsRangeInUse = true
	}
}

func (m *NominalToBinary) SetInvertSelection(set bool) {
	m.InvertSel = set
}

func (m *NominalToBinary) GetSelectedAttributes(numAttributes int) {
	if !m.InvertSel {
		m.SelectedAttributes = m.Columns
	} else {
		//a very costly implementation, **must be changed in the future**
		for j := range m.Columns {
			for i := 0; i < numAttributes; i++ {
				if j != i {
					contains := func() bool {
						for _, k := range m.SelectedAttributes {
							if i == k {
								return true
							}
						}
						return false
					}
					if !contains() {
						m.SelectedAttributes = append(m.SelectedAttributes, i)
					}
				} else {
					break
				}
			}
		}
	}
	m.Columns = m.SelectedAttributes
}

func (m *NominalToBinary) OutputAll() data.Instances {
	return m.Output_
}

func (m *NominalToBinary) Output() data.Instance {
	if !m.OutputQueue.IsEmpty() {
		if result, ok := m.OutputQueue.Pop().(data.Instance); ok {
			return result
		}
	}
	return data.NewInstance()
}

// This method does the function of calling in weka ConvertInstance(Instance) and then Output_()
// due in this implementation does not exists the m_OutputQueue value
func (m *NominalToBinary) ConvertAndReturn(instance data.Instance) data.Instance {
	return m.ConvertInstance(instance)
}
