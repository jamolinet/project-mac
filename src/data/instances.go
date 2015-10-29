package data

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/cosn/collections/set"
	"github.com/project-mac/src/utils"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

var _attributes = make([]Attribute, 0)

type Instances struct {
	//Dataset's name
	DatasetName_ string
	//The Instances_
	Instances_ []Instance
	//The Attributes_ info
	Attributes_ []Attribute
	//Class attribute's index
	ClassIndex_ int
}

func NewInstances() Instances {
	_attributes = make([]Attribute, 0)
	var inst Instances
	inst.Instances_ = make([]Instance, 0)
	return inst
}

func NewInstancesNameAttCap(name string, attInfor []Attribute, capacity int) Instances {
	_attributes = make([]Attribute, 0)
	var names set.S
	names.Init()
	nonUniqueNames := ""
	for _, att := range attInfor {
		if names.Has(att.Name()) {
			nonUniqueNames += "'" + att.Name() + "'"
		}
		names.Add(att.Name())
	}
	if names.Len() != len(attInfor) {
		panic("Attribute names are not unique! Causes: " + nonUniqueNames)
	}
	names.Clear()
	var inst Instances
	inst.DatasetName_ = name
	inst.ClassIndex_ = -1
	inst.Attributes_ = attInfor
	for i := 0; i < inst.NumAttributes(); i++ {
		inst.Attribute(i).SetIndex(i)
	}
	inst.Instances_ = make([]Instance, capacity)
	return inst
}

func NewInstancesWithInst(inst Instances, capacity int) Instances {
	_attributes = make([]Attribute, 0)
	var i Instances
	if capacity < 0 {
		capacity = 0
	}
	i.Instances_ = make([]Instance, 0, capacity)
	if inst.ClassIndex() >= 0 {
		i.ClassIndex_ = inst.ClassIndex()
	}
	i.DatasetName_ = inst.DatasetName()
	i.Attributes_ = inst.Attributes()
	return i
}

func NewInstancesWithClassIndex(ClassIndex_ int) Instances {
	_attributes = make([]Attribute, 0)
	var inst Instances
	inst.Instances_ = make([]Instance, 0)
	inst.ClassIndex_ = ClassIndex_
	return inst
}

//Changed
func (i *Instances) Attribute(idx int) *Attribute {
	return &i.Attributes_[idx]
}

func (i Instances) AttributeNoPTR(idx int) Attribute {
	return i.Attributes_[idx]
}

func (i *Instances) Instance(idx int) *Instance {
	return &i.Instances_[idx]
}

func (i Instances) InstanceNoPtr(idx int) Instance {
	return i.Instances_[idx]
}

//Parse file dataset
func (inst *Instances) ParseFile(filepath string) error {
	err := inst.processHeader(filepath)
	if err != nil {
		return err
	}
	err = inst.parseInstances(filepath)
	if err != nil {
		return err
	}
	return nil
}

// Calculates summary statistics on the values that appear in this set of Instances_ for a specified attribute
func (i *Instances) AttributeStats(index int) AttributeStats {
	result := NewAttributeStats()
	if i.Attribute(index).IsNominal() {
		result.NominalCounts = make([]int, i.Attribute(index).NumValues())
	}
	if i.Attribute(index).IsNumeric() {
		result.numericStats = NewStats()
	}
	result.TotalCount = i.NumInstances()

	attVals := i.attributeToDoubleArray(index)
	sorted := utils.SortFloat(attVals)
	currentCount := 0
	prev := math.NaN() //for the moment is the missing value
	for j := 0; j < i.NumInstances(); j++ {
		current := *i.Instance(sorted[j])
		if current.IsMissingValue(index) {
			result.MissingCount = i.NumInstances() - j
			break
		}
		if current.RealValues()[index] == prev {
			currentCount++
		} else {
			result.AddDistinct(prev, currentCount)
			currentCount = 1
			prev = current.RealValues()[index]
		}
	}
	result.AddDistinct(prev, currentCount)
	result.DistinctCount-- // So we don't count "missing" as a value
	return result
}

// Gets the value of all Instances_ in this dataset for a particular attribute
func (i *Instances) attributeToDoubleArray(index int) []float64 {
	result := make([]float64, i.NumInstances())
	for j := range result {
		result[j] = i.Instances_[j].Value(index)
	}
	return result
}

//Process att
func (inst *Instances) processHeader(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Println("Processing Attributes_ definition in file: " + file.Name())
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
				fmt.Println("Reading ", "@attribute", " field")
				inst.parseAttribute(line, attrIndex)
				attrIndex++
			}
		}
		inst.Attributes_ = _attributes
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
		if attrIndex == inst.ClassIndex_ {
			attr.SetDirection(1)
		} else {
			attr.SetDirection(0)
		}
	} else if attr_type == attr.Arff_String {
		//Parse string attribute
		//fmt.Println("parsing string attribute")
		attr.SetType(STRING)
		attr.SetIndex(attrIndex)
		if attrIndex == inst.ClassIndex_ {
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
		if attrIndex == inst.ClassIndex_ {
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
	//fmt.Println("Parsing Instances_.")
	reader := bufio.NewScanner(file)
	for reader.Scan() {
		line := reader.Text()
		//Go down the file until the beginning of Instances_ definitions
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
		//Attributes_' values first as strings
		buff := bytes.NewBuffer([]byte(line))
		csvReader := csv.NewReader(buff)
		attVals, err := csvReader.Read()
		if err != nil {
			panic(fmt.Errorf("Malformed instance in line %s", line))
		}
		if len(attVals) <= inst.ClassIndex() {
			panic("The number of Attributes_ in an instance can't be less than the class index number")
		}
		for x := 0; x < len(attVals); x++ {
			//for idx, val := range attVals {
			//fmt.Printf("Index: %d, Value: %s", idx, val)
			//fmt.Println()
			//attr := &inst.Attributes_[idx]
			attr := &inst.Attributes_[x]
			direction := attr.Direction()
			//instance.AddValues(val)
			//inst.readValue(attr, direction, val, idx, &instance)
			inst.readValue(attr, direction, attVals[x], x, &instance)
		}
		instance.SetWeight(1.0)
		instance.SetNumAttributes(len(instance.Values()))
		inst.Instances_ = append(inst.Instances_, instance)
		//fmt.Println(len(instance.RealValues()))
		//fmt.Println(instance.NumAttributes())
	}
	return nil
}

func (inst *Instances) readValue(attr *Attribute, direction int, val string, idx int, instance *Instance) {
	if strings.EqualFold(val, "<null>") || strings.EqualFold(val, "?") {

		//switch attr.Type() {
		//case NUMERIC:
		//		instance.AddValues(val)
		//		instance.AddRealValues(instance.MissingValue)
		//	break
		//case NOMINAL:
		//case STRING:
		//	instance.AddValues(val)
		//	instance.AddRealValues(instance.MissingValue)
		//	break
		//}
	} else {
		switch attr.Type() {
		case NUMERIC:
			value, _ := strconv.ParseFloat(val, 64)
			instance.AddValues(val)
			instance.AddRealValues(value)
			break
		case NOMINAL:
			indx := inst.Attributes_[idx].ValuesIndexes()[val]
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
	if numFolds > len(i.Instances_) {
		panic("The number of folds can't be greater than number of Instances_")
	}
	numInstForFold = len(i.Instances_) / numFolds
	if numFold < len(i.Instances_)%numFolds {
		numInstForFold++
		offset = numFold
	} else {
		offset = len(i.Instances_) % numFolds
	}
	train = NewInstancesWithInst(*i, len(i.Instances_)-numInstForFold)
	first = numFold*(len(i.Instances_)/numFolds) + offset
	i.copyInstances(0, &train, first)
	i.copyInstances(first+numInstForFold, &train, len(i.Instances_)-first-numInstForFold)
	train.Randomize(seed)
	return train
}

//Creates the training set for one fold of a cross-validation on the dataset
func (i Instances) TrainCVRand(numFolds, numFold int, random *rand.Rand) Instances {
	var numInstForFold, first, offset int
	var train Instances
	if numFolds < 2 {
		panic("The number of folds should be at least 2 or more.")
	}
	if numFolds > len(i.Instances_) {
		panic("The number of folds can't be greater than number of Instances_")
	}
	numInstForFold = len(i.Instances_) / numFolds
	if numFold < len(i.Instances_)%numFolds {
		numInstForFold++
		offset = numFold
	} else {
		offset = len(i.Instances_) % numFolds
	}
	train = NewInstancesWithInst(i, len(i.Instances_)-numInstForFold)
	first = numFold*(len(i.Instances_)/numFolds) + offset
	i.copyInstances(0, &train, first)
	i.copyInstances(first+numInstForFold, &train, len(i.Instances_)-first-numInstForFold)
	train.Randomizes(random)
	return train
}

func (i Instances) TestCV(numFolds, numFold int) Instances {
	var numInstForFold, first, offset int
	var test Instances
	if numFolds < 2 {
		panic("The number of folds should be at least 2 or more.")
	}
	if numFolds > len(i.Instances_) {
		panic("The number of folds can't be greater than number of Instances_")
	}
	numInstForFold = len(i.Instances_) / numFolds
	if numFold < len(i.Instances_)%numFolds {
		numInstForFold++
		offset = numFold
	} else {
		offset = len(i.Instances_) % numFolds
	}
	test = NewInstancesWithInst(i, numInstForFold)
	first = numFold*(i.NumInstances()/numFolds) + offset
	i.copyInstances(first, &test, numInstForFold)
	//i.copyInstances(first+numInstForFold, &train, len(i.Instances_)-first-numInstForFold)
	return test
}

//Copies Instances_ from one set to the end of another one
func (i Instances) copyInstances(from int, dest *Instances, num int) {
	for j := 0; j < num; j++ {
		data := *i.Instance(from + j)
		//fmt.Println(data.weight)
		dest.Instances_ = append(dest.Instances_, data)
	}
}

//Shuffles the Instances_ in the set so that they are ordered randomly
func (i *Instances) Randomize(seed int) {
	rand.Seed(int64(seed))
	for j := range i.Instances_ {
		i.swap(j, rand.Intn(j+1))
	}
}

func (i *Instances) Randomizes(random *rand.Rand) {
	for j := range i.Instances_ {
		i.swap(j, random.Intn(j+1))
	}
}

//Swaps two Instances_ in the set
func (i *Instances) swap(j, k int) {
	temp := i.Instances_[j]
	i.Instances_[j] = i.Instances_[k]
	i.Instances_[k] = temp
}

func (i *Instances) SumOfWeights() float64 {
	sum := 0.0
	for _, inst := range i.Instances() {
		sum += inst.Weight_
	}
	return sum
}

func (i *Instances) DeleteWithMissing(attIndex int) {
	newInstances := make([]Instance, 0)
	for j := 0; j < len(i.Instances()); j++ {
		if !i.Instance(j).IsMissingValue(attIndex) {
			newInstances = append(newInstances, i.Instances_[j])
			//			fmt.Println(i.Instances_[j].weight)
		}
	}
	i.Instances_ = newInstances
}

func (i *Instances) DeleteWithMissingClass() {
	if i.ClassIndex_ < 0 {
		panic("Class index is negative (not set)!")
	}
	i.DeleteWithMissing(i.ClassIndex_)
}

func (i *Instances) MeanOrMode(idx int) float64 {
	var result, found float64
	var counts []int
	if i.Attribute(idx).IsNumeric() {
		result, found = 0, 0
		for _, inst := range i.Instances() {
			if inst.IsMissingValue(idx) {
				found += inst.Weight()
				result += inst.Weight() * inst.Value(idx)
			}
		}
		if found <= 0 {
			return 0
		} else {
			return result / found
		}
	} else if i.Attribute(idx).IsNominal() {
		counts = make([]int, i.Attribute(idx).NumValues())
		for _, inst := range i.Instances() {
			if inst.IsMissingValue(idx) {
				counts[int(inst.Value(idx))] += int(inst.Weight())
			}
		}
		return float64(utils.MaxIndexInts(counts))
	} else {
		return 0
	}
}

func (i *Instances) Variance(idx int) float64 {
	var sum, sumSquared, sumOfWeights float64
	if i.Attribute(idx).IsNumeric() {
		for _, inst := range i.Instances() {
			if inst.IsMissingValue(idx) {
				sum += inst.Weight() * inst.Value(idx)
				sumSquared += inst.Weight() * inst.Value(idx) * inst.Value(idx)
				sumOfWeights += inst.Weight()
			}
		}
		if sumOfWeights <= 1 {
			return 0
		}
		result := (sumSquared - (sum * sum / sumOfWeights)) / (sumOfWeights - 1)
		// We don't like negative variance
		if result < 0 {
			return 0
		} else {
			return result
		}
	} else {
		panic("Attribute not numeric!!!!!!!!")
	}
}

func (i *Instances) Stratify(numFolds int) {
	if numFolds <= 1 {
		panic("Number of folds must be greater than 1")
	}
	if i.ClassIndex_ < 0 {
		panic("Class index is negative (not set)!")
	}

	if i.ClassAttribute().IsNominal() {
		// sort by class
		index := 1
		for index < i.NumInstances() {
			instance1 := *i.Instance(index - 1)
			//fmt.Println(instance1.RealValues(), "values instance 1")
			for j := index; j < i.NumInstances(); j++ {
				instance2 := *i.Instance(j)
				//fmt.Println(instance2.RealValues(), "values instance 2")
				if instance1.ValueSparse(i.ClassIndex_) == instance2.ValueSparse(i.ClassIndex_) || (instance1.ClassMissing(i.ClassIndex_) && instance2.ClassMissing(i.ClassIndex_)) {
					i.swap(index, j)
					index++
				}
			}
			index++
		}
		i.stratStep(numFolds)
	}
}

func (i *Instances) stratStep(numFolds int) {
	newVec := make([]Instance, 0)
	start := 0
	var j int

	// create stratified batch
	for len(newVec) < i.NumInstances() {
		j = start
		for j < i.NumInstances() {
			newVec = append(newVec, i.Instances_[j])
			j = j + numFolds
		}
		start++
	}
	i.Instances_ = newVec

}

func (i Instances) StringFreeStructure() Instances {
	newAtts := make([]Attribute, 0)
	for i, att := range i.Attributes() {
		if att.Type() == STRING {
			temp := NewAttributeWithName(att.Name())
			temp.SetIndex(i)
			temp.SetType(att.Type())
			newAtts = append(newAtts, temp)
		}
	}
	if len(newAtts) == 0 {
		return NewInstancesWithInst(i, 0)
	}
	atts := i.Attributes_
	for i := range newAtts {
		atts[newAtts[i].Index()] = newAtts[i]
	}
	result := NewInstancesWithInst(i, 0)
	result.Attributes_ = atts
	return result
}

func (i *Instances) Add(inst Instance) {
	i.Instances_ = append(i.Instances_, inst)
}

func (i *Instances) NumInstances() int {
	return len(i.Instances_)
}

func (i *Instances) NumAttributes() int {
	return len(i.Attributes())
}

func (i *Instances) NumClasses() int {
	if i.ClassIndex_ < 0 {
		panic("Class index is negative (not set)!")
	}
	att := i.ClassAttribute()
	if !att.IsNominal() {
		return 1
	} else {
		return att.NumValues()
	}
}

func (i *Instances) ClassAttribute() Attribute {
	if i.ClassIndex_ < 0 {
		panic("Class index is negative (not set)!")
	}
	//fmt.Println(i.Attributes_)
	return i.Attributes_[i.ClassIndex_]
}

//Gets methods

func (i *Instances) DatasetName() string {
	return i.DatasetName_
}

func (i *Instances) Instances() []Instance {
	return i.Instances_
}

func (i *Instances) Attributes() []Attribute {
	return i.Attributes_
}

func (i *Instances) ClassIndex() int {
	return i.ClassIndex_
}

//Sets methods

func (i *Instances) SetDatasetName(name string) {
	i.DatasetName_ = name
}

func (i *Instances) SetInstances(insts []Instance) {
	i.Instances_ = insts
}

func (i *Instances) SetAttributes(attrs []Attribute) {
	i.Attributes_ = attrs
}

func (i *Instances) SetClassIndex(ClassIndex_ int) {
	i.ClassIndex_ = ClassIndex_
}

func (i *Instances) SortAtt(att Attribute) {
	i.SortIndex(att.Index())
}

func (m *Instances) SortIndex(index int) {
	vals := make([]float64, m.NumInstances())
	for i := 0; i < len(vals); i++ {
		val := m.Instance(i).ValueSparse(index)
		if math.IsNaN(val) {
			vals[i] = math.MaxFloat64
		} else {
			vals[i] = val
		}
	}
	sortOrder := utils.SortFloatWithNoMissingValues(vals)
	backup := make([]Instance, len(vals))
	for i := range vals {
		backup[i] = m.InstanceNoPtr(i)
	}
	for i:=range vals {
		m.Instances_[i] = backup[sortOrder[i]]
	}
}
