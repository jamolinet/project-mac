package data

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

var _attributes = make([]Attribute, 0)

type Instances struct {
	//Dataset's name
	datasetName string
	//The instances
	instances []Instance
	//The attributes info
	attributes []Attribute
	//Class attribute's index
	classIndex int
}

func NewInstances() Instances {
	var inst Instances
	inst.instances = make([]Instance, 0)
	return inst
}

func NewInstancesWithInst(inst Instances, capacity int) Instances {
	var i Instances
	i.instances = make([]Instance, 0, capacity)
	if inst.ClassIndex() >= 0 {
		i.classIndex = inst.ClassIndex()
	}
	i.datasetName = inst.DatasetName()
	i.attributes = inst.Attributes()
	return i
}

func NewInstancesWithClassIndex(classIndex int) Instances {
	var inst Instances
	inst.instances = make([]Instance, 0)
	inst.classIndex = classIndex
	return inst
}

func (i *Instances) Attribute(idx int) *Attribute {
	return &i.attributes[idx]
}

func (i *Instances) Instance(idx int) *Instance {
	return &i.instances[idx]
}

//Parse file dataset
func (inst *Instances) ParseFile(filepath string) error {
	inst.processHeader(filepath)
	inst.parseInstances(filepath)
	return nil
}

//Process att
func (inst *Instances) processHeader(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Println("Processing attributes definition in file: " + file.Name())
	reader := bufio.NewScanner(file)
	attrIndex := 0
	for reader.Scan() {
		line := reader.Text()
		if !strings.HasPrefix(line, "@data") {
			if strings.HasPrefix(line, "@relation") {
				//fmt.Println("Reading @relation field")
				inst.SetDatasetName(strings.TrimPrefix(line, "@relation "))
				inst.SetDatasetName(strings.TrimPrefix(line, "@relation "))
			} else if strings.HasPrefix(line, "@attribute") {
				//fmt.Println("Reading ", "@attribute", " field")
				inst.parseAttribute(line, attrIndex)
				attrIndex++
			}
		}
		inst.attributes = _attributes
		//var att Attribute
	}
	return nil
}

func (inst *Instances) parseAttribute(line string, attrIndex int) {
	fields := strings.Fields(strings.ToLower(strings.TrimPrefix(line, "@attribute")))
	//fmt.Println(len(fields))
	//fmt.Println(fields[0])
	if len(fields) < 2 {
		panic("Attribute's name is not defined, check your dataset.")
	}
	//	if len(fields) < 3 {
	//		panic("Attribute's type is not defined, check your dataset.")
	//	}
	name := fields[0]
	attr := NewAttribute()
	attr.SetName(name)
	attr_type := fields[1]
	if attr_type == attr.Arff_Integer || attr_type == attr.Arff_Numeric || attr_type == attr.Arff_Real {
		//Parse numeric attribute
		//fmt.Println("parsing numeric attribute")
		attr.SetType(NUMERIC)
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			//fmt.Println("parsing bounds")
			bounds := line[strings.Index(line, "[")+1 : strings.Index(line, "]")] //example: "[23, 89]", bounds = "23, 89"
			//fmt.Println("bounds: ", bounds)
			min := strings.TrimSpace(bounds[:strings.Index(bounds, ",")])
			//fmt.Println("min: ", min)
			max := strings.TrimSpace(bounds[strings.Index(bounds, ",")+1:])
			//fmt.Println("max: ", max)
			min_float, err := strconv.ParseFloat(min, 64)
			if err != nil {
				panic(fmt.Errorf("Impossible to cast from string to float, bad bounds declaration in min at line '%s'", line))
			}
			attr.SetMin(min_float)
			max_float, err := strconv.ParseFloat(max, 64)
			if err != nil {
				panic(fmt.Errorf("Impossible to cast from string to float, bad bounds declaration in max at line '%s'", line))
			}
			attr.SetMax(max_float)
			attr.SetHasFixedBounds(true)
		} else {
			attr.SetHasFixedBounds(false)
		}
		attr.SetIndex(attrIndex)
		if attrIndex == inst.classIndex {
			attr.SetDirection(1)
		} else {
			attr.SetDirection(0)
		}
	} else if attr_type == attr.Arff_String {
		//Parse string attribute
		//fmt.Println("parsing string attribute")
		attr.SetType(STRING)
		attr.SetIndex(attrIndex)
		if attrIndex == inst.classIndex {
			attr.SetDirection(1)
		} else {
			attr.SetDirection(0)
		}
	} else if strings.Contains(line, "{") && strings.Contains(line, "}") {
		//is nominal attribute
		//fmt.Println("parsing nominal attribute")
		attr.SetType(NOMINAL)
		attr.SetHasFixedBounds(true)
		nominalValues(line[strings.Index(line, "{")+1:strings.Index(line, "}")], &attr)
		if attrIndex == inst.classIndex {
			attr.SetDirection(1)
		} else {
			attr.SetDirection(0)
		}
	} else {
		panic(fmt.Errorf("Unsupported attribute type '%s' or bad nominal attribute definition", attr_type))
	}
	_attributes = append(_attributes, attr)
}

func nominalValues(line string, attr *Attribute) {
	//fmt.Println(line)
	line = strings.TrimSpace(strings.Replace(strings.Replace(line, " ", "", -1), ",", " ", -1))
	vals := strings.Fields(line)
	valuesIndexes := make(map[string]int)
	values := make([]string, len(vals))
	for index, value := range vals {
		valuesIndexes[value] = index
		values[index] = value
	}
	//fmt.Println(values)
	attr.SetValues(values)
	attr.SetValuesIndexes(valuesIndexes)
}

//func (i *Instances) populateAttributes() {
//	attrs := NewAttributes()
//	for _, attribute := range _attributes {
//		attrs.AddAttribute(attribute)
//	}
//	attrs.SetTotalAttrs(len(attrs.Attributes()))
//	i.SetAttributes(attrs)
//}

func (inst *Instances) parseInstances(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	//fmt.Println("Parsing instances.")
	reader := bufio.NewScanner(file)
	for reader.Scan() {
		line := reader.Text()
		//Go down the file until the beginning of instances definitions
		if strings.HasPrefix(line, "@data") {
			break
		}
	}
	for reader.Scan() {
		line := reader.Text()
		if len(line) == 0 {
			continue
		}
		//fmt.Println(line)
		instance := NewInstance()
		//make sure the instance is well-read
		line = strings.TrimSpace(line)
		//attributes' values first as strings
		buff := bytes.NewBuffer([]byte(line))
		csvReader := csv.NewReader(buff)
		attVals, err := csvReader.Read()
		if err != nil {
			panic(fmt.Errorf("Malformed instance in line %s", line))
		}
		for idx, val := range attVals {
			//fmt.Printf("Index: %d, Value: %s", idx, val)
			//fmt.Println()
			attr := &inst.attributes[idx]
			direction := attr.Direction()
			//instance.AddValues(val)
			inst.readValue(attr, direction, val, idx, &instance)
		}
		instance.SetWeight(1.0)
		instance.SetNumAttributes(len(instance.Values()))
		inst.instances = append(inst.instances, instance)
		//fmt.Println(line)
		fmt.Println(instance.NumAttributes())
	}
	return nil
}

func (inst *Instances) readValue(attr *Attribute, direction int, val string, idx int, instance *Instance) {
	if strings.EqualFold(val, "<null>") || strings.EqualFold(val, "?") {

		switch attr.Type() {
		case NUMERIC:
			instance.AddValues(val)
			instance.AddRealValues(instance.MissingValue)
			break
		case NOMINAL:
		case STRING:
			instance.AddValues(val)
			instance.AddRealValues(instance.MissingValue)
			break
		}
	} else {
		switch attr.Type() {
		case NUMERIC:
			value, _ := strconv.ParseFloat(val, 64)
			instance.AddValues(val)
			instance.AddRealValues(value)
			break
		case NOMINAL:
			indx := inst.attributes[idx].ValuesIndexes()[val]
			instance.AddRealValues(float64(indx))
			instance.AddValues(val)
			break
		case STRING:
			//fmt.Println(" String in readValue")
			val = strings.Trim(val, "'")
			instance.AddValues(val)
			instance.AddRealValues(float64(attr.AddStringValue(val)))
			//fmt.Println(val)
			break
		}
	}
	//fmt.Println(attr.Values())
}

//Creates the training set for one fold of a cross-validation on the dataset
func (i *Instances) TrainCV(numFolds, numFold, seed int) Instances {
	var numInstForFold, first, offset int
	var train Instances
	if numFolds < 2 {
		panic("The number of folds should be at least 2 or more.")
	}
	if numFolds > len(i.instances) {
		panic("The number of folds can't be greater than number of instances")
	}
	numInstForFold = len(i.instances) / numFolds
	if numFold < len(i.instances)%numFolds {
		numInstForFold++
		offset = numFold
	} else {
		offset = len(i.instances) % numFolds
	}
	train = NewInstancesWithInst(*i, len(i.instances)-numInstForFold)
	first = numFold*(len(i.instances)/numFolds) + offset
	i.copyInstances(0, &train, first)
	i.copyInstances(first+numInstForFold, &train, len(i.instances)-first-numInstForFold)
	train.Randomize(seed)
	return train
}

//Copies instances from one set to the end of another one
func (i *Instances) copyInstances(from int, dest *Instances, num int) {
	for j := 0; j < num; j++ {
		data := *i.Instance(from + j)
		dest.instances = append(dest.instances, data)
	}
}

//Shuffles the instances in the set so that they are ordered randomly
func (i *Instances) Randomize(seed int) {
	rand.Seed(int64(seed))
	for j := range i.instances {
		i.swap(j, rand.Intn(j+1))
	}
}

//Swaps two instances in the set
func (i *Instances) swap(j, k int) {
	temp := i.instances[j]
	i.instances[j] = i.instances[k]
	i.instances[k] = temp
}

//Gets methods

func (i *Instances) DatasetName() string {
	return i.datasetName
}

func (i *Instances) Instances() []Instance {
	return i.instances
}

func (i *Instances) Attributes() []Attribute {
	return i.attributes
}

func (i *Instances) ClassIndex() int {
	return i.classIndex
}

//Sets methods

func (i *Instances) SetDatasetName(name string) {
	i.datasetName = name
}

func (i *Instances) SetInstances(insts []Instance) {
	i.instances = insts
}

func (i *Instances) SetAttributes(attrs []Attribute) {
	i.attributes = attrs
}

func (i *Instances) SetClassIndex(classIndex int) {
	i.classIndex = classIndex
}

func (i *Instances) String() string {
	return fmt.Sprintf("fdfsffsfdfsffs+++++++++++++ %s", i.DatasetName())
}
