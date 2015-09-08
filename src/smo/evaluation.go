package smo

import (
	datas "github.com/project-mac/src/data"
	"github.com/project-mac/src/utils"
	"math"
	"math/rand"
)

const (
	MarginResolution = 500
	MIN_SF_PROB      = -math.MaxFloat64
)

type Evaluation struct {
	numClasses, numFolds                                      int
	incorrect, correct, unclassified, missingClass, withClass float64
	confusionMatrix                                           [][]float64
	classNames                                                []string
	classIsNominal                                            bool
	classPriors                                               []float64
	classPriorsSum                                            float64
	costMatrix                                                CostMatrix
	totalCost, sumError, sumAbsError, sumClass, sumSqrClass,
	sumPredicted, sumSqrPredicted, sumClassPredicted, sumPriorAbsError,
	sumPriorSqrError, sumKBInfo float64
	marginCounts      []float64
	numTrainCLassVals int
	trainClassVals    []float64
	trainClassWeights []float64
	//priorErrorEstimator, errorEstimator
	sumPriorEntropy float64
	predictions     []float64
	noPriors        bool
}

func NewEvaluation(data datas.Instances) Evaluation {
	var m Evaluation
	m = NewEvaluationDataCM(data, m.costMatrix)
	m.noPriors = false
	return m
}

func NewEvaluationDataCM(data datas.Instances, costMatrix CostMatrix) Evaluation {
	var m Evaluation
	m.noPriors = false
	m.numClasses = data.NumInstances()
	m.numFolds = 1
	m.classIsNominal = data.ClassAttribute().IsNominal()

	if m.classIsNominal {
		m.confusionMatrix = make([][]float64, m.numClasses)
		for i := range m.confusionMatrix {
			m.confusionMatrix[i] = make([]float64, m.numClasses)
		}
		m.classNames = make([]string, m.numClasses)
		for i := range m.classNames {
			m.classNames[i] = data.ClassAttribute().Value(i)
		}
	}
	m.costMatrix = costMatrix
	if &costMatrix == nil {
		if !m.classIsNominal {
			panic("Class has to be nominal if cost matrix given!")
		}
		if m.costMatrix.Size() != m.numClasses {
			panic("Cost matrix not compatible with data!")
		}
	}
	m.classPriors = make([]float64, m.numClasses)
	m.SetPriors(data)
	m.marginCounts = make([]float64, MarginResolution+1)
	return m
}

// Sets the class prior probabilities
func (m *Evaluation) SetPriors(train datas.Instances) {
	m.noPriors = false

	if !m.classIsNominal {
		m.numTrainCLassVals = 0
		m.trainClassVals = nil
		m.trainClassWeights = nil
		//Put to nil the two estimators

		for _, currentInst := range train.Instances() {
			if !currentInst.ClassMissing(train.ClassIndex()) {
				m.addNumericTrainClass(currentInst.ClassValue(train.ClassIndex()), currentInst.Weight())
			}
		}
	} else {
		for i := 0; i < m.numClasses; i++ {
			m.classPriors[i] = 1
		}
		m.classPriorsSum = float64(m.numClasses)
		for _, inst := range train.Instances() {
			if !inst.ClassMissing(train.ClassIndex()) {
				m.classPriors[int(inst.ClassValue(train.ClassIndex()))] += inst.Weight()
				m.classPriorsSum += inst.Weight()
			}
		}
	}
}

// Adds a numeric (non-missing) training class value and weight to the buffer of stored values.
func (m *Evaluation) addNumericTrainClass(classValue, weight float64) {
	if m.trainClassVals == nil {
		m.trainClassVals = make([]float64, 100)
		m.trainClassWeights = make([]float64, 100)
	}

	if m.numTrainCLassVals == len(m.trainClassVals) {
		temp := make([]float64, len(m.trainClassVals)*2)
		copy(temp, m.trainClassVals)
		m.trainClassVals = temp

		temp = make([]float64, len(m.trainClassWeights)*2)
		copy(temp, m.trainClassWeights)
		m.trainClassWeights = temp
	}
	m.trainClassVals[m.numTrainCLassVals] = classValue
	m.trainClassWeights[m.numTrainCLassVals] = weight
	m.numTrainCLassVals++
}

func (m *Evaluation) CrossValidateModel(classifier Classifier, data datas.Instances, numFolds int, random *rand.Rand) {

	//Make a copy of the data we can reorder
	data.Randomizes(random)
	if data.ClassAttribute().IsNominal() {
		data.Stratify(numFolds)
	}

	//Do the folds
	for i := 0; i < numFolds; i++ {
		train := data.TrainCVRand(numFolds, i, random)
		m.SetPriors(train)
		copiedClassifier := classifier
		copiedClassifier.BuildClassifier(train)
		test := data.TestCV(numFolds, i)
		//m.ev
	}
	m.numFolds = numFolds
}

func (m *Evaluation) EvaluateModel(classifier Classifier, data datas.Instances) []float64 {

	predictions := make([]float64, data.NumInstances())

	// Need to be able to collect predictions if appropriate (for AUC)
	for i := 0; i < data.NumInstances(); i++ {

	}

}

// Evaluates the classifier on a single instance and records the prediction
// (if the class is nominal).
func (m *Evaluation) EvaluateModelOnceAndRecordPrediction(classifier Classifier, instance datas.Instance, classIndex, numClasses int) {
	
	classMissing := instance
	pred := 0.0
	classMissing.SetClassMissing(classIndex)
	if m.classIsNominal {
		if m.predictions == nil {
			m.predictions = make([]float64,0)
		}
		dist := classifier.DistributionForInstance(classMissing,numClasses)
		pred := utils.MaxIndex(dist)
		if dist[int(pred)] <= 0 {
			pred = int(instance.MissingValue)
		}
		//m.UpdateStatsForClassfier(dist,instance)
		m.predictions = append(m.predictions,)
	}
	
}

///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
