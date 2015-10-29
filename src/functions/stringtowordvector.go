package functions

import (
	"fmt"
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/omap"
	"math"
	"sort"
	"strings"
	//	"strconv"
)

const (
	TF     = 0
	IDF    = 1
	TF_IDF = 2
)

type Count struct {
	Count, DocCount int
}

type StringToWordVector struct {
	FirstTime bool
	//Contains a mapping of valid words to attribute indexes.
	Dictionary *omap.Map
	//True if output instances should contain word frequency rather than boolean 0 or 1.
	OutputsCounts bool
	//Contains the number of documents (instances) a particular word appears in.
	//The counts are stored with the same indexing as given by m_Dictionary.
	DocsCounts []int
	// Contains the number of documents (instances) in the input format from
	//which the Dictionary is created. It is used in IDF transform.
	NumInstances int
	// Contains the average length of documents (among the first batch of
	//instances aka training data). This is used in length normalization of
	//documents which will be normalized to average document length.
	AvgDocLength float64
	//The default number of words (per class if there is a class attribute
	//assigned) to attempt to keep.
	WordsToKeep int
	//Which transformation
	TF_transformation, IDF_transformation bool
	//The percentage at which to periodically prune the Dictionary.
	PerdiodicPruningRate int
	//Use normalization or not
	Normalize bool
	//whether to operate on per-class basis
	PerClass bool
	//the minimum (per-class) word frequency.
	MinTermFreq rune
	//The input format for instances.
	InputFormat data.Instances
	//The output format for instances.
	OutputFormat data.Instances
}

//New StringToWordVector function with default values
func NewStringToWordVectorInst(inputData data.Instances) StringToWordVector {
	var stwv StringToWordVector
	stwv.Dictionary = omap.NewStringKeyed()
	stwv.OutputsCounts = false
	stwv.DocsCounts = make([]int, 0)
	stwv.AvgDocLength = -1
	stwv.WordsToKeep = 1000
	stwv.NumInstances = -1
	stwv.PerdiodicPruningRate = -1
	stwv.MinTermFreq = 1
	stwv.PerClass = true
	stwv.Normalize = true
	stwv.InputFormat = inputData
	//stwv.OutputFormat = data.NewInstancesWithClassIndex(inputData.ClassIndex())
	stwv.FirstTime = true
	stwv.TF_transformation, stwv.IDF_transformation = true, true
	return stwv
}

func NewStringToWordVector() StringToWordVector {
	var stwv StringToWordVector
	stwv.Dictionary = omap.NewStringKeyed()
	stwv.OutputsCounts = false
	stwv.DocsCounts = make([]int, 0)
	stwv.AvgDocLength = -1
	stwv.WordsToKeep = 1000
	stwv.NumInstances = -1
	stwv.PerdiodicPruningRate = -1
	stwv.MinTermFreq = 1
	stwv.PerClass = true
	stwv.Normalize = true
	stwv.FirstTime = true
	stwv.TF_transformation, stwv.IDF_transformation = true, true
	return stwv
}

func (m *StringToWordVector) SetInputFormatAndOutputFormat(inputData data.Instances) {
	m.InputFormat = inputData
	//m.OutputFormat = data.NewInstancesWithClassIndex(inputData.ClassIndex())
}

//Start the execution of the function
func (stwv *StringToWordVector) Exec() data.Instances {
	/* TODO: first check that the input format is initialized*/
	//	if stwv.InputFormat != nil {
	//		panic("No input instace defined")
	//	}
	inst := &stwv.InputFormat
	if stwv.FirstTime {
		//check if the class attribute is not nominal to turn off the per-class basis
		classAttr := inst.Attributes()[inst.ClassIndex()]
		if classAttr.Type() != data.NOMINAL {
			stwv.PerClass = false
		}
		//Determine the Dictionary for the input format
		stwv.DetermineDictionary(inst)
		//Convert all instances without Normalize
		fv := make([]data.Instance, 0)
		firstCopy := 0
		for i := 0; i < stwv.NumInstances; i++ {
			firstcopy, v := stwv.ConvertInstancewoDocNorm(stwv.InputFormat.Instances()[i])
			fv = append(fv, v)
			firstCopy = firstcopy
		}
		// Need to compute average document length if necessary
		stwv.AvgDocLength = 0
		for _, inst := range fv {
			docLength := float64(0)
			for j := 0; j < len(inst.RealValues()); j++ {
				if inst.Indices()[j] >= firstCopy {
					docLength += inst.RealValues()[j] * inst.RealValues()[j]
				}
			}
			stwv.AvgDocLength += math.Sqrt(docLength)
		}
		stwv.AvgDocLength /= float64(stwv.NumInstances)
		// Perform normalization if necessary.
		if stwv.Normalize {
			for _, inst := range fv {
				stwv.NormalizeInstance(&inst, firstCopy)
			}
		}
		stwv.OutputFormat.SetInstances(fv)
	} else {
		fv := make([]data.Instance, 0)
		firstCopy := 0
		for i := 0; i < stwv.NumInstances; i++ {
			firstcopy, v := stwv.ConvertInstancewoDocNorm(*stwv.InputFormat.Instance(i))
			fv = append(fv, v)
			firstCopy = firstcopy
		}
		if stwv.Normalize {
			for _, inst := range fv {
				stwv.NormalizeInstance(&inst, firstCopy)
			}
		}
	}
	fmt.Println("Done!")
	stwv.FirstTime = false
	return stwv.OutputFormat
}

func (stwv *StringToWordVector) Input(instance data.Instance) data.Instance {
	if !stwv.FirstTime {
		firstcopy, inst := stwv.ConvertInstancewoDocNorm(instance)
		if stwv.Normalize {
			stwv.NormalizeInstance(&inst, firstcopy)
		}
		return inst
	}
	return instance
}

func (stwv *StringToWordVector) DetermineDictionary(inst *data.Instances) {
	/* TODO: see if use a stopwords list*/
	fmt.Println("Determing Dictionary!")
	classInd := inst.ClassIndex()
	values := 1
	if stwv.PerClass && (classInd != -1) {
		values = len(inst.Attributes()[classInd].Values())
	}
	dicA := make([]*omap.Map, values)
	for i := 0; i < values; i++ {
		dicA[i] = omap.NewStringKeyed()
	}
	// Tokenize all training text into an orderedMap of "words".
	pruneRate := int64((stwv.PerdiodicPruningRate / 100) * len(inst.Instances()))
	for i, instance := range inst.Instances() {
		vInd := int(0)
		if stwv.PerClass && (classInd != -1) {
			vInd = int(instance.RealValues()[classInd])
		}
		//Iterate through all relevant string attributes of the current instance
		hashtable := make(map[string]int, 0)
		for j := 0; j < instance.NumAttributes(); j++ {
			if !instance.IsMissingValue(j) && inst.Attributes()[j].IsString() {
				// Iterate through tokens, perform stemming, and remove stopwords
				// (if required)
				words := strings.Fields(instance.Values()[j])
				for _, word := range words {
					_, present := hashtable[word]
					if !present {
						hashtable[word] = 0
					}
					if count, present := dicA[vInd].Find(word); !present {
						dicA[vInd].Insert(word, Count{1, 0})
					} else {
						count, _ := count.(Count)
						count.Count++
						dicA[vInd].Insert(word, count)
					}
				}
			}
		}
		//updating the docCount for the words that have occurred in this
		//instance(document).
		enumeration := make([]string, 0, len(hashtable))
		for word, _ := range hashtable { //only the words
			enumeration = append(enumeration, word)
		}
		for _, word := range enumeration {
			if count, present := dicA[vInd].Find(word); present {
				count := count.(Count)
				count.DocCount++
				//delete(dicA[vInd], word)
				dicA[vInd].Insert(word, count)
			} else {
				panic("Check the code, there must be a word in the Dictionary")
			}
		}

		if pruneRate > 0 {
			if int64(i)%pruneRate == 0 && i > 0 {
				for z := 0; z < values; z++ {
					d := make([]string, 1000)
					dicA[z].Do(func(key, value interface{}) {
						word, _ := key.(string)
						count, _ := value.(Count)
						if count.Count <= 1 {
							d = append(d, word)
						}
					})
					//					for word, _ := range dicA[z] {
					//						count := dicA[z][word]
					//						if count.Count <= 1 {
					//							d = append(d, word)
					//						}
					//					}
					for _, word := range d {
						dicA[z].Delete(word)
						//delete(dicA[z], word)
					}
				}
			}
		}
	}
	// Figure out the minimum required word frequency
	totalSize := int(0)
	prune := make([]int, values)
	for z := 0; z < values; z++ {
		totalSize += dicA[z].Len()
		array := make([]int, dicA[z].Len())
		pos := int(0)
		dicA[z].Do(func(key, value interface{}) {
			//_, _ := key.(string)
			count, _ := value.(Count)
			array[pos] = count.Count
			pos++
		})
		sort.Ints(array)
		if len(array) < stwv.WordsToKeep {
			// if there aren't enough words, set the threshold to
			// minFreq
			prune[z] = int(stwv.MinTermFreq)
		} else {
			// otherwise set it to be at least minFreq
			idx := len(array) - stwv.WordsToKeep
			prune[z] = int(math.Max(float64(stwv.MinTermFreq), float64(array[idx])))
		}
	}
	// Convert the Dictionary into an attribute index
	// and create one attribute per word
	attributes := make([]data.Attribute, 0, totalSize+len(inst.Attributes()))
	// Add the non-converted attributes
	classIndex := int(-1)
	for i, attr := range stwv.InputFormat.Attributes() {
		if !attr.IsString() {
			if inst.ClassIndex() == i {
				classIndex = len(attributes)
			}
			attributes = append(attributes, attr)
		}
	}
	// Add the word vector attributes (eliminating duplicates
	// that occur in multiple classes)
	newDic := omap.NewStringKeyed()
	index := len(attributes)
	for z := 0; z < values; z++ {
		dicA[z].Do(func(key, value interface{}) {
			word, _ := key.(string)
			count, _ := value.(Count)
			if count.Count >= prune[z] {
				if _, present := newDic.Find(word); !present {
					newDic.Insert(word, int(index))
					index++
					att := data.NewAttribute()
					att.SetName(word)
					att.SetType(data.NUMERIC)
					attributes = append(attributes, att)
				}
			}

		})
	}
	// Compute document frequencies
	stwv.DocsCounts = make([]int, len(attributes))
	//idx := 0
	newDic.Do(func(key, value interface{}) {
		word, _ := key.(string)
		idx, _ := value.(int)
		docsCount := 0
		for j := 0; j < values; j++ {
			if count, present := dicA[j].Find(word); present {
				count := count.(Count)
				docsCount += count.DocCount
			}
		}
		stwv.DocsCounts[idx] = docsCount
	})
	
	stwv.Dictionary = newDic
	stwv.NumInstances = len(inst.Instances())
	stwv.OutputFormat = data.NewInstances()
	stwv.OutputFormat.SetAttributes(attributes)
	stwv.OutputFormat.SetClassIndex(classIndex)
	}

func (stwv *StringToWordVector) ConvertInstancewoDocNorm(inst data.Instance) (int, data.Instance) {

	// Convert the instance into a sorted set of indexes
	contained := omap.NewIntKeyed()
	mapKeys := make([]float64, 0)

	// Copy all non-converted attributes from input to output
	firstCopy := 0
	for i, _ := range stwv.InputFormat.Attributes() {
		if !stwv.InputFormat.Attributes()[i].IsString() {
			// Add simple nominal and numeric attributes directly
			if inst.RealValues()[i] != 0 {
				contained.Insert(firstCopy, inst.RealValues()[i])
				mapKeys = append(mapKeys, float64(firstCopy))
				firstCopy++
			} else {
				firstCopy++
			}
		} else if inst.IsMissingValue(i) {
			mapKeys = append(mapKeys, float64(firstCopy))
			firstCopy++
		} else if stwv.InputFormat.Attributes()[i].IsString() {
			//if i have to implement the range selector then code this part
		}
	}
	//Copy the converted attributes
	for j := 0; j < inst.NumAttributes(); j++ {
		if stwv.InputFormat.Attributes()[j].IsString() && inst.IsMissingValue(j) == false {
			words := strings.Fields(inst.Values()[j])
			for _, word := range words {
				if index, present := stwv.Dictionary.Find(word); present {
					if stwv.OutputsCounts {
						if count, isthere := contained.Find(index); isthere {
							if count, ok := count.(float64); ok { //type assertion
								contained.Insert(int(index.(int)), count+1)
								mapKeys = append(mapKeys, float64(index.(int)))
							}
						} else {
							contained.Insert(int(index.(int)), float64(1))
							mapKeys = append(mapKeys, float64(index.(int)))
						}
					} else {
						contained.Insert(int(index.(int)), float64(1))
						mapKeys = append(mapKeys, float64(index.(int)))
					}
				}
			}
		}
	}
	//To calculate frequencies
	indexes := make([]int, contained.Len())
	_values := make([]float64, contained.Len())
	n := 0
	contained.Do(func(key, value interface{}) {
		index, _ := key.(int)
		_value, _ := value.(float64)
		indexes[n] = index
		_values[n] = _value
		n++
	})
	//------------
	//TF_freq transform
	if stwv.TF_transformation {
		for i := 0; i < len(indexes); i++ {
			index := indexes[i]
			if index >= firstCopy {
				val := _values[i]
				val = math.Log(val + 1)
				contained.Insert(index, val)
			}
		}
	}
	indexes = make([]int, contained.Len())
	_values = make([]float64, contained.Len())
	n = 0
	contained.Do(func(key, value interface{}) {
		index, _ := key.(int)
		_value, _ := value.(float64)
		indexes[n] = index
		_values[n] = _value
		n++
	})
	//IDF_freq transform
	if stwv.IDF_transformation {
		for i := 0; i < len(indexes); i++ {
			index := indexes[i]
			if index >= firstCopy {
				val := _values[i]
				val = val * math.Log(float64(stwv.NumInstances)/float64(stwv.DocsCounts[index]))
				contained.Insert(index, val)
			}
		}
	}

	// Convert the set to structures needed to create a sparse instance.
	values := make([]float64, contained.Len())
	indices := make([]int, contained.Len())
	i := 0
	contained.Do(func(key, value interface{}) {
		index, _ := key.(int)
		_value, _ := value.(float64)
		values[i] = _value
		indices[i] = index
		i++
	})
	instSparse := data.NewInstance()
	for k, i := range indices {
		if stwv.OutputFormat.Attribute(i).IsNominal() {
			if math.IsNaN(values[k]) {
				instSparse.AddValues("?")
			} else {
				instSparse.AddValues(stwv.OutputFormat.Attributes()[i].Values()[int(values[k])])
			}
		} else if stwv.OutputFormat.Attributes()[i].IsNominal() && !stwv.OutputFormat.Attributes()[i].IsString() {
			instSparse.AddValues(stwv.OutputFormat.Attributes()[i].Values()[i])
		} else {
			instSparse.AddValues(stwv.OutputFormat.Attributes()[i].Name())
		}

	}
	instSparse.SetIndices(indices)
	instSparse.SetRealValues(values)
	instSparse.SetWeight(inst.Weight())
	instSparse.SetNumAttributes(stwv.OutputFormat.NumAttributes())
	return firstCopy, instSparse
}

func (stwv *StringToWordVector) NormalizeInstance(inst *data.Instance, firstCopy int) {
	docLength := float64(0)
	if stwv.AvgDocLength < 0 {
		panic("Average document length not set.")
	}
	// Compute length of document vector
	for j := 0; j < len(inst.RealValues()); j++ {
		if inst.Indices()[j] >= firstCopy {
			docLength += inst.RealValues()[j] * inst.RealValues()[j]
		}
	}
	docLength = math.Sqrt(docLength)
	// Normalize document vector
	for j := 0; j < len(inst.RealValues()); j++ {
		if inst.Indices()[j] >= firstCopy {
			val := inst.RealValues()[j] * stwv.AvgDocLength / docLength
			inst.AddRealValuesIndex(j, val)
			if val == 0 {
				//fmt.Println("Setting value %d to zero", inst.Indices()[j])
				j--
			}
		}
	}

}

func (stwv *StringToWordVector) ConvertedInstances() data.Instances {
	return stwv.OutputFormat
}

func (stwv *StringToWordVector) SetTF_Transformation(set bool) {
	stwv.TF_transformation = set
}

func (stwv *StringToWordVector) SetIDF_Transformation(set bool) {
	stwv.IDF_transformation = set
}

func (stwv *StringToWordVector) SetWordsToKeep(WordsToKeep int) {
	stwv.WordsToKeep = WordsToKeep
}

func (stwv *StringToWordVector) SetOutputsCounts(oc bool) {
	stwv.OutputsCounts = oc
}

func (stwv *StringToWordVector) SetNormalize(norm bool) {
	stwv.Normalize = norm
}

func (stwv *StringToWordVector) SetPerClass(PerClass bool) {
	stwv.PerClass = PerClass
}

func (stwv *StringToWordVector) SetMinTermFreq(mtq rune) {
	stwv.MinTermFreq = mtq
}
