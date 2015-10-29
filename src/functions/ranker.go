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
	IsRangeInUse  bool      //if we using a range of of not include attributes
	NotInclude    []int     //Indexes of not include atributes, attributes not include are attributes that it won't be processed
	NumAttributes int       //The number of attribtes
	HasClass      bool      //Data has class attribute
	ClassIndex    int       //Class index of the data
	AttrList      []int     //Holds the ordered list of attributes
	AttrMerit     []float64 //Holds the list of attribute merit scores
	Threshold     float64   // A Threshold by which to discard attributes
	NumToSelect   int       //The number of attributes to select. -1 indicates that all attributes
	//are to be retained. Has precedence over Threshold
	CalculateNumToSelect int
}

//Create a new Ranker
func NewRanker() Ranker {
	var r Ranker
	r.IsRangeInUse = false
	r.NumToSelect = -1
	r.Threshold = -math.MaxFloat64
	return r
}

//Kind of a dummy search algorithm. Calls a Attribute evaluator to
//evaluate each attribute not included in the NotInclude array and then sorts
//them to produce a ranked list of attributes.
func (r *Ranker) Search(evaluator InfoGain, instances data.Instances) []int {
	var i, j int
	r.NumAttributes = len(instances.Attributes())
	r.ClassIndex = instances.ClassIndex()
	if r.ClassIndex >= 0 {
		r.HasClass = true
	} else {
		r.HasClass = false
	}
	sl := 0
	if r.IsRangeInUse || len(r.NotInclude) > 0 {
		sl = len(r.NotInclude)
	}
	if (r.IsRangeInUse || len(r.NotInclude) > 0) && r.HasClass {
		// see if the supplied list contains the class index
		isIn := false
		for _, sel := range r.NotInclude {
			if sel == r.ClassIndex {
				isIn = true
				break
			}
		}
		if !isIn {
			sl++
		}
	} else {
		if r.HasClass {
			sl++
		}
	}
	r.AttrList = make([]int, r.NumAttributes-sl)
	r.AttrMerit = make([]float64, r.NumAttributes-sl)
	// add in those attributes that are include to select
	for i, j = 0, 0; i < r.NumAttributes; i++ {
		if !r.InNotInclude(i) {
			r.AttrList[j] = i
			j++
		}
	}
	for i := range r.AttrList {
		r.AttrMerit[i] = evaluator.EvaluateAttribute(r.AttrList[i])
	}
	tempRanked := r.RankedAttributes()
	rankedAttributes := make([]int, len(r.AttrList))
	for i := range rankedAttributes {
		rankedAttributes[i] = int(tempRanked[i][0])
	}
	return rankedAttributes
}

//Sorts the evaluated attribute list
func (r *Ranker) RankedAttributes() [][]float64 {
	var i, j int
	if len(r.AttrList) == 0 || len(r.AttrMerit) == 0 {
		panic("Fisrt execute the search to obtain the ranked attribute list")
	}
	rank := make([]int, len(r.AttrMerit))
	h := 0
	for i := range r.AttrMerit {
		rank[i] = h
		h++
	}
	ranked := utils.SortFloat(r.AttrMerit)
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
		bestToWorst[i][0] = float64(r.AttrList[temp])
		bestToWorst[i][1] = r.AttrMerit[temp]
	}
	if r.NumToSelect > len(bestToWorst) {
		panic("More attributes requested than exist in the data")
	}
	if r.NumToSelect <= 0 {
		if r.Threshold == -math.MaxFloat64 {
			r.CalculateNumToSelect = len(bestToWorst)
		} else {
			r.DetermineNumToSelectFromThreshold(bestToWorst)
		}
	}

	return bestToWorst
}

func (r *Ranker) DetermineNumToSelectFromThreshold(ranking [][]float64) {
	count := 0
	for i := range ranking {
		if ranking[i][1] > r.Threshold {
			count++
		}
	}
	r.CalculateNumToSelect = count
}

func (r *Ranker) InNotInclude(f int) bool {
	// omit the class from the evaluation
	if r.HasClass && r.ClassIndex == f {
		return true
	}

	if !r.IsRangeInUse || len(r.NotInclude) == 0 {
		return false
	}

	for _, sel := range r.NotInclude {
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
		r.NotInclude = selected
		r.IsRangeInUse = true
	}
}

func (r *Ranker) GenerateRanking() bool {
	return true
}

func (r *Ranker) GetCalculatedNumToSelect() int {
	if r.NumToSelect > 0 {
		r.CalculateNumToSelect = r.NumToSelect
	}
	return r.CalculateNumToSelect
}

func (r *Ranker) NotInclude_() []int {
	return r.NotInclude
}

func (r *Ranker) SetThreshold(Threshold float64) {
	r.Threshold = Threshold
}
func (r *Ranker) SetNumToSelect(NumToSelect int) {
	r.NumToSelect = NumToSelect
}
