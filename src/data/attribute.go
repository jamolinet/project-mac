package data

import (
	"fmt"
)

const (
	NOMINAL = 0
	INTEGER = 1
	REAL = 2
	NUMERIC = 3
	STRING = 4
	INPUT = 5
	OUTPUT = 6
)
//it can be implemented as constants
type Attribute struct {
	//Constants
	Arff_Attribute, Arff_String, Arff_Integer, Arff_Real, Arff_Numeric string
	//Attribute type
	attr_type int
	//If attribute is input or output, 0 input, 1 output
	direction int
	//Values that a nominal or string attribute can hold
	values []string
	//Mapping of values to indexes, weka's m_Hashtable
	valuesIndexes map[string]int
	//Minimun and Maximun values for the attribute
	min, max float64
	//If max and min values are defined in the attribute's definition
	hasFixedBounds bool
	//Attribute's name
	name string
	//Attribute's weight
	weight float64
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
	attr.valuesIndexes = make(map[string]int,0)
	attr.values = make([]string, 0)
	return attr
}

func (a *Attribute) AddStringValue(val string) int {
	index, present := a.valuesIndexes[val]
	if present {
		//fmt.Println("in index")
		return index
	} else {
		//fmt.Println("in len")
		index := len(a.values)
		a.values = append(a.values, val)
		a.valuesIndexes[val] = index
		return index
	}
}

func (a *Attribute) NumValues() int {
	if !a.IsNominal() && !a.IsString() {
		return 0
	}
	fmt.Println(len(a.values), "values")
	return len(a.values)
}

func (a *Attribute) IsString() bool {
	return a.attr_type == STRING
}

func (a *Attribute) IsNominal() bool {
	return a.attr_type == NOMINAL
}

//Sets methods

func(a *Attribute) SetIndex(index int) {
	a.index = index
}

func (a *Attribute) SetWeight(weight float64) {
	a.weight = weight
}

func (a *Attribute) SetDirection(direction int) {
	a.direction = direction
}

func (a *Attribute) SetMin(min float64) {
	a.min = min
}

func (a *Attribute) SetMax(max float64) {
	a.max = max
}

func (a *Attribute) SetName(name string) {
	a.name = name
}

func (a *Attribute) SetHasFixedBounds(hfb bool) {
	a.hasFixedBounds = hfb
}

func (a *Attribute) SetType(a_type int) {
	a.attr_type = a_type
}

func (a *Attribute) SetValuesIndexes(valuesIndexes map[string]int) {
	a.valuesIndexes = valuesIndexes
}

func (a *Attribute) SetValues(values []string) {
	a.values = values
}

//Gets methods

func (a *Attribute) Index() int {
	return a.index
}

func (a *Attribute) Type() int {
	return a.attr_type
}

func (a *Attribute) Weight() float64 {
	return a.weight
}

func (a *Attribute) Min() float64 {
	return a.min
}

func (a *Attribute) Max() float64 {
	return a.max
}

func (a *Attribute) Name() string {
	return a.name
}

func (a *Attribute) HasFixedBounds() bool {
	return a.hasFixedBounds
}

func (a *Attribute) Direction() int {
	return a.direction
}

func (a *Attribute) Values() []string {
	return a.values
}

func (a *Attribute) ValuesIndexes() map[string]int {
	return a.valuesIndexes
}