package functions

import (
	"github.com/project-mac/src/data"
	"math"
	//	"utils"
)

type AttributeSelection struct {
	//the selected attributes
	SelectedAttributeSet, SelectedAttributes []int
	//the attribute indexes and associated merits if a ranking is produced
	AttributeRanking [][]float64
	//number of attributes requested from ranked results,
	//the number of folds to use for cross validation
	NumToSelect, NumFolds int
	//
	AttributeFilter Remove
	//rank features,
	//do cross validation
	DoRank, DoXval bool
	//the instances to select attributes from
	trainInstances data.Instances
	Evaluator      InfoGain
	Search         Ranker
	Seed           int //used ofr randomize instances in crossvalidation
	//these are only for validation statistics, if not useful can be remove later
	rankResults   [][]float64
	subSetResults []float64
	trials        int

	Input_, Output_ data.Instances
	HasClass        bool
}

func NewAttributeSelection() AttributeSelection {
	var as AttributeSelection
	return as
}

//Start the selection attributes process
func (as *AttributeSelection) StartSelection(instances data.Instances) {
	as.Input_ = data.NewInstancesWithInst(instances, len(instances.Attributes()))
	as.Input_ = instances
	as.Output_ = data.NewInstances()
	as.HasClass = as.Input_.ClassIndex() >= 0
	as.SelectedAttributes = as.SelectAttributes(as.Input_)
	if len(as.SelectedAttributes) == 0 {
		panic("No selected attributes")
	}
	//Set Output_
	as.Output_ = data.NewInstances()
	attributes := make([]data.Attribute, 0)
	for i := range as.SelectedAttributes {
		attributes = append(attributes, *as.Input_.Attribute(as.SelectedAttributes[i]))
	}
	as.Output_.SetDatasetName(as.Input_.DatasetName())
	as.Output_.SetAttributes(attributes)
	if as.HasClass {
		as.Output_.SetClassIndex(len(as.SelectedAttributes) - 1)
	}
	// Convert pending Input_ instances
	tmpInst := make([]data.Instance, 0)
	for _, in := range as.Input_.Instances() {
		tmpInst = append(tmpInst, as.convertInstance(in))

	}
	as.Output_.SetInstances(tmpInst)
}

//Convert a single instance over
func (as *AttributeSelection) convertInstance(inst data.Instance) data.Instance {
	newVasl := make([]float64, 0, len(as.Output_.Attributes()))
	for _, current := range as.SelectedAttributes {
		newVasl = append(newVasl, inst.Value(current))
	}
	newInst := data.NewInstance()
	newInst.SetNumAttributes(len(as.Output_.Attributes()))
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
	for k, i := range indices {
		if as.Output_.Attribute(i).IsNominal() {
			newInst.AddValues(as.Output_.Attribute(i).Values()[int(values[k])])
		} else {
			newInst.AddValues(as.Output_.Attribute(i).Name())
		}
	}
	newInst.SetIndices(indices)
	newInst.SetRealValues(values)
	newInst.SetWeight(inst.Weight())
	return newInst
}

func (as *AttributeSelection) ConvertInstance(inst data.Instance) data.Instance {
	newVasl := make([]float64, len(as.Output_.Attributes()))
	for i, current := range as.SelectedAttributes {
		newVasl[i] = inst.Value(current)
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
	newInst.SetIndices(indices)
	newInst.SetRealValues(values)
	newInst.SetWeight(inst.Weight())
	newInst.SetNumAttributes(len(newVasl))
	return newInst
}

func (as *AttributeSelection) SelectAttributes(data_ data.Instances) []int {
	//***********attributeSet := make([]int, 0)
	as.trainInstances = data_
	as.DoRank = as.Search.GenerateRanking()
	// check that a class index has been set
	if as.trainInstances.ClassIndex() < 0 {
		as.trainInstances.SetClassIndex(len(as.trainInstances.Attributes()) - 1)
	}
	// Initialize the attribute Evaluator
	as.Evaluator.BuildEvaluator(as.trainInstances)
	//fieldWith := int(math.Log(float64(len(as.trainInstances.Attributes()) + 1)))
	// Do the Search
	//***********attributeSet =
	as.Search.Search(as.Evaluator, as.trainInstances)
	// InfoGain do not implements postprocessing in weka

	//I won't use this check because in this implementation it will always be true
	//due that Search method always is going to be Ranker
	if as.DoRank {
	}
	as.AttributeRanking = as.Search.RankedAttributes()
	// retrieve the number of attributes to retain
	as.NumToSelect = as.Search.GetCalculatedNumToSelect()
	// determine fieldwidth for merit
	f_p, w_p := 0, 0
	for i := 0; i < as.NumToSelect; i++ {
		precision := math.Abs(as.AttributeRanking[i][1]) - math.Abs(as.AttributeRanking[i][1])
		intPart := int(math.Abs(as.AttributeRanking[i][1]))
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
		} else if (math.Abs((math.Log(math.Abs(as.AttributeRanking[i][1])) / math.Log(10))) + 1) > float64(w_p) {
			if as.AttributeRanking[i][1] > 0 {
				w_p = int(math.Abs((math.Log(math.Abs(as.AttributeRanking[i][1])) / math.Log(10))) + 1)
			}
		}
	}
	// set up the selected attributes array - usable by a filter or
	// whatever
	if as.trainInstances.ClassIndex() >= 0 {
		as.SelectedAttributeSet = make([]int, as.NumToSelect+1)
		as.SelectedAttributeSet[as.NumToSelect] = as.trainInstances.ClassIndex()
	} else {
		as.SelectedAttributeSet = make([]int, as.NumToSelect)
	}
	for i := 0; i < as.NumToSelect; i++ {
		as.SelectedAttributeSet[i] = int(as.AttributeRanking[i][0])
	}

	if as.DoXval {
		as.CrossValidateAttribute()
	}
	if as.SelectedAttributeSet != nil && !as.DoXval {
		as.AttributeFilter = NewRemove()
		as.AttributeFilter.SetSelectedColumns(as.SelectedAttributeSet)
		as.AttributeFilter.SetInvertSelection(true)
		as.AttributeFilter.SetInputFormat(as.trainInstances)
	}
	as.trainInstances = data.NewInstancesWithInst(as.trainInstances, 0)
	return as.SelectedAttributeSet
}

func (as *AttributeSelection) CrossValidateAttribute() {
	cvData := as.trainInstances
	var train data.Instances
	cvData.Randomize(as.Seed)
	for i := 0; i < as.NumFolds; i++ {
		train = cvData.TrainCV(as.NumFolds, i, as.Seed)
		as.selectAttributesCVSplit(train)
	}

}

//Select attributes for a split of the data
func (as *AttributeSelection) selectAttributesCVSplit(split data.Instances) {
	AttributeRanking := make([][]float64, 0)
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
	as.Evaluator.BuildEvaluator(split)
	// Do the Search
	attributeSet := as.Search.Search(as.Evaluator, split)
	if as.DoRank {
		AttributeRanking = as.Search.RankedAttributes()
		for j := range AttributeRanking {
			// merit
			as.rankResults[0][int(AttributeRanking[j][0])] += AttributeRanking[j][1]
			// squared merit
			as.rankResults[2][int(AttributeRanking[j][0])] += (AttributeRanking[j][1] * AttributeRanking[j][1])
			// rank
			as.rankResults[1][int(AttributeRanking[j][0])] += float64(j + 1)
			// squared rank
			as.rankResults[3][int(AttributeRanking[j][0])] += float64((j + 1) * (j + 1))
		}
	} else {
		for j := range attributeSet {
			as.subSetResults[attributeSet[j]]++
		}
	}
	as.trials++
}

func (as *AttributeSelection) SetEvaluator(eval InfoGain) {
	as.Evaluator = eval
}

func (as *AttributeSelection) SetSearchMethod(method Ranker) {
	as.Search = method
}

func (as *AttributeSelection) Output() data.Instances {
	return as.Output_
}
