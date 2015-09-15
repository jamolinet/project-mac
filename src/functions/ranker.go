package functions

import (
	"github.com/project-mac/src/data"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"github.com/project-mac/src/utils"
)

type Ranker struct {
	isRangeInUse  bool      //if we using a range of of not include attributes
	notInclude    []int     //Indexes of not include atributes, attributes not include are attributes that it won't be processed
	numAttributes int       //The number of attribtes
	hasClass      bool      //Data has class attribute
	classIndex    int       //Class index of the data
	attrList      []int     //Holds the ordered list of attributes
	attrMerit     []float64 //Holds the list of attribute merit scores
	threshold     float64   // A threshold by which to discard attributes
	numToSelect   int       //The number of attributes to select. -1 indicates that all attributes
	//are to be retained. Has precedence over threshold
	calculateNumToSelect int
}

//Create a new Ranker
func NewRanker() Ranker {
	var r Ranker
	r.isRangeInUse = false
	r.numToSelect = -1
	r.threshold = -math.MaxFloat64
	return r
}

//Kind of a dummy search algorithm. Calls a Attribute evaluator to
//evaluate each attribute not included in the notInclude array and then sorts
//them to produce a ranked list of attributes.
func (r *Ranker) Search(evaluator InfoGain, instances data.Instances) []int {
	var i, j int
	r.numAttributes = len(instances.Attributes())
	r.classIndex = instances.ClassIndex()
	if r.classIndex >= 0 {
		r.hasClass = true
	} else {
		r.hasClass = false
	}
	sl := 0
	if r.isRangeInUse || len(r.notInclude) > 0 {
		sl = len(r.notInclude)
	}
	if (r.isRangeInUse || len(r.notInclude) > 0) && r.hasClass {
		// see if the supplied list contains the class index
		isIn := false
		for _, sel := range r.notInclude {
			if sel == r.classIndex {
				isIn = true
				break
			}
		}
		if !isIn {
			sl++
		}
	} else {
		if r.hasClass {
			sl++
		}
	}
	//fmt.Println(len(r.notInclude), "noInclude")
	//fmt.Println(sl, "sl")
	r.attrList = make([]int, r.numAttributes-sl)
	r.attrMerit = make([]float64, r.numAttributes-sl)
	// add in those attributes that are include to select
	for i, j = 0, 0; i < r.numAttributes; i++ {
		if !r.inNotInclude(i) {
			r.attrList[j] = i
			j++
		}
	}
	//fmt.Println(r.attrList, "r.attrList")
	for i := range r.attrList {
		r.attrMerit[i] = evaluator.evaluateAttribute(r.attrList[i])
	}
	//fmt.Println(r.attrMerit, "merrit")
	tempRanked := r.rankedAttributes()
	rankedAttributes := make([]int, len(r.attrList))
	for i := range rankedAttributes {
		rankedAttributes[i] = int(tempRanked[i][0])
	}
	//fmt.Println(rankedAttributes, "rankedAttributes")
	return rankedAttributes
}

//Sorts the evaluated attribute list
func (r *Ranker) rankedAttributes() [][]float64 {
	var i, j int
	if len(r.attrList) == 0 || len(r.attrMerit) == 0 {
		panic("Fisrt execute the search to obtain the ranked attribute list")
	}
	rank := make([]int, len(r.attrMerit))
	h := 0
	for i := range r.attrMerit {
		rank[i] = h
		h++
	}
	ranked := utils.SortFloat(r.attrMerit)
	//fmt.Println(ranked, r.attrMerit, "ranked")
	// reverse the order of the ranked indexes
	bestToWorst := make([][]float64, len(ranked))
	for i := range bestToWorst {
		bestToWorst[i] = make([]float64, 2)
	}
	for i, j = len(ranked)-1, 0; i >= 0; i-- {
		bestToWorst[j][0] = float64(ranked[i])
		j++
	}
	// convert the indexes to attribute indexes
	for i := range bestToWorst {
		temp := int(bestToWorst[i][0])
		bestToWorst[i][0] = float64(r.attrList[temp])
		bestToWorst[i][1] = r.attrMerit[temp]
	}
	if r.numToSelect > len(bestToWorst) {
		panic("More attributes requested than exist in the data")
	}
	if r.numToSelect <= 0 {
		if r.threshold == -math.MaxFloat64 {
			r.calculateNumToSelect = len(bestToWorst)
		} else {
			r.determineNumToSelectFromThreshold(bestToWorst)
		}
	}
	//fmt.Println(bestToWorst, "bestToWorst")
	return bestToWorst
}

func (r *Ranker) determineNumToSelectFromThreshold(ranking [][]float64) {
	count := 0
	for i := range ranking {
		if ranking[i][1] > r.threshold {
			count++
		}
	}
	r.calculateNumToSelect = count
}

func (r *Ranker) inNotInclude(f int) bool {
	// omit the class from the evaluation
	if r.hasClass && r.classIndex == f {
		return true
	}

	if !r.isRangeInUse || len(r.notInclude) == 0 {
		return false
	}

	for _, sel := range r.notInclude {
		if sel == f {
			return true
		}
	}

	return false
}

func (r *Ranker) SetRange(rang string) {
	if strings.EqualFold(rang, "") {
		panic("The range cannot be empty")
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
		r.notInclude = selected
		r.isRangeInUse = true
	}
}

func (r *Ranker) GenerateRanking() bool {
	return true
}

func (r *Ranker) GetCalculatedNumToSelect() int {
	if r.numToSelect > 0 {
		r.calculateNumToSelect = r.numToSelect
	}
	return r.calculateNumToSelect
}

func (r *Ranker) NotInclude() []int {
	return r.notInclude
}

func (r *Ranker) SetThreshold(threshold float64) {
	r.threshold = threshold
}
func (r *Ranker) SetNumToSelect(numToSelect int) {
	r.numToSelect = numToSelect
}
