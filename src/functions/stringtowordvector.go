package functions

import (
	"data"
	"fmt"
	"math"
	"omap"
	"strings"
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
	firstTime bool
	//Contains a mapping of valid words to attribute indexes.
	dictionary map[string]float64
	//True if output instances should contain word frequency rather than boolean 0 or 1.
	outputsCounts bool
	//Contains the number of documents (instances) a particular word appears in.
	//The counts are stored with the same indexing as given by m_Dictionary.
	docsCounts []int
	// Contains the number of documents (instances) in the input format from
	//which the dictionary is created. It is used in IDF transform.
	numInstances int
	// Contains the average length of documents (among the first batch of
	//instances aka training data). This is used in length normalization of
	//documents which will be normalized to average document length.
	avgDocLength float64
	//The default number of words (per class if there is a class attribute
	//assigned) to attempt to keep.
	wordsToKeep int
	//Which transformation
	transformation int
	//The percentage at which to periodically prune the dictionary.
	perdiodicPruningRate int
	//Use normalization or not
	normalize bool
	//whether to operate on per-class basis
	perClass bool
	//the minimum (per-class) word frequency.
	minTermFreq rune
	//The input format for instances.
	inputFormat data.Instances
	//The output format for instances.
	outputFormat data.Instances
}

//New StringToWordVector function with default values
func NewStringToWordVectorInst(inputData data.Instances) StringToWordVector {
	var stwv StringToWordVector
	stwv.dictionary = make(map[string]float64, 0)
	stwv.outputsCounts = false
	stwv.docsCounts = make([]int, 0)
	stwv.avgDocLength = -1
	stwv.wordsToKeep = 3
	stwv.numInstances = -1
	stwv.perdiodicPruningRate = -1
	stwv.minTermFreq = -1
	stwv.perClass = true
	stwv.normalize = true
	stwv.inputFormat = inputData
	stwv.outputFormat = data.NewInstancesWithClassIndex(inputData.ClassIndex())
	stwv.firstTime = true
	stwv.transformation = 0
	return stwv
}

//Start the execution of the function
func (stwv *StringToWordVector) Exec() data.Instances {
	/* TODO: first check that the input format is initialized*/
	//	if stwv.inputFormat != nil {
	//		panic("No input instace defined")
	//	}
	inst := &stwv.inputFormat
	if stwv.firstTime {
		//check if the class attribute is not nominal to turn off the per-class basis
		//fmt.Println(inst.ClassIndex(), "classIndex")
		classAttr := inst.Attributes()[inst.ClassIndex()]
		if classAttr.Type() != data.NOMINAL {
			stwv.perClass = false
		}
		//Determine the dictionary for the input format
		stwv.determineDictionary(inst)
		//Convert all instances without normalize
		fv := make([]data.Instance, 0)
		firstCopy := 0
		for i := 0; i < stwv.numInstances; i++ {
			firstcopy, v := stwv.convertInstancewoDocNorm(stwv.inputFormat.Instances()[i])
			fv = append(fv, v)
			firstCopy = firstcopy
		}
		//fmt.Println(firstCopy, " firstcopy")
		// Need to compute average document length if necessary
		stwv.avgDocLength = 0
		for _, inst := range fv {
			docLength := float64(0)
			for j := 0; j < len(inst.RealValues()); j++ {
				if inst.Indices()[j] >= firstCopy {
					docLength += math.Pow(inst.RealValues()[j], 2)
				}
			}
			stwv.avgDocLength += math.Sqrt(docLength)
		}
		stwv.avgDocLength /= float64(stwv.numInstances)
		// Perform normalization if necessary.
		if stwv.normalize {
			for _, inst := range fv {
				stwv.normalizeInstance(&inst, firstCopy)
			}
		}
		stwv.outputFormat.SetInstances(fv)
	}
	fmt.Println("Done!")
	return stwv.outputFormat
}

func (stwv *StringToWordVector) determineDictionary(inst *data.Instances) {
	/* TODO: see if use a stopwords list*/
	fmt.Println("Determing dictionary!")
	classInd := inst.ClassIndex()
	values := 1
	if stwv.perClass && (classInd != -1) {
		values = len(inst.Attributes()[classInd].Values())
	}
	dicA := make([]map[string]Count, values)
	for i := 0; i < values; i++ {
		dicA[i] = make(map[string]Count, 0)
	}
	// Tokenize all training text into an orderedMap of "words".
	pruneRate := int64((stwv.perdiodicPruningRate / 100) * len(inst.Instances()))
	for i, instance := range inst.Instances() {
		vInd := int(0)
		if stwv.perClass && (classInd != -1) {
			vInd = int(instance.RealValues()[classInd])
		}
		//Iterate through all relevant string attributes of the current instance
		hashtable := make(map[string]int, 0)
		for j := 0; j < instance.NumAttributes(); j++ {
			if !instance.IsMissingValue(j) && inst.Attributes()[j].IsString() {
				// Iterate through tokens, perform stemming, and remove stopwords
				// (if required)
				//fmt.Println(instance.Values())
				words := strings.Fields(instance.Values()[j])
				for _, word := range words {
					_, present := hashtable[word]
					if !present {
						hashtable[word] = 0
					}
					//fmt.Println(word)
					if count, present := dicA[vInd][word]; !present {
						dicA[vInd][word] = Count{1, 0}
					} else {
						count.Count++
						dicA[vInd][word] = count
					}
					//fmt.Println(dicA[vInd][word])
				}
			}
		}
		//updating the docCount for the words that have occurred in this
		//instance(document).
		enumeration := make([]string, len(hashtable))
		m := 0
		for word, _ := range hashtable { //only the words
			enumeration[m] = word
			m++
		}
		for _, word := range enumeration {
			if count, present := dicA[vInd][word]; present {
				count.DocCount++
				//delete(dicA[vInd], word)
				dicA[vInd][word] = count
				//fmt.Println(word, " ",dicA[vInd][word])
			} else {
				panic("Check the code, there must be a word in the dictionary")
			}
		}
		if pruneRate > 0 {
			if int64(i)%pruneRate == 0 && i > 0 {
				for z := 0; z < values; z++ {
					d := make([]string, 1000)
					for word, _ := range dicA[z] {
						count := dicA[z][word]
						if count.Count <= 1 {
							d = append(d, word)
						}
					}
					for _, word := range d {
						delete(dicA[z], word)
					}
				}
			}
		}
		//fmt.Println("new instance-----------------------------------------------------------")
	}
	//fmt.Println(dicA)
	// Figure out the minimum required word frequency
	totalSize := int(0)
	prune := make([]int, values)
	for z := 0; z < values; z++ {
		totalSize += len(dicA[z])
		array := make([]int, len(dicA[z]))
		pos := int(0)
		for word, _ := range dicA[z] {
			count := dicA[z][word]
			array[pos] = count.Count
			pos++
		}
		array = sortArray(array)
		if len(array) < stwv.wordsToKeep {
			// if there aren't enough words, set the threshold to
			// minFreq
			prune[z] = int(stwv.minTermFreq)
		} else {
			// otherwise set it to be at least minFreq
			prune[z] = int(math.Max(float64(stwv.minTermFreq), float64(array[len(array)-stwv.wordsToKeep])))
		}
	}
	// Convert the dictionary into an attribute index
	// and create one attribute per word
	attributes := make([]data.Attribute, 0, totalSize+len(inst.Attributes()))
	// Add the non-converted attributes
	classIndex := int(-1)
	for i, attr := range stwv.inputFormat.Attributes() {
		if !attr.IsString() {
			if inst.ClassIndex() == i {
				classIndex = len(attributes)
			}
			fmt.Println(attr)
			attributes = append(attributes, attr)
		}
	}
	// Add the word vector attributes (eliminating duplicates
	// that occur in multiple classes)
	newDic := make(map[string]float64, 0)
	index := len(attributes)
	//fmt.Println("-----------/********************//////////////////", index)
	for z := 0; z < values; z++ {
		for word, _ := range dicA[z] {
			count := dicA[z][word]
			if count.Count >= prune[z] {
				if _, present := newDic[word]; !present {
					newDic[word] = float64(index)
					index++
					att := data.NewAttribute()
					att.SetName(word)
					att.SetType(data.STRING)
					attributes = append(attributes, att)
				}
			}
		}
	}
	// Compute document frequencies
	stwv.docsCounts = make([]int, len(attributes))
	//idx := 0
	for word, idx := range newDic {
		docsCount := 0
		for j := 0; j < values; j++ {
			if count, present := dicA[j][word]; present {
				docsCount += count.DocCount
			}
		}
		stwv.docsCounts[int(idx)] = docsCount
		//idx++

	}
	//fmt.Println(newDic)
	stwv.dictionary = newDic
	////fmt.Println("numInst", len(inst.Instances()))
	stwv.numInstances = len(inst.Instances())
	stwv.outputFormat = data.NewInstances()
	stwv.outputFormat.SetAttributes(attributes)
	stwv.outputFormat.SetClassIndex(classIndex)
}

func (stwv *StringToWordVector) convertInstancewoDocNorm(inst data.Instance) (int, data.Instance) {

	// Convert the instance into a sorted set of indexes
	contained := omap.NewIntKeyed()
	mapKeys := make([]float64, 0)
	// Copy all non-converted attributes from input to output
	firstCopy := 0

	for i, _ := range stwv.inputFormat.Attributes() {
		//fmt.Println("input attrs: ", i)
		if !stwv.inputFormat.Attributes()[i].IsString() {
			// Add simple nominal and numeric attributes directly
			if inst.RealValues()[i] != 0 {
				contained.Insert(firstCopy, inst.RealValues()[i])
				mapKeys = append(mapKeys, float64(firstCopy))
				firstCopy++
			} else {
				firstCopy++
			}
		} else if inst.IsMissingValue(i) {
			//fmt.Println("print 1.2")
			contained.Insert(firstCopy, inst.MissingValue)
			mapKeys = append(mapKeys, float64(firstCopy))
			firstCopy++
		} else if stwv.inputFormat.Attributes()[i].IsString() {
			//if i have to implement the range selector then code this part
		}
	}
	//Copy the converted attributes
	//fmt.Println("print 2.0" , inst.NumAttributes())
	for j := 0; j < inst.NumAttributes(); j++ {
		//fmt.Println("print 2.0.1" , stwv.inputFormat.Attributes()[1].IsString())
		if stwv.inputFormat.Attributes()[j].IsString() && inst.IsMissingValue(j) == false {
			//fmt.Println("print 2")
			words := strings.Fields(inst.Values()[j])
			//fmt.Println(stwv.dictionary)
			fmt.Println("------------------------------------------------")
			for _, word := range words {
				//fmt.Println("print 3", idx)
				if index, present := stwv.dictionary[word]; present {
					if stwv.outputsCounts {
						if count, isthere := contained.Find(index); isthere {
							if count, ok := count.(float64); ok { //type assertion
								contained.Insert(int(index), count+1)
								mapKeys = append(mapKeys, index)
							}
						} else {
							//fmt.Println(index)
							contained.Insert(int(index), float64(1))
							mapKeys = append(mapKeys, index)
						}
					} else {
						//fmt.Println(index)
						contained.Insert(int(index), float64(1))
						mapKeys = append(mapKeys, index)
					}
				}
			}
		}
	}
	//fmt.Println(mapKeys)

	//TF_freq transform
	if stwv.transformation == TF {
		contained.Do(func(key, value interface{}) {
			k, _ := key.(int)
			val, _ := value.(float64)
			if k >= firstCopy {
				val = math.Log(val + 1)
				contained.Insert(k, val)
			}
		})
	}
	//IDF_freq transform
	if stwv.transformation == IDF {
		contained.Do(func(key, value interface{}) {
			k, _ := key.(int)
			val, _ := value.(float64)
			if k >= firstCopy {
				val = val * math.Log(float64(stwv.numInstances)/float64(stwv.docsCounts[k]))
				contained.Insert(k, val)
			}
		})
	}
	//TF_IDF_freq transform
	if stwv.transformation == TF_IDF {
		contained.Do(func(key, value interface{}) {
			k, _ := key.(int)
			val, _ := value.(float64)
			if k >= firstCopy {
				val = (val * math.Log(float64(stwv.numInstances)/float64(stwv.docsCounts[k]))) * math.Log(val+1)
				contained.Insert(k, val)
			}
		})
	}
	// contained.Do(func(key, value interface{}) {
	// 	fmt.Println(key, " ", value)
	// })
	// Convert the set to structures needed to create a sparse instance.
	values := make([]float64, contained.Len())
	indices := make([]int, contained.Len())
	i := 0
	fmt.Println(contained.Len())
	contained.Do(func(key, value interface{}) {
		index, _ := key.(int)
		_value, _ := value.(float64)
		values[i] = _value
		indices[i] = index
		i++
	})
	instSparse := data.NewInstance()
	fmt.Println(indices)
	fmt.Println(values)
	for k, i := range indices {
		if stwv.outputFormat.Attributes()[i].IsNominal() {
			fmt.Println(stwv.outputFormat.Attributes()[i].Name(), values[i], i, k)
			instSparse.AddValues(stwv.outputFormat.Attributes()[i].Values()[int(values[k])])
		} else if stwv.outputFormat.Attributes()[i].IsNominal() && !stwv.outputFormat.Attributes()[i].IsString() {
			instSparse.AddValues(stwv.outputFormat.Attributes()[i].Values()[i])
		} else {
			instSparse.AddValues(stwv.outputFormat.Attributes()[i].Name())
		}

	}
	instSparse.SetIndices(indices)
	instSparse.SetRealValues(values)
	instSparse.SetWeight(inst.Weight())
	instSparse.SetNumAttributes(len(values))
	return firstCopy, instSparse
}

func sortArray(array []int) []int {
	i, j, h := int(0), int(0), int(0)
	N := len(array) - 1
	for h = 1; h <= N/9; h = 3*h + 1 {
	}

	for ; h > 0; h /= 3 {
		for i = h + 1; i <= N; i++ {
			v := array[i]
			j = i
			for j > h && array[j-h] > v {
				array[j] = array[j-h]
				j -= h
			}
			array[j] = v
		}
	}
	for _, v := range array {
		fmt.Print(v, " ")
	}
	fmt.Println(array)
	return array

}

func (stwv *StringToWordVector) normalizeInstance(inst *data.Instance, firstCopy int) {

	docLength := float64(0)
	if stwv.avgDocLength < 0 {
		panic("Average document length not set.")
	}
	//	fmt.Println("valores: ", inst.RealValues())
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
			val := inst.RealValues()[j] * stwv.avgDocLength / docLength
			inst.AddRealValuesIndex(j, val)
			if val == 0 {
				fmt.Println("Setting value %d to zero", inst.Indices()[j])
				j--
			}
		}
	}

}

func (stwv *StringToWordVector) ConvertedInstances() data.Instances {
	return stwv.outputFormat
}
