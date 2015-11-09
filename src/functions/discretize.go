package functions

import (
	"fmt"
	"github.com/cosn/collections/queue"
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/utils"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Discretize struct {

	// Stores which Columns to act on
	Columns            []int
	SelectedAttributes []int
	IsRangeInUse       bool

	/** Store the current cutpoints */
	CutPoints [][]float64
	/** Output binary attributes for discretized attributes. */
	MakeBinary bool
	/** Use better encoding of split point for MDL. */
	UseBetterEncoding bool
	/** Use Kononenko's MDL criterion instead of Fayyad et al.'s */
	UseKononenko bool

	Input_, Output_ data.Instances
	FirstTime       bool
	InvertSel       bool
	OutputQueue     queue.Q
	IsNil           string
}

func NewDiscretize() Discretize {
	var dis Discretize
	dis.CutPoints = nil
	dis.MakeBinary = false
	dis.UseBetterEncoding = false
	dis.UseKononenko = false
	dis.IsRangeInUse = false
	dis.OutputQueue.Init()
	dis.IsNil = "no"
	dis.FirstTime = true
	dis.InvertSel = false
	dis.SetRange("all")
	return dis
}

func (m *Discretize) Exec() {

}

func (m *Discretize) BatchFinished() {

	if len(m.Input_.Instances()) == 0 {
		panic("No input instance format defined.")
	}

	if m.CutPoints == nil {
		m.calculateCutPoints()
		m.SetOutputFormat()

		//Convert pending input instances
		for _, instance := range m.Input_.Instances() {
			m.ConvertInstance(instance)
		}
	}
}

// Set the output format. Takes the currently defined cutpoints and
// m_InputFormat and calls setOutputFormat(Instances) appropriately.

func (m *Discretize) SetOutputFormat() {
	if m.CutPoints == nil {
		m.Output_ = data.NewInstances()
		m.OutputQueue.Init()
		return
	}

	attributes := make([]data.Attribute, 0)
	classIndex := m.Input_.ClassIndex()
	for i := 0; i < m.Input_.NumAttributes(); i++ {
		if m.IsInRange(i) && m.Input_.Attribute(i).IsNumeric() {
			if !m.MakeBinary {
				attribValues := make([]string, 1)
				if m.CutPoints[i] == nil {
					attribValues = append(attribValues, "'All'")
				} else {
					for j, cutPoint := range m.CutPoints[i] {
						if j == 0 {
							attribValues = append(attribValues, "'(-inf-"+utils.Float64ToStringNoWitdh(cutPoint, 6)+"]'")
						} else if j == len(m.CutPoints[i]) {
							attribValues = append(attribValues, "'("+utils.Float64ToStringNoWitdh(m.CutPoints[i][j-1], 6)+"-inf)'")
						} else {
							attribValues = append(attribValues, "'("+utils.Float64ToStringNoWitdh(m.CutPoints[i][j-1], 6)+"-"+utils.Float64ToStringNoWitdh(cutPoint, 6)+"]'")
						}
					}
				}
				newA := data.NewAttributeWithName(m.Input_.Attribute(i).Name())
				newA.SetValues(attribValues)
				newA.SetWeight(m.Input_.Attribute(i).Weight())
				newA.SetType(data.STRING)
				attributes = append(attributes, newA)
			} else {
				if m.CutPoints[i] == nil {
					attribValues := make([]string, 1)
					attribValues = append(attribValues, "'All'")
					newA := data.NewAttributeWithName(m.Input_.Attribute(i).Name())
					newA.SetValues(attribValues)
					newA.SetWeight(m.Input_.Attribute(i).Weight())
					newA.SetType(data.STRING)
					attributes = append(attributes, newA)
				} else {
					if i < m.Input_.ClassIndex() {
						classIndex += len(m.CutPoints[i]) - 1
					}
					for j, cutPoint := range m.CutPoints[i] {
						attribValues := make([]string, 2)
						attribValues = append(attribValues, "'(-inf-"+utils.Float64ToStringNoWitdh(cutPoint, 6)+"]'")
						attribValues = append(attribValues, "'("+utils.Float64ToStringNoWitdh(cutPoint, 6)+"-inf)'")
						newA := data.NewAttributeWithName(m.Input_.Attribute(i).Name() + "_" + fmt.Sprint(j+1))
						newA.SetValues(attribValues)
						newA.SetWeight(m.Input_.Attribute(i).Weight())
						newA.SetType(data.STRING)
						attributes = append(attributes, newA)
					}
				}
			}
		} else {
			attributes = append(attributes, m.Input_.AttributeNoPTR(i))
		}
	}
	outputFormat := data.NewInstancesNameAttCap(m.Input_.DatasetName(), attributes, 0)
	outputFormat.SetClassIndex(classIndex)
	m.Output_ = outputFormat
}

// Gets the cuts points for an attribute
func (m *Discretize) GetCutPoints(attributeIndex int) []float64 {
	if m.CutPoints == nil {
		return nil
	}
	return m.CutPoints[attributeIndex]
}

/** Generate the cutpoints for each attribute */

func (m *Discretize) calculateCutPoints() {

	var copyy data.Instances
	m.CutPoints = make([][]float64, m.Input_.NumAttributes())
	for i := m.Input_.NumAttributes() - 1; i >= 0; i-- {
		if m.IsInRange(i) && m.Input_.Attribute(i).IsNumeric() {
			//Use copy to preserve order
			if &copyy == nil {
				copyy = data.NewInstancesWithInst(m.Input_, m.Input_.NumInstances())
			}
			m.calculateCutPointsByMDL(i, copyy)
		}
	}
}

// Set cutpoints for a single attribute using MDL.
func (m *Discretize) calculateCutPointsByMDL(index int, data data.Instances) {

	//Sort instances
	data.SortAtt(data.AttributeNoPTR(index))

	//Find first instances that's missing
	firstMissing := data.NumInstances()
	for i := 0; i < data.NumInstances(); i++ {
		if data.Instance(i).IsMissingValue(index) {
			firstMissing = i
			break
		}
	}
	m.CutPoints[index] = m.CutPointsForSubset(data, index, 0, firstMissing)
}

// Test using Kononenko's MDL criterion.
func (m *Discretize) KononenkosMDL(priorCounts []float64, bestCounts [][]float64, numInstances float64, numCutPoints int) bool {
	var distPrior, instPrior, sum, before, after float64
	distAfter, instAfter := 0.0, 0.0
	var numClassesTotal int

	// Number of classes occuring in the set
	numClassesTotal = 0
	for i := range priorCounts {
		if priorCounts[i] > 0 {
			numClassesTotal++
		}
	}

	// Encode distribution prior to split
	distPrior = utils.Log2Binomial(numInstances+float64(numClassesTotal)-1, float64(numClassesTotal-1))

	// Encode instances prior to split.
	instPrior = utils.Log2Multinomial(numInstances, priorCounts)

	before = instPrior + distPrior

	// Encode distributions and instances after split.
	for i := range bestCounts {
		sum = utils.Sum(bestCounts[i])
		distAfter += utils.Log2Binomial(sum+float64(numClassesTotal)-1, float64(numClassesTotal)-1)
		instAfter += utils.Log2Multinomial(sum, bestCounts[i])
	}

	// Coding cost after split
	after = utils.Log2(float64(numCutPoints)) + distAfter + instAfter

	// Check if split is to be accepted
	return before > after
}

// Test using Fayyad and Irani's MDL criterion.
func (m *Discretize) FayyadAndIranisMDL(priorCounts []float64, bestCounts [][]float64, numInstances float64, numCutPoints int) bool {
	var priorEntropy, entropy, gain float64
	var entropyLeft, entropyRight, delta float64
	var numClassesTotal, numClassesLeft, numClassesRight int

	//Compute entropy before split
	priorEntropy = Entropy(priorCounts)

	//Compute entropy after split
	entropy = EntropyConditionedOnRows(bestCounts)

	//Compute information gain
	gain = priorEntropy - entropy

	//Number of classes occuring in the set
	numClassesTotal = 0
	for i := range priorCounts {
		if priorCounts[i] > 0 {
			numClassesTotal++
		}
	}

	//Number of classes occuring in the left subset
	numClassesLeft = 0
	for i := range bestCounts[0] {
		if bestCounts[0][i] > 0 {
			numClassesLeft++
		}
	}

	//Number of classes occuring in the right subset
	numClassesRight = 0
	for i := range bestCounts[1] {
		if bestCounts[1][i] > 0 {
			numClassesRight++
		}
	}

	//Entropy of the left and the right subsets
	entropyLeft = Entropy(bestCounts[0])
	entropyRight = Entropy(bestCounts[1])

	//Compute terms for MDL formula
	delta = utils.Log2(math.Pow(3, float64(numClassesTotal))-2) - ((float64(numClassesTotal) * priorEntropy) - (float64(numClassesRight) * entropyRight) - (float64(numClassesLeft) * entropyLeft))

	//Check if split is to be accepted
	return gain > (utils.Log2(float64(numCutPoints))+delta)/float64(numInstances)
}

// Selects cutpoints for sorted subset.
func (m *Discretize) CutPointsForSubset(instances data.Instances, attIndex, first, lastPlusOne int) []float64 {

	var counts, bestCounts [][]float64
	var priorCounts, left, right, cutPoints []float64
	currentCutPoint, bestCutPoint := -math.MaxFloat64, -1.0
	var currentEntropy, bestEntropy, priorEntropy, gain float64
	bestIndex, numCutPoints, numInstances := -1, 0, 0.0

	//Compute number of instances in set
	if (lastPlusOne - first) < 2 {
		return nil
	}

	//Compute class counts
	counts = make([][]float64, 2)
	for i := range counts {
		counts[i] = make([]float64, instances.NumClasses())
	}
	for i := first; i < lastPlusOne; i++ {
		numInstances += instances.Instance(i).Weight()
		counts[1][int(instances.Instance(i).ClassValue(instances.ClassIndex()))] += instances.Instance(i).Weight()
	}

	//Save prior counts
	priorCounts = make([]float64, instances.NumClasses())
	copy(priorCounts, counts[1])

	//Entropy of the full set
	priorEntropy = Entropy(priorCounts)
	bestEntropy = priorEntropy

	//Find best entropy
	bestCounts = make([][]float64, 2)
	for i := range bestCounts {
		bestCounts[i] = make([]float64, instances.NumClasses())
	}
	for i := first; i < (lastPlusOne - 1); i++ {
		counts[0][int(instances.Instance(i).ClassValue(instances.ClassIndex()))] += instances.Instance(i).Weight()
		counts[1][int(instances.Instance(i).ClassValue(instances.ClassIndex()))] -= instances.Instance(i).Weight()
		if instances.Instance(i).ValueSparse(attIndex) < instances.Instance(i+1).ValueSparse(attIndex) {
			currentCutPoint = (instances.Instance(i).ValueSparse(attIndex) + instances.Instance(i+1).ValueSparse(attIndex)) / 2.0
			currentEntropy = EntropyConditionedOnRows(counts)
			if currentEntropy < bestEntropy {
				bestCutPoint = currentCutPoint
				bestEntropy = currentEntropy
				bestIndex = i
				copy(bestCounts[0], counts[0])
				copy(bestCounts[1], counts[1])
			}
			numCutPoints++
		}
	}

	//Use worse encoding?
	if !m.UseBetterEncoding {
		numCutPoints = (lastPlusOne - first) - 1
	}

	//Checks if gain is zero
	gain = priorEntropy - bestEntropy
	if gain <= 0 {
		return nil
	}

	//Check if split is to be accepted
	if (m.UseKononenko && m.KononenkosMDL(priorCounts, bestCounts, numInstances, numCutPoints)) || (!m.UseKononenko && m.FayyadAndIranisMDL(priorCounts, bestCounts, numInstances, numCutPoints)) {

		//Select split points for the left and right subsets
		left = m.CutPointsForSubset(instances, attIndex, first, bestIndex+1)
		right = m.CutPointsForSubset(instances, attIndex, bestIndex+1, lastPlusOne)

		//Merge cutpoints an return them
		if left == nil && right == nil {
			cutPoints = make([]float64, 1)
			cutPoints[0] = bestCutPoint
		} else if right == nil {
			cutPoints = make([]float64, len(left)+1)
			copy(cutPoints, left)
			cutPoints[len(left)] = bestCutPoint
		} else if left == nil {
			cutPoints = make([]float64, len(right)+1)
			cutPoints[0] = bestCutPoint
			copy(cutPoints[1:], right)
		} else {
			cutPoints = make([]float64, len(left)+len(right)+1)
			copy(cutPoints, left)
			cutPoints[len(left)] = bestCutPoint
			copy(cutPoints[len(left)+1:], right)
		}

		return cutPoints
	} else {
		return nil
	}
}

func (m *Discretize) Input(instance data.Instance) {
	m.OutputQueue.Init()
	if m.CutPoints != nil {
		m.ConvertInstance(instance)
		return
	}

	m.BufferInput(instance)
}

func (m *Discretize) BufferInput(instance data.Instance) {
	m.Input_.Add(instance)
}

func (m *Discretize) SetInputFormat(data data.Instances) {
	m.Input_ = data
	m.GetSelectedAttributes(len(data.Attributes()))
	m.CutPoints = nil
}

func (m *Discretize) ConvertInstance(instance data.Instance) {
	index := 0
	vals := make([]float64, m.Output_.NumAttributes())
	// Copy and convert the values
	for i, att := range m.Input_.Attributes() {
		if m.IsInRange(i) && att.IsNumeric() {
			j := 0
			currentval := instance.Value(i)
			if m.CutPoints[i] == nil {
				if instance.IsMissingSparse(i) {
					vals[index] = math.NaN()
				} else {
					vals[index] = 0
				}
				index++
			} else {
				if !m.MakeBinary {
					if instance.IsMissingSparse(i) {
						vals[index] = math.NaN()
					} else {
						for ; j < len(m.CutPoints[i]); j++ {
							if currentval <= m.CutPoints[i][j] {
								break
							}
						}
						vals[index] = float64(j)
					}
					index++
				} else {
					for j = 0; j < len(m.CutPoints[i]); j++ {
						if instance.IsMissingSparse(i) {
							vals[index] = math.NaN()
						} else if currentval <= m.CutPoints[i][j] {
							vals[index] = 0
						} else {
							vals[index] = 1
						}
						index++
					}
				}
			}
		} else {
			vals[index] = instance.Value(i)
			index++
		}
	}
	inst := data.NewInstance()
	//Get values and indexes different to zero
	indices := make([]int, len(vals))
	values := make([]float64, len(vals))
	idx := 0
	for i, val := range vals {
		if val != 0 {
			values[idx] = vals[i]
			indices[idx] = i
			idx++
		}
	}
	
	inst = data.NewSparseInstance(instance.Weight(),vals,m.Input_.Attributes())
	inst.SetNumAttributes(len(values))
	m.Output_.Add(inst)
	m.OutputQueue.Push(inst)
	//println("pushed")
}

func (m *Discretize) SetRange(rang string) {
	if strings.EqualFold(rang, "") {
		panic("The range cannot be empty")
	}
	if strings.EqualFold(rang, "all") {
		m.IsRangeInUse = false
		return
	}
	selected := make([]int, 0)
	attrs := strings.Split(rang, ",")
	for _, attr := range attrs {
		if strings.Contains(attr, "-") {
			bounds := strings.Split(attr, "-")
			if len(bounds) > 2 {
				panic("It is only permitted to establish a lower bound and an upper bound")
			}
			lowBound, err1 := strconv.ParseInt(bounds[0], 10, 0)
			upBound, err2 := strconv.ParseInt(bounds[1], 10, 0)
			if err1 != nil || err2 != nil {
				panic(fmt.Errorf("Make sure the bound %s is correctly defined, allow nummber-number only", attr))
			}
			lowBound = lowBound - 1
			upBound = upBound - 1
			for lowBound <= upBound {
				selected = append(selected, int(lowBound))
				lowBound++
			}

		} else {
			index, err := strconv.ParseInt(attr, 10, 0)
			if err != nil {
				panic(fmt.Errorf("Only numbers allow in %s ", attr))
			}
			index = index - 1
			selected = append(selected, int(index))
		}
		sort.Ints(selected)
		m.Columns = selected
		m.IsRangeInUse = true
	}
}

func (m *Discretize) SetInvertSelection(set bool) {
	m.InvertSel = set
}

func (m *Discretize) GetSelectedAttributes(numAttributes int) {
	if !m.InvertSel {
		m.SelectedAttributes = m.Columns
	} else {
		//a very costly implementation, **must be changed in the future**
		for j := range m.Columns {
			for i := 0; i < numAttributes; i++ {
				if j != i {
					contains := func() bool {
						for _, k := range m.SelectedAttributes {
							if i == k {
								return true
							}
						}
						return false
					}
					if !contains() {
						m.SelectedAttributes = append(m.SelectedAttributes, i)
					}
				} else {
					break
				}
			}
		}
	}
	m.Columns = m.SelectedAttributes
}
func (r *Discretize) IsInRange(f int) bool {
	// omit the class from the evaluation
	//	if r.hasClass && r.classIndex == f {
	//		return true
	//	}
	//
	//	if !r.IsRangeInUse || len(r.notInclude) == 0 {
	//		return false
	//	}

	for _, sel := range r.Columns {
		if sel == f {
			return true
		}
	}

	return false
}
func (m *Discretize) SetUseBetterEncoding(val bool) {
	m.UseBetterEncoding = val
}


func (m *Discretize) Output() data.Instances {
	return m.Output_
}