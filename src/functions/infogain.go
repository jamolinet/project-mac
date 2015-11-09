package functions

import (
	"fmt"
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/utils"
	"math"
)

type InfoGain struct {
	// Treat missing values as a seperate value
	MissingMerge bool
	//Just Binarize numeric attributes
	Binarize bool
	// The info gain for each attribute
	InfoGains []float64
}

func NewInfoGain() InfoGain {
	var ig InfoGain
	ig.MissingMerge = true
	ig.Binarize = false
	ig.InfoGains = make([]float64, 0)
	return ig
}

func (ig *InfoGain) BuildEvaluator(instances data.Instances) {
	classIndex := instances.ClassIndex()
	numInstances := len(instances.Instances())

	if ig.Binarize { //Binarize instances
		ntb := NewNumericToBinary()
		ntb.Exec(instances)
		instances = ntb.Output()
	} else { //discretize instances
		dis := NewDiscretize()
		dis.SetUseBetterEncoding(true)
		dis.SetRange("all")
		dis.SetInputFormat(instances)
		dis.BatchFinished()
		instances = dis.Output()
	}
	numClasses := instances.Attribute(classIndex).NumValues()
	// Reserve space and initialize counters
	counts := make([][][]float64, len(instances.Attributes())) //initialize first dimension
	for k := range instances.Attributes() {
		if k != classIndex {
			numValues := len(instances.Attributes()[k].Values())
			counts[k] = make([][]float64, numValues+1) //initialize second dimension
			for i := range counts[k] {
				counts[k][i] = make([]float64, numClasses+1) //initialize third dimension
			}
		}
	}
	// Initialize counters
	temp := make([]float64, numClasses+1)
	for k := 0; k < numInstances; k++ {
		inst := instances.Instance(k)
		if inst.ClassMissing(classIndex) { //check that class if the class is missing /*implement method to do that*/
			temp[numClasses] += inst.Weight()
		} else {
			fmt.Print()
			temp[int(inst.ClassValue(classIndex))] += inst.Weight() //get the index of the value of the class
		}
	}
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
//					fmt.Println(len(counts), "counts", inst.Index(i),inst.ValueSparse(i),inst.ClassValue(classIndex))
//					fmt.Println(counts[inst.Index(i)], "[inst.Index(i)]")
//					fmt.Println(inst.RealValues_)
					counts[inst.Index(i)][int(inst.ValueSparse(i))][int(inst.ClassValue(classIndex))] += inst.Weight()
					counts[inst.Index(i)][0][int(inst.ClassValue(classIndex))] -= inst.Weight()
				}
			}
		}
	}
	// distribute missing counts if required
	if ig.MissingMerge {
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
	ig.InfoGains = make([]float64, len(instances.Attributes()))
	for i := range instances.Attributes() {
		if i != classIndex {
			ig.InfoGains[i] = EntropyOverColumns(counts[i]) - EntropyConditionedOnRows(counts[i])
		}
	}
	//fmt.Println(ig.InfoGains, "infogain")
}

func Entropy(array []float64) float64 {
	returnValue, sum := 0.0, 0.0

	for i := range array {
		returnValue = returnValue - LnFunc(array[i])
		sum += array[i]
	}
	if utils.Eq(sum, 0) {
		return 0
	} else {
		return (returnValue + LnFunc(sum)) / (sum * math.Log(2))
	}
}

func EntropyOverColumns(matrix [][]float64) float64 {
	returnValue, total := 0.0, 0.0
	var sumForColumn float64
	for j := range matrix[0] {
		sumForColumn = 0
		for i := range matrix {
			sumForColumn += matrix[i][j]
		}
		returnValue = returnValue - LnFunc(sumForColumn)
		total += sumForColumn
	}
	if utils.Eq(total, 0) {
		return 0
	}
	return (returnValue + LnFunc(total)) / (total * math.Log(2))
}

func EntropyConditionedOnRows(matrix [][]float64) float64 {
	returnValue, total := 0.0, 0.0
	var sumForRow float64
	for i := range matrix {
		sumForRow = 0
		for j := range matrix[0] {
			returnValue = returnValue + LnFunc(matrix[i][j])
			sumForRow += matrix[i][j]
		}
		returnValue = returnValue - LnFunc(sumForRow)
		total += sumForRow
	}
	if utils.Eq(total, 0) {
		return 0
	}
	return -returnValue / (total * math.Log(2))
}

func LnFunc(num float64) float64 {
	if num <= 0 {
		return 0
	} else {
		return num * math.Log(num)
	}
}

func (ig *InfoGain) EvaluateAttribute(attr int) float64 {
	return ig.InfoGains[attr]
}

func (ig *InfoGain) SetMissingMerge(mm bool) {
	ig.MissingMerge = mm
}

func (ig *InfoGain) SetBinarize(Binarize bool) {
	ig.Binarize = Binarize
}

func (ig *InfoGain) InfoGains_() []float64 {
	return ig.InfoGains
}
