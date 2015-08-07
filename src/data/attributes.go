package data

import (

)

type Attributes struct {
	//List of all attributes
	attributes []Attribute
	//List of input attributes
	inputAttrs []Attribute
	//List of output attributes
	outputAttrs []Attribute
	//Constants
	hasNominal, hasInteger, hasReal, hasString, hasMissing bool
	//Name of the dataset
	relationName string
	//Total number of attributes
	totalAttrs int
}

//Attributes' struct constructor
func NewAttributes() Attributes {
	var attrs Attributes
	attrs.attributes = make([]Attribute,0)
	attrs.inputAttrs = make([]Attribute,0)
	attrs.outputAttrs = make([]Attribute,0)
	attrs.hasMissing, attrs.hasInteger, attrs.hasNominal, attrs.hasReal, attrs.hasString = 
	false, false, false, false, false
	attrs.totalAttrs = 0
	return attrs
}

//Add a new attribute and depending of it's direction adds it to inputAttrs or outputAttrs
//if is an input attribute or output attribute respectively
func (a *Attributes) AddAttribute(at Attribute) {
	a.attributes = append(a.attributes, at)
	if at.Direction() == INPUT {
		a.inputAttrs = append(a.inputAttrs, at)
	} else {
		a.outputAttrs = append(a.outputAttrs, at)
	}
}

//Gets methods

func (a *Attributes) Attributes() []Attribute {
	return a.attributes
}

func (a *Attributes) InputAttrs() []Attribute {
	return a.inputAttrs
}

func (a *Attributes) OutputAttrs() []Attribute {
	return a.outputAttrs
}

func (a *Attributes) RelationName() string {
	return a.relationName
}

func (a *Attributes) HasMissing() bool {
	return a.hasMissing
}

func (a *Attributes) HasReal() bool {
	return a.hasReal
}

func (a *Attributes) HasNominal() bool {
	return a.hasNominal
}

func (a *Attributes) HasInteger() bool {
	return a.hasInteger
}

func (a *Attributes) HasString() bool {
	return a.hasString
}

func (a *Attributes) TotalAttrs() int {
	return a.totalAttrs
}

//Sets methods

func (a *Attributes) SetAttributes(attrs []Attribute) {
	a.attributes = attrs
}

func (a *Attributes) SetInputAttrs(attrs []Attribute) {
	a.inputAttrs = attrs
}

func (a *Attributes) SetOutputAttrs(attrs []Attribute) {
	a.outputAttrs = attrs
}

func (a *Attributes) SetRelationName(name string) {
	a.relationName = name
}

func (a *Attributes) SetHasMissing(val bool) {
	a.hasMissing = val
}

func (a *Attributes) SetHasInteger(val bool) {
	a.hasInteger = val
}

func (a *Attributes) SetHasNominal(val bool) {
	a.hasNominal = val
}

func (a *Attributes) SetHasReal(val bool) {
	a.hasReal = val
}

func (a *Attributes) SetHasString(val bool) {
	a.hasString = val
}

func (a *Attributes) SetTotalAttrs(total int) {
	a.totalAttrs = total
}

