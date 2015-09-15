package functions

import (
	"github.com/project-mac/src/data"
	"math"
	//	"utils"
	"fmt"
)

type AttributeSelection struct {
	//the selected attributes
	selectedAttributeSet, selectedAttributes []int
	//the attribute indexes and associated merits if a ranking is produced
	attributeRanking [][]float64
	//number of attributes requested from ranked results,
	//the number of folds to use for cross validation
	numToSelect, numFolds int
	//
	attributeFilter Remove
	//rank features,
	//do cross validation
	doRank, doXval bool
	//the instances to select attributes from
	trainInstances data.Instances
	evaluator      InfoGain
	search         Ranker
	seed           int //used ofr randomize instances in crossvalidation
	//these are only for validation statistics, if not useful can be remove later
	rankResults   [][]float64
	subSetResults []float64
	trials        int

	input, output data.Instances
	hasClass      bool
}

func NewAttributeSelection() AttributeSelection {
	var as AttributeSelection
	return as
}

//Start the selection attributes process
func (as *AttributeSelection) StartSelection(instances data.Instances) {
	as.input = data.NewInstancesWithInst(instances, len(instances.Attributes()))
	as.input = instances
	as.output = data.NewInstances()
	as.hasClass = as.input.ClassIndex() >= 0
	as.selectedAttributes = as.SelectAttributes(as.input)
	if len(as.selectedAttributes) == 0 {
		panic("No selected attributes")
	}
	//Set output
	//fmt.Println(as.selectedAttributes, "as.selectedAttributes")
	as.output = data.NewInstances()
	attributes := make([]data.Attribute, 0)
	for i := range as.selectedAttributes {
		attributes = append(attributes, *as.input.Attribute(as.selectedAttributes[i]))
	}
	//fmt.Println(attributes, "attributes")
	as.output.SetDatasetName(as.input.DatasetName())
	as.output.SetAttributes(attributes)
	if as.hasClass {
		as.output.SetClassIndex(len(as.selectedAttributes) - 1)
	}
	// Convert pending input instances
	tmpInst := make([]data.Instance, 0)
	for _, in := range as.input.Instances() {
		tmpInst = append(tmpInst, as.convertInstance(in))
		
	}
	as.output.SetInstances(tmpInst)
}

//Convert a single instance over
func (as *AttributeSelection) convertInstance(inst data.Instance) data.Instance {
	newVasl := make([]float64, 0, len(as.output.Attributes()))
	//fmt.Println(cap(newVasl), "newVals")
	for _, current := range as.selectedAttributes {
		//fmt.Println(current, i, inst.RealValues())
		newVasl =  append(newVasl,inst.Value(current))
		//newVasl[i] = inst.Value(current)
		//fmt.Println(newVasl[i], "newVasl[i]")
	}
	//fmt.Println("----------------------------------------------")
	newInst := data.NewInstance()
	newInst.SetNumAttributes(len(as.output.Attributes()))
	//fmt.Println(newInst.NumAttributes())
	values_ := make([]float64, len(newVasl))
	indices_ := make([]int, len(newVasl))
	vals := 0
	for i := 0; i < len(newVasl); i++ {
		if newVasl[i] != 0 {
			values_[vals] = newVasl[i]
			indices_[vals] = i
			vals++
		}
	}
	values := make([]float64, vals)
	indices := make([]int, vals)
	copy(values, values_)
	copy(indices, indices_)
//	fmt.Println(values, "values")
//	fmt.Println(indices, "indices")
	for k, i := range indices {
		if as.output.Attribute(i).IsNominal() {
			newInst.AddValues(as.output.Attribute(i).Values()[int(values[k])])
		} else {
			newInst.AddValues(as.output.Attribute(i).Name())
		}
	}
	newInst.SetIndices(indices)
	newInst.SetRealValues(values)
	newInst.SetWeight(inst.Weight())
	return newInst
}

func (as *AttributeSelection) ConvertInstance(inst data.Instance) data.Instance {
	//fmt.Println(len(as.output.Attributes()))
	newVasl := make([]float64, len(as.output.Attributes()))
	for i, current := range as.selectedAttributes {
		newVasl[i] =  inst.Value(current)
	}

	newInst := data.NewInstance()
	values_ := make([]float64, len(newVasl))
	indices_ := make([]int, len(newVasl))
	vals := 0
	for i := 0; i < len(newVasl); i++ {
		if newVasl[i] != 0 {
			values_[vals] = newVasl[i]
			indices_[vals] = i
			vals++
		}
	}
	values := make([]float64, vals)
	indices := make([]int, vals)
	copy(values, values_)
	copy(indices, indices_)
	//fmt.Println(len(indices))
//	for k, i := range indices {
//		if as.output.Attribute(i).IsNominal() {
//			newInst.AddValues(as.output.Attribute(i).Values()[int(values[k])])
//		} else {
//			newInst.AddValues(as.output.Attribute(i).Name())
//		}
//	}
fmt.Print()
	newInst.SetIndices(indices)
	newInst.SetRealValues(values)
	newInst.SetWeight(inst.Weight())
	newInst.SetNumAttributes(len(newVasl))
	return newInst
}

func (as *AttributeSelection) SelectAttributes(data_ data.Instances) []int {
	//***********attributeSet := make([]int, 0)
	as.trainInstances = data_
	as.doRank = as.search.GenerateRanking()
	// check that a class index has been set
	if as.trainInstances.ClassIndex() < 0 {
		as.trainInstances.SetClassIndex(len(as.trainInstances.Attributes()) - 1)
	}
	// Initialize the attribute evaluator
	as.evaluator.BuildEvaluator(as.trainInstances)
	//fieldWith := int(math.Log(float64(len(as.trainInstances.Attributes()) + 1)))
	// Do the search
	//***********attributeSet =
	as.search.Search(as.evaluator, as.trainInstances)
	// InfoGain do not implements postprocessing in weka

	//I won't use this check because in this implementation it will always be true
	//due that search method always is going to be Ranker
	if as.doRank {
	}
	as.attributeRanking = as.search.rankedAttributes()
	// retrieve the number of attributes to retain
	as.numToSelect = as.search.GetCalculatedNumToSelect()
	//fmt.Println(as.numToSelect, "as.numToSelect")
	// determine fieldwidth for merit
	f_p, w_p := 0, 0
	for i := 0; i < as.numToSelect; i++ {
		precision := math.Abs(as.attributeRanking[i][1]) - math.Abs(as.attributeRanking[i][1])
		intPart := int(math.Abs(as.attributeRanking[i][1]))
		if precision > 0 {
			precision = math.Abs((math.Log(math.Abs(precision)) / math.Log(10))) + 3
		}
		if precision > float64(f_p) {
			f_p = int(precision)
		}
		if intPart == 0 {
			if w_p < 2 {
				w_p = 2
			}
		} else if (math.Abs((math.Log(math.Abs(as.attributeRanking[i][1])) / math.Log(10))) + 1) > float64(w_p) {
			if as.attributeRanking[i][1] > 0 {
				w_p = int(math.Abs((math.Log(math.Abs(as.attributeRanking[i][1])) / math.Log(10))) + 1)
			}
		}
	}
	// set up the selected attributes array - usable by a filter or
	// whatever
	if as.trainInstances.ClassIndex() >= 0 {
		as.selectedAttributeSet = make([]int, as.numToSelect+1)
		as.selectedAttributeSet[as.numToSelect] = as.trainInstances.ClassIndex()
	} else {
		as.selectedAttributeSet = make([]int, as.numToSelect)
	}
	for i := 0; i < as.numToSelect; i++ {
		as.selectedAttributeSet[i] = int(as.attributeRanking[i][0])
	}
	//fmt.Println(as.selectedAttributeSet, "as.selectedAttributeSet")
	if as.doXval {
		as.CrossValidateAttribute()
	}
	if as.selectedAttributeSet != nil && !as.doXval {
		as.attributeFilter = NewRemove()
		as.attributeFilter.SetSelectedColumns(as.selectedAttributeSet)
		as.attributeFilter.SetInvertSelection(true)
		as.attributeFilter.SetInputFormat(as.trainInstances)
	}
	as.trainInstances = data.NewInstancesWithInst(as.trainInstances, 0)
	return as.selectedAttributeSet
}

func (as *AttributeSelection) CrossValidateAttribute() {
	cvData := as.trainInstances
	var train data.Instances
	cvData.Randomize(as.seed)
	for i := 0; i < as.numFolds; i++ {
		train = cvData.TrainCV(as.numFolds, i, as.seed)
		as.selectAttributesCVSplit(train)
	}

}

//Select attributes for a split of the data
func (as *AttributeSelection) selectAttributesCVSplit(split data.Instances) {
	attributeRanking := make([][]float64, 0)
	//this is only helpfull if this method is called from outside not from inner method of the object
	//	if as.trainInstances.(nil)  {
	//		as.trainInstances =  split
	//	}

	// create space to hold statistics
	if as.rankResults == nil && as.subSetResults == nil {
		as.subSetResults = make([]float64, len(split.Attributes()))
		as.rankResults = make([][]float64, 4)
		for i := range as.rankResults {
			as.rankResults[i] = make([]float64, len(split.Attributes()))
		}

	}
	as.evaluator.BuildEvaluator(split)
	// Do the search
	attributeSet := as.search.Search(as.evaluator, split)
	if as.doRank {
		attributeRanking = as.search.rankedAttributes()
		for j := range attributeRanking {
			// merit
			as.rankResults[0][int(attributeRanking[j][0])] += attributeRanking[j][1]
			// squared merit
			as.rankResults[2][int(attributeRanking[j][0])] += (attributeRanking[j][1] * attributeRanking[j][1])
			// rank
			as.rankResults[1][int(attributeRanking[j][0])] += float64(j + 1)
			// squared rank
			as.rankResults[3][int(attributeRanking[j][0])] += float64((j + 1) * (j + 1))
		}
	} else {
		for j := range attributeSet {
			as.subSetResults[attributeSet[j]]++
		}
	}
	as.trials++
}

func (as *AttributeSelection) SetEvaluator(eval InfoGain) {
	as.evaluator = eval
}

func (as *AttributeSelection) SetSearchMethod(method Ranker) {
	as.search = method
}

func (as *AttributeSelection) Output() data.Instances {
	return as.output
}
