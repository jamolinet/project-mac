package functions

import (
	"github.com/project-mac/src/data"
	"math"
	"github.com/project-mac/src/utils"
	"fmt"
)

type InfoGain struct {
	// Treat missing values as a seperate value
	missingMerge bool
	//Just binarize numeric attributes
	binarize bool
	// The info gain for each attribute
	infoGains []float64
}

func NewInfoGain() InfoGain {
	var ig InfoGain
	ig.missingMerge = true
	ig.binarize = false
	ig.infoGains = make([]float64, 0)
	return ig
}

func (ig *InfoGain) BuildEvaluator(instances data.Instances) {
	classIndex := instances.ClassIndex()
	numInstances := len(instances.Instances())

	if ig.binarize { //binarize instances
		//implement NumericToBinary function
		ntb := NewNumericToBinary()
		ntb.Exec(instances)
		instances =  ntb.Output()
		//fmt.Println(instances.Instances())
	} else { //discretize instances
		//implement Discretize function
	}
	numClasses := instances.Attribute(classIndex).NumValues()
	// Reserve space and initialize counters
	counts := make([][][]float64, len(instances.Attributes())) //initialize first dimension
	for k := range instances.Attributes() {
		//fmt.Println(k)
		if k != classIndex {
			numValues := len(instances.Attributes()[k].Values())
			counts[k] = make([][]float64, numValues+1) //initialize second dimension
			for i := range counts[k] {
				counts[k][i] = make([]float64, numClasses+1) //initialize third dimension
			}
		}
	}
	// Initialize counters
	//fmt.Println(numClasses, "numclasses")
	temp := make([]float64, numClasses+1)
	for k := 0; k < numInstances; k++ {
		inst := instances.Instance(k)
		if inst.ClassMissing(classIndex) { //check that class if the class is missing /*implement method to do that*/
			temp[numClasses] += inst.Weight()
		} else {
			//fmt.Println(int(inst.ClassValue(classIndex)), "classIndexes", inst.Weight(), "weights")
			fmt.Print()
			temp[int(inst.ClassValue(classIndex))] += inst.Weight() //get the index of the value of the class
		}
	}
	//fmt.Println(temp)
	for k := range counts {
		if k != classIndex {
			for i := range temp {
				counts[k][0][i] = temp[i]
			}
		}
	}
	// Get counts
	//inst.RealValues()[classIndex]) check this after finish, may contains errors, its have to be check if the classIndex exists if not return 0 /*see weka*/
	//implement the necessary methods to make easier this implementation and not bugs friendly
	//New methods already implemented!!!!!!!! Later check it's functioning
	for k := 0; k < numInstances; k++ {
		inst := instances.Instance(k)
		for i := range inst.RealValues() {
			if inst.Index(i) != classIndex {
				if inst.IsMissingValue(i) || inst.ClassMissing(classIndex) { //if is missing the real value and the class
					if !inst.IsMissingValue(i) {
						counts[inst.Index(i)][int(inst.ValueSparse(i))][numClasses] += inst.Weight()
						counts[inst.Index(i)][0][numClasses] -= inst.Weight()
					} else if !inst.IsMissingValue(classIndex) {
						counts[inst.Index(i)][instances.Attribute(inst.Index(i)).NumValues()][int(inst.ClassValue(classIndex))] += inst.Weight() //tongue twister, now its not
						counts[inst.Index(i)][0][int(inst.ClassValue(classIndex))] -= inst.Weight()
					} else {
						counts[inst.Index(i)][instances.Attribute(inst.Index(i)).NumValues()][numClasses] += inst.Weight()
						counts[inst.Index(i)][0][numClasses] -= inst.Weight()
					}
				} else {
					counts[inst.Index(i)][int(inst.ValueSparse(i))][int(inst.ClassValue(classIndex))] += inst.Weight()
					counts[inst.Index(i)][0][int(inst.ClassValue(classIndex))] -= inst.Weight()
				}
			}
		}
	}
	// distribute missing counts if required
	if ig.missingMerge {
		for k := range instances.Attributes() {
			if k != classIndex {
				numValues := len(instances.Attributes()[k].Values())
				// Compute marginals
				rowSums := make([]float64, numValues)
				columnSums := make([]float64, numClasses)
				sum := 0.0
				for i := 0; i < numValues; i++ {
					for j := 0; j < numClasses; j++ {
						rowSums[i] += counts[k][i][j]
						columnSums[j] += counts[k][i][j]
					}
					sum += rowSums[i]
				}
				if utils.Gr(sum, 0) {
					additions := make([][]float64, numValues) //initializes slices
					for i := range additions {
						additions[i] = make([]float64, numClasses)
					}
					// Compute what needs to be added to each row
					for i := range additions {
						for j := range additions[i] {
							additions[i][j] = (rowSums[i] / sum) * counts[k][numValues][j]
						}
					}
					// Compute what needs to be added to each column
					for i := 0; i < numClasses; i++ {
						for j := 0; j < numValues; j++ {
							additions[j][i] += (columnSums[i] / sum) * counts[k][j][numClasses]
						}
					}
					// Compute what needs to be added to each cell
					for i := 0; i < numClasses; i++ {
						for j := 0; j < numValues; j++ {
							additions[j][i] += (counts[k][j][i] / sum) * counts[k][numValues][numClasses]
						}
					}
					// Make new contingency table
					newTable := make([][]float64, numValues) //initializes slices
					for i := range newTable {
						newTable[i] = make([]float64, numClasses)
					}
					for i := range newTable {
						for j := range newTable[i] {
							newTable[i][j] = counts[k][i][j] + additions[i][j]
						}
					}
					counts[k] = newTable
				}
			}
		}
	}
	// Compute info gains
	ig.infoGains = make([]float64, len(instances.Attributes()))
	for i := range instances.Attributes() {
		if i != classIndex {
			ig.infoGains[i] = entropyOverColumns(counts[i]) - entropyConditionedOnRows(counts[i])
		}
	}
	//fmt.Println(ig.infoGains, "infogain")
}

func entropyOverColumns(matrix [][]float64) float64 {
	returnValue, total := 0.0, 0.0
	var sumForColumn float64
	for j := range matrix[0] {
		sumForColumn = 0
		for i := range matrix {
			sumForColumn += matrix[i][j]
		}
		returnValue = returnValue - lnFunc(sumForColumn)
		total += sumForColumn
	}
	if utils.Eq(total, 0) {
		return 0
	}
	return (returnValue + lnFunc(total)) / (total * math.Log(2))
}

func entropyConditionedOnRows(matrix [][]float64) float64 {
	returnValue, total := 0.0, 0.0
	var sumForRow float64
	for i := range matrix {
		sumForRow = 0
		for j := range matrix[0] {
			returnValue = returnValue + lnFunc(matrix[i][j])
			sumForRow += matrix[i][j]
		}
		returnValue = returnValue - lnFunc(sumForRow)
		total += sumForRow
	}
	if utils.Eq(total, 0) {
		return 0
	}
	return -returnValue / (total * math.Log(2))
}

func lnFunc(num float64) float64 {
	if num <= 0 {
		return 0
	} else {
		return num * math.Log(num)
	}
}

func (ig *InfoGain) evaluateAttribute(attr int) float64 {
	return ig.infoGains[attr]
}

func (ig *InfoGain) SetMissingMerge(mm bool) {
	ig.missingMerge = mm
}

func (ig *InfoGain) SetBinarize(binarize bool) {
	ig.binarize = binarize
}

func (ig *InfoGain) InfoGains() []float64 {
	return ig.infoGains
}
