package data

import (
	"math"
	//"fmt"
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

func (i *Instance) IsMissingValue(idx int) bool {
	return i.realValues[idx] == math.NaN()
}

func (i *Instance) AddValues(value string) {
	i.values = append(i.values, value)
	//fmt.Println(i.values)
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
