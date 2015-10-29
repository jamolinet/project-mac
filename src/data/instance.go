package data

import (
	"math"
	"fmt"
)

//this struct is like a mix of weka's Instance and SparseInstance with my own implementation
type Instance struct {
	//Constants
	//MissingValue          float64
	Input_Att, Output_Att int
	//All Values_ of the instance's attributes
	Values_ []string
	//Stores only the nominal or strings attributes' Values_
	//nominalStringValues[0] = input attributes, nominalStringValues[1] = output attributes
	//nominalStringValues [][]string
	//Stores the nominal or string attributes to the corresponding indexes in nominalStringValues
	//intNominalStringValues[0] = input attributes, intNominalStringValues[1] = output attributes
	//intNominalStringValues [][]int
	//Stores the instance's attribute Values_ //m_AttValues
	RealValues_ []float64
	//Stores the missing Values_
	//missingValues [][]bool
	//Number of input attributes
	//numInputAttrs int
	//NUmber of output attributes
	//numOutputAttrs int
	//Stores the instance's attribute Values_
	//floatValues []float64
	//Instance's Weight_
	Weight_ float64
	//The index of the attribute associated with each stored value
	Indices_ []int
	//The maximum number of Values_ that can be stored
	NumAttributes_ int
}

func NewInstance() Instance {
	var inst Instance
	//inst.MissingValue = math.NaN()
	inst.Input_Att = 0
	inst.Output_Att = 1
	inst.Values_ = []string{}
	//	inst.nominalStringValues = make([][]string, 2)
	//	inst.intNominalStringValues = make([][]int, 2)
	inst.RealValues_ = []float64{}
	//	inst.missingValues = make([][]bool, 2)
	//	inst.numInputAttrs = 0
	//	inst.numOutputAttrs = 0
	inst.Weight_ = 0.0
	inst.Indices_ = make([]int, 0)
	inst.NumAttributes_ = 0
	return inst
}

func NewInstanceWeightValues(Weight_ float64, Values_ []float64) Instance {
	var inst Instance
	//inst.MissingValue = math.NaN()
	inst.Input_Att = 0
	inst.Output_Att = 1
	inst.Weight_ = Weight_
	inst.RealValues_ = Values_
	return inst
}

func NewSparseInstance(Weight_ float64, vals []float64, atts []Attribute) Instance {
	var inst Instance
	//inst.MissingValue = math.NaN()
	inst.Input_Att = 0
	inst.Output_Att = 1
	inst.Weight_ = Weight_

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
	inst.RealValues_ = tmpValues
	inst.Indices_ = tmpInd
	inst.NumAttributes_ = len(vals)
	return inst
}

func NewSparseInstanceWithIndexes(Weight_ float64, tmpValues []float64, tmpInd []int, atts []Attribute) Instance {
	var inst Instance
	//inst.MissingValue = math.NaN()
	inst.Input_Att = 0
	inst.Output_Att = 1
	inst.Weight_ = Weight_

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
	inst.RealValues_ = tmpValues
	inst.Indices_ = tmpInd
	inst.NumAttributes_ = len(atts)
	return inst
}

func (i *Instance) Index(idx int) int {
	return i.Indices_[idx]
}

func (i *Instance) ValueSparse(idx int) float64 {
	return i.RealValues_[idx]
}

//for sparse instances only
func (i *Instance) ClassValue(classIndex int) float64 {
	if classIndex < 0 {
		panic("Class is not set")
	}
	index := i.findIndex(classIndex)
	if (index >= 0) && (i.Indices_[index] == classIndex) {
		return i.RealValues_[index]
	}
	fmt.Print()
	return 0.0
}

func (i *Instance) ClassValueNotSparse(classIndex int) float64 {
	if classIndex < 0 {
		panic("Class is not set")
	}
	return i.RealValues_[classIndex]
}


func (i *Instance) Value(idx int) float64 {
	index := i.findIndex(idx)
	if (index >= 0) && (i.Indices_[index] == idx) {
		return i.RealValues_[index]
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
	return math.IsNaN(i.RealValues_[idx])
}

func (i *Instance) AddValues(value string) {
	i.Values_ = append(i.Values_, value)
	//fmt.Println(i.Values_)
}

func (i *Instance) Float64Slice(numAtt int) []float64 {
	newvalues := make([]float64, numAtt)
	for j:=range i.RealValues_ {
		newvalues[i.Indices_[j]] = i.RealValues_[j]
	}
	return newvalues
}

//for sparse instances only
func (i *Instance) findIndex(index int) int {
	min := 0
	max := len(i.Indices_) - 1
	if max == -1 {
		return -1
	}
	//Binary search
	for (i.Indices_[min] <= index) && (i.Indices_[max] >= index) {
		current := (max + min) / 2
		if i.Indices_[current] > index {
			max = current - 1
		} else if i.Indices_[current] < index {
			min = current + 1
		} else {
			return current
		}
	}
	if i.Indices_[max] < index {
		return max
	} else {
		return min - 1
	}
}

func (i *Instance) SparseInstance(Weight_ float64, Values_ []float64, Indices_ []int, maxValues int, atts []Attribute) {

	for j := 0; j < len(Values_); j++ {
		if Values_[j] != 0 {
			i.RealValues_ = append(i.RealValues_, Values_[j])
			i.Indices_ = append(i.Indices_, Indices_[j])
		}
	}
	for k, j := range Indices_ {
		if atts[j].IsNominal() {
			if math.IsNaN(Values_[k]) {
				i.AddValues("?")
			} else {
				i.AddValues(atts[j].Values()[int(Values_[k])])
			}
		} else if atts[j].IsNominal() && !atts[j].IsString() {
			i.AddValues(atts[j].Values()[j])
		} else {
			i.AddValues(atts[j].Name())
		}
	}
		
	i.SetWeight(Weight_)
	i.NumAttributes_ = maxValues
//	fmt.Println(i.Indices())	
//		fmt.Println(i.RealValues(),"iyiyiiyiyiyiy", i.NumAttributes())
}

func (i *Instance) SetClassMissing(classIndex int) {
	if classIndex < 0 {
		panic("Class is not set!")
	}
	//i.SetMissing(classIndex)
}

//func (i *Instance) SetMissing(attIndex int) {
//	i.SetValue(attIndex, math.NaN())
//}

func (i *Instance) SetValue(attIndex int, value float64) {
	i.freshAttributeVector()
	i.RealValues_[attIndex] = value
}

/**
 * Clones the attribute vector of the instance and
 * overwrites it with the clone.
 */
func (i *Instance) freshAttributeVector() {
	i.RealValues_ = i.ToFloat64Slice()
}

// Returns the Values_ of each attribute as an array of float64.
func (i *Instance) ToFloat64Slice() []float64 {
	newValues := make([]float64, len(i.RealValues_))
	copy(newValues, i.RealValues_)
	return newValues
}

//func (i *Instance) AddValuesWithIndex(idx int, value string) {
//	i.Values_[idx] = value
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
	i.RealValues_ = append(i.RealValues_, value)
}

func (i *Instance) AddRealValuesIndex(idx int, value float64) {
	i.RealValues_[idx] = value
}

//func (i *Instance) AddMissingValues(direction int, idx int, value bool) {
//	i.missingValues[direction][idx] = value
//}

func (i *Instance) AddIndices(idx int) {
	i.Indices_ = append(i.Indices_, idx)
}

//Gets methods

func (i *Instance) Indices() []int {
	return i.Indices_
}

func (i *Instance) NumAttributes() int {
	if i.NumAttributes_ > len(i.RealValues_) {
		return len(i.RealValues_)
	}
	return i.NumAttributes_
}

func (i *Instance) NumAttributesSparse() int {
	return i.NumAttributes_
}

func (i *Instance) NumAttributesTest() int {
	
	
	return i.NumAttributes_
}

func (i *Instance) Values() []string {
	return i.Values_
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
	return i.RealValues_
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
	return i.Weight_
}

//Sets methods

func (i *Instance) SetIndices(Indices_ []int) {
	i.Indices_ = Indices_
}

func (i *Instance) SetNumAttributes(numAttrs int) {
	i.NumAttributes_ = numAttrs
}

func (i *Instance) SetValues(Values_ []string) {
	i.Values_ = Values_
}

//func (i *Instance) SetNominalStringValues(nsv [][]string) {
//	i.nominalStringValues = nsv
//}
//
//func (i *Instance) SetIntNominalStringValues(insv [][]int) {
//	i.intNominalStringValues = insv
//}
//
func (i *Instance) SetRealValues(RealValues_ []float64) {
	i.RealValues_ = RealValues_
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

func (i *Instance) SetWeight(Weight_ float64) {
	i.Weight_ = Weight_
}
