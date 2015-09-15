package data

import (
	"math"
	"fmt"
)

//this struct is like a mix of weka's Instance and SparseInstance with my own implementation
type Instance struct {
	//Constants
	MissingValue          float64
	Input_Att, Output_Att int
	//All values of the instance's attributes
	values []string
	//Stores only the nominal or strings attributes' values
	//nominalStringValues[0] = input attributes, nominalStringValues[1] = output attributes
	//nominalStringValues [][]string
	//Stores the nominal or string attributes to the corresponding indexes in nominalStringValues
	//intNominalStringValues[0] = input attributes, intNominalStringValues[1] = output attributes
	//intNominalStringValues [][]int
	//Stores the instance's attribute values //m_AttValues
	realValues []float64
	//Stores the missing values
	//missingValues [][]bool
	//Number of input attributes
	//numInputAttrs int
	//NUmber of output attributes
	//numOutputAttrs int
	//Stores the instance's attribute values
	//floatValues []float64
	//Instance's weight
	weight float64
	//The index of the attribute associated with each stored value
	indices []int
	//The maximum number of values that can be stored
	numAttributes int
}

func NewInstance() Instance {
	var inst Instance
	inst.MissingValue = math.NaN()
	inst.Input_Att = 0
	inst.Output_Att = 1
	inst.values = []string{}
	//	inst.nominalStringValues = make([][]string, 2)
	//	inst.intNominalStringValues = make([][]int, 2)
	inst.realValues = []float64{}
	//	inst.missingValues = make([][]bool, 2)
	//	inst.numInputAttrs = 0
	//	inst.numOutputAttrs = 0
	inst.weight = 0.0
	inst.indices = make([]int, 0)
	inst.numAttributes = 0
	return inst
}

func NewInstanceWeightValues(weight float64, values []float64) Instance {
	var inst Instance
	inst.MissingValue = math.NaN()
	inst.Input_Att = 0
	inst.Output_Att = 1
	inst.weight = weight
	inst.realValues = values
	return inst
}

func NewSparseInstance(weight float64, vals []float64, atts []Attribute) Instance {
	var inst Instance
	inst.MissingValue = math.NaN()
	inst.Input_Att = 0
	inst.Output_Att = 1
	inst.weight = weight

	tmpValues := make([]float64, 0)
	tmpInd := make([]int, 0)
	for i, v := range vals {
		if v != 0 {
			tmpValues = append(tmpValues, v)
			tmpInd = append(tmpInd, i)
		}
	}
	for k, j := range tmpInd {
		if atts[j].IsNominal() {
			if math.IsNaN(tmpValues[k]) {
				inst.AddValues("?")
			} else {
				inst.AddValues(atts[j].Values()[int(tmpValues[k])])
			}
		} else if atts[j].IsNominal() && !atts[j].IsString() {
			inst.AddValues(atts[j].Values()[j])
		} else {
			inst.AddValues(atts[j].Name())
		}
	}
	inst.realValues = tmpValues
	inst.indices = tmpInd
	inst.numAttributes = len(vals)
	return inst
}

func NewSparseInstanceWithIndexes(weight float64, tmpValues []float64, tmpInd []int, atts []Attribute) Instance {
	var inst Instance
	inst.MissingValue = math.NaN()
	inst.Input_Att = 0
	inst.Output_Att = 1
	inst.weight = weight

	for k, j := range tmpInd {
		if atts[j].IsNominal() {
			if math.IsNaN(tmpValues[k]) {
				inst.AddValues("?")
			} else {
				inst.AddValues(atts[j].Values()[int(tmpValues[k])])
			}
		} else if atts[j].IsNominal() && !atts[j].IsString() {
			inst.AddValues(atts[j].Values()[j])
		} else {
			inst.AddValues(atts[j].Name())
		}
	}
	inst.realValues = tmpValues
	inst.indices = tmpInd
	inst.numAttributes = len(atts)
	return inst
}

func (i *Instance) Index(idx int) int {
	return i.indices[idx]
}

func (i *Instance) ValueSparse(idx int) float64 {
	return i.realValues[idx]
}

//for sparse instances only
func (i *Instance) ClassValue(classIndex int) float64 {
	if classIndex < 0 {
		panic("Class is not set")
	}
	index := i.findIndex(classIndex)
	if (index >= 0) && (i.indices[index] == classIndex) {
		return i.realValues[index]
	}
	fmt.Print()
	return 0.0
}

func (i *Instance) ClassValueNotSparse(classIndex int) float64 {
	if classIndex < 0 {
		panic("Class is not set")
	}
	return i.realValues[classIndex]
}


func (i *Instance) Value(idx int) float64 {
	index := i.findIndex(idx)
	if (index >= 0) && (i.indices[index] == idx) {
		return i.realValues[index]
	}
	return 0.0
}

func (i *Instance) ClassMissing(classIndex int) bool {
	if classIndex < 0 {
		panic("Class is not set")
	}
	return i.IsMissingValue(classIndex)
}

func (i *Instance) IsMissingValue(idx int) bool {
	return math.IsNaN(i.Value(idx))
}

func (i *Instance) IsMissingSparse(idx int) bool {
	return math.IsNaN(i.realValues[idx])
}

func (i *Instance) AddValues(value string) {
	i.values = append(i.values, value)
	//fmt.Println(i.values)
}

func (i *Instance) Float64Slice(numAtt int) []float64 {
	newvalues := make([]float64, numAtt)
	for j:=range i.realValues {
		newvalues[i.indices[j]] = i.realValues[j]
	}
	return newvalues
}

//for sparse instances only
func (i *Instance) findIndex(index int) int {
	min := 0
	max := len(i.indices) - 1
	if max == -1 {
		return -1
	}
	//Binary search
	for (i.indices[min] <= index) && (i.indices[max] >= index) {
		current := (max + min) / 2
		if i.indices[current] > index {
			max = current - 1
		} else if i.indices[current] < index {
			min = current + 1
		} else {
			return current
		}
	}
	if i.indices[max] < index {
		return max
	} else {
		return min - 1
	}
}

func (i *Instance) SparseInstance(weight float64, values []float64, indices []int, maxValues int, atts []Attribute) {

	for j := 0; j < len(values); j++ {
		if values[j] != 0 {
			i.realValues = append(i.realValues, values[j])
			i.indices = append(i.indices, indices[j])
		}
	}
	for k, j := range indices {
		if atts[j].IsNominal() {
			if math.IsNaN(values[k]) {
				i.AddValues("?")
			} else {
				i.AddValues(atts[j].Values()[int(values[k])])
			}
		} else if atts[j].IsNominal() && !atts[j].IsString() {
			i.AddValues(atts[j].Values()[j])
		} else {
			i.AddValues(atts[j].Name())
		}
	}
		
	i.SetWeight(weight)
	i.numAttributes = maxValues
//	fmt.Println(i.Indices())	
//		fmt.Println(i.RealValues(),"iyiyiiyiyiyiy", i.NumAttributes())
}

func (i *Instance) SetClassMissing(classIndex int) {
	if classIndex < 0 {
		panic("Class is not set!")
	}
	i.SetMissing(classIndex)
}

func (i *Instance) SetMissing(attIndex int) {
	i.SetValue(attIndex, math.NaN())
}

func (i *Instance) SetValue(attIndex int, value float64) {
	i.freshAttributeVector()
	i.realValues[attIndex] = value
}

/**
 * Clones the attribute vector of the instance and
 * overwrites it with the clone.
 */
func (i *Instance) freshAttributeVector() {
	i.realValues = i.ToFloat64Slice()
}

// Returns the values of each attribute as an array of float64.
func (i *Instance) ToFloat64Slice() []float64 {
	newValues := make([]float64, len(i.realValues))
	copy(newValues, i.realValues)
	return newValues
}

//func (i *Instance) AddValuesWithIndex(idx int, value string) {
//	i.values[idx] = value
//}
//
//func (i *Instance) AddNominalStringValues(direction int, idx int, value string) {
//	if len(i.nominalStringValues[direction]) <= 0 {
//		i.nominalStringValues[direction] = make([]string, idx)
//	}
//	i.nominalStringValues[direction][idx] = value
//}
//
//func (i *Instance) AddIntNominalStringValues(direction int, idx int, value int) {
//	i.intNominalStringValues[direction][idx] = value
//}

func (i *Instance) AddRealValues(value float64) {
	i.realValues = append(i.realValues, value)
}

func (i *Instance) AddRealValuesIndex(idx int, value float64) {
	i.realValues[idx] = value
}

//func (i *Instance) AddMissingValues(direction int, idx int, value bool) {
//	i.missingValues[direction][idx] = value
//}

func (i *Instance) AddIndices(idx int) {
	i.indices = append(i.indices, idx)
}

//Gets methods

func (i *Instance) Indices() []int {
	return i.indices
}

func (i *Instance) NumAttributes() int {
	if i.numAttributes > len(i.realValues) {
		return len(i.realValues)
	}
	return i.numAttributes
}

func (i *Instance) NumAttributesSparse() int {
	return i.numAttributes
}

func (i *Instance) NumAttributesTest() int {
	
	
	return i.numAttributes
}

func (i *Instance) Values() []string {
	return i.values
}

//func (i *Instance) NominalStringValues() [][]string {
//	return i.nominalStringValues
//}
//
//func (i *Instance) IntNominalStringValues() [][]int {
//	return i.intNominalStringValues
//}
//
func (i *Instance) RealValues() []float64 {
	return i.realValues
}

//
//func (i *Instance) MissingValues() [][]bool {
//	return i.missingValues
//}
//
//func (i *Instance) NumOutputAttrs() int {
//	return i.numOutputAttrs
//}
//
//func (i *Instance) NumInputAttrs() int {
//	return i.numInputAttrs
//}

func (i *Instance) Weight() float64 {
	return i.weight
}

//Sets methods

func (i *Instance) SetIndices(indices []int) {
	i.indices = indices
}

func (i *Instance) SetNumAttributes(numAttrs int) {
	i.numAttributes = numAttrs
}

func (i *Instance) SetValues(values []string) {
	i.values = values
}

//func (i *Instance) SetNominalStringValues(nsv [][]string) {
//	i.nominalStringValues = nsv
//}
//
//func (i *Instance) SetIntNominalStringValues(insv [][]int) {
//	i.intNominalStringValues = insv
//}
//
func (i *Instance) SetRealValues(realValues []float64) {
	i.realValues = realValues
}

//
//func (i *Instance) SetMissingValues(missing [][]bool) {
//	i.missingValues = missing
//}
//
//func (i *Instance) SetNumOutputAttrs(noa int) {
//	i.numOutputAttrs = noa
//}
//
//func (i *Instance) SetNumInputAttrs(niu int) {
//	i.numInputAttrs = niu
//}

func (i *Instance) SetWeight(weight float64) {
	i.weight = weight
}
