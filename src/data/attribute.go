package data

import (
	"fmt"
)

const (
	NOMINAL = 0
	INTEGER = 1
	REAL    = 2
	NUMERIC = 3
	STRING  = 4
	INPUT   = 5
	OUTPUT  = 6
)

//it can be implemented as constants
type Attribute struct {
	//Constants
	Arff_Attribute, Arff_String, Arff_Integer, Arff_Real, Arff_Numeric string
	//Attribute type
	Attr_type int
	//If attribute is input or output, 0 input, 1 output
	direction int
	//Values that a nominal or string attribute can hold
	Values_ []string
	//Mapping of Values_ to indexes, weka's m_Hashtable
	ValuesIndexes_ map[string]int
	//Minimun and Maximun Values_ for the attribute
	Min_, Max_ float64
	//If Max_ and Min_ Values_ are defined in the attribute's definition
	HasFixedBounds_ bool
	//Attribute's Name_
	Name_ string
	//Attribute's Weight_
	Weight_ float64
	//Attribute index
	index int
}

//Attribute's struct constructor
func NewAttribute() Attribute {
	var attr Attribute
	//attr := new(Attribute)
	attr.Arff_Attribute = "@attribute"
	attr.Arff_Integer = "integer"
	attr.Arff_Numeric = "numeric"
	attr.Arff_Real = "real"
	attr.Arff_String = "string"
	attr.ValuesIndexes_ = make(map[string]int, 0)
	attr.Values_ = make([]string, 0)
	attr.Values_ = append(attr.Values_,"*DUMMY*STRING*FOR*STRING*ATTRIBUTES*")
	return attr
}

func NewAttributeWithName(Name_ string) Attribute {
	var attr Attribute
	//attr := new(Attribute)
	attr.Arff_Attribute = "@attribute"
	attr.Arff_Integer = "integer"
	attr.Arff_Numeric = "numeric"
	attr.Arff_Real = "real"
	attr.Arff_String = "string"
	attr.ValuesIndexes_ = make(map[string]int, 0)
	attr.Values_ = make([]string, 0)
	attr.Values_ = append(attr.Values_,"*DUMMY*STRING*FOR*STRING*ATTRIBUTES*")
	attr.Name_ = Name_
	return attr
}

func (a *Attribute) AddStringValue(val string) int {
	index, present := a.ValuesIndexes_[val]
	if present {
		//fmt.Println("in index")
		return index
	} else {
		//fmt.Println("in len")
		index := len(a.Values_)
		a.Values_ = append(a.Values_, val)
		a.ValuesIndexes_[val] = index
		return index
	}
}

func (a *Attribute) NumValues() int {
	if !a.IsNominal() && !a.IsString() {
		return 0
	}
	//fmt.Println(len(a.Values_), "Values_")
	fmt.Print()
	return len(a.Values_)
}

func (a Attribute) Value(valindex int) string {
	if !a.IsNominal() && !a.IsString() {
		return ""
	} else {
		return a.Values_[valindex]
	}
}

func (a Attribute) IsString() bool {
	return a.Attr_type == STRING
}

func (a Attribute) IsNominal() bool {
	return a.Attr_type == NOMINAL
}

func (a Attribute) IsNumeric() bool {
	return a.Attr_type == NUMERIC
}

//Sets methods

func (a *Attribute) SetIndex(index int) {
	a.index = index
}

func (a *Attribute) SetWeight(Weight_ float64) {
	a.Weight_ = Weight_
}

func (a *Attribute) SetDirection(direction int) {
	a.direction = direction
}

func (a *Attribute) SetMin(Min_ float64) {
	a.Min_ = Min_
}

func (a *Attribute) SetMax(Max_ float64) {
	a.Max_ = Max_
}

func (a *Attribute) SetName(Name_ string) {
	a.Name_ = Name_
}

func (a *Attribute) SetHasFixedBounds(hfb bool) {
	a.HasFixedBounds_ = hfb
}

func (a *Attribute) SetType(a_type int) {
	a.Attr_type = a_type
}

func (a *Attribute) SetValuesIndexes(ValuesIndexes_ map[string]int) {
	a.ValuesIndexes_ = ValuesIndexes_
}

func (a *Attribute) SetValues(Values_ []string) {
	a.Values_ = Values_
}

//Gets methods

func (a *Attribute) Index() int {
	return a.index
}

func (a *Attribute) Type() int {
	return a.Attr_type
}

func (a *Attribute) Weight() float64 {
	return a.Weight_
}

func (a *Attribute) Min() float64 {
	return a.Min_
}

func (a *Attribute) Max() float64 {
	return a.Max_
}

func (a *Attribute) Name() string {
	return a.Name_
}

func (a *Attribute) HasFixedBounds() bool {
	return a.HasFixedBounds_
}

func (a *Attribute) Direction() int {
	return a.direction
}

func (a *Attribute) Values() []string {
	return a.Values_
}

func (a *Attribute) ValuesIndexes() map[string]int {
	return a.ValuesIndexes_
}
