package smo

import (
	"bytes"
	"fmt"
	datas "github.com/project-mac/src/data"
	"github.com/project-mac/src/utils"
	"math"
	"math/rand"
	"strconv"
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
	totalCost, sumErr, sumAbsErr, sumClass, sumSqrClass, sumSqrErr,
	sumPredicted, sumSqrPredicted, sumClassPredicted, sumPriorAbsErr,
	sumPriorSqrErr, sumKBInfo float64
	marginCounts      []float64
	numTrainCLassVals int
	trainClassVals    []float64
	trainClassWeights []float64
	//priorErrorEstimator, errorEstimator
	sumPriorEntropy  float64
	predictions      []NominalPrediction
	noPriors         bool
	sumSchemeEntropy float64
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
	m.numClasses = data.NumClasses()
	m.numFolds = 1
	m.classIsNominal = data.ClassAttribute().IsNominal()

	if m.classIsNominal {
		m.confusionMatrix = make([][]float64, m.numClasses)
		for i := 0; i < m.numClasses; i++ {
			m.confusionMatrix[i] = make([]float64, m.numClasses)
		}
		m.classNames = make([]string, m.numClasses)
		for i := 0; i < m.numClasses; i++ {
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
				m.addNumericTrainClass(currentInst.ClassValueNotSparse(train.ClassIndex()), currentInst.Weight())
			}
		}
	} else {
		for i := 0; i < m.numClasses; i++ {
			m.classPriors[i] = 1
		}
		m.classPriorsSum = float64(m.numClasses)
		for _, inst := range train.Instances() {
			if !inst.ClassMissing(train.ClassIndex()) {
				m.classPriors[int(inst.ClassValueNotSparse(train.ClassIndex()))] += inst.Weight()
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

func (m *Evaluation) CrossValidateModel(classifier Classifier, data datas.Instances, numFolds int, random *rand.Rand) Classifier {
	var class Classifier

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
		fmt.Println("-----------------------------------------------------------------------")
		test := data.TestCV(numFolds, i)
		m.EvaluateModel(copiedClassifier, test)
		class = copiedClassifier
	}
	m.numFolds = numFolds
	return class
}

func (m *Evaluation) EvaluateModel(classifier Classifier, data datas.Instances) []float64 {

	predictions := make([]float64, data.NumInstances())

	// Need to be able to collect predictions if appropriate (for AUC)
	for i := 0; i < data.NumInstances(); i++ {
		predictions[i] = m.EvaluateModelOnceAndRecordPrediction(classifier, data.InstanceNoPtr(i), &data)
	}
	fmt.Println(predictions, "predictions")
	return predictions
}

// Evaluates the classifier on a single instance and records the prediction
// (if the class is nominal).
func (m *Evaluation) EvaluateModelOnceAndRecordPrediction(classifier Classifier, instance datas.Instance, data *datas.Instances) float64 {
	classMissing := instance
	pred := 0.0
	classMissing.SetClassMissing(data.ClassIndex())
	if m.classIsNominal {
		if m.predictions == nil {
			m.predictions = make([]NominalPrediction, 0)
		}
		dist := classifier.DistributionForInstance(classMissing, 0, data.NumClasses())
		pred := utils.MaxIndex(dist)
		if dist[int(pred)] <= 0 {
			pred = int(math.NaN())
		}
		m.updateStatsForClassfier(dist, instance, data)
		m.predictions = append(m.predictions, NewNominalPredictionWeight(instance.ValueSparse(data.ClassIndex()), dist, instance.Weight()))
	} else {
		/* THIS IS VERY IMPORTANT!!!!!!!!!!
		*
		*  In this implementation I will always assume that the class is nominal due to the nature of the problem we're treating.
		*  It's a classification problem not a prediction one.
		*  IN OTHER CASE THAN IMPLEMENT THIS PART OF THE PROGRAM
		 */
		//pred = classifier.ClassifyInstances(classMissing, data)
		//m.updateStatsForPredictor(pred,instance)
		println("Nothing to do for the moment!!!!!!!!!!")
	}
	return pred
}

func (m *Evaluation) UpdateStatsForClassfier(predictedDistribution []float64, instance datas.Instance, data *datas.Instances) {
	m.updateStatsForClassfier(predictedDistribution, instance, data)
}

// Updates all the statistics about a classifiers performance for the current
// test instance.
func (m *Evaluation) updateStatsForClassfier(predictedDistribution []float64, instance datas.Instance, data *datas.Instances) {
	actualClass := int(instance.ValueSparse(data.ClassIndex()))
	if !instance.ClassMissing(data.ClassIndex()) {
		m.updateMargins(predictedDistribution, actualClass, instance.Weight(), data)

		// Determine the predicted class (doesn't detect multiple
		// classifications)
		predictedClass := -1
		bestProb := 0.0
		for i := 0; i < data.NumClasses(); i++ {
			if predictedDistribution[i] > bestProb {
				predictedClass = i
				bestProb = predictedDistribution[i]
			}
		}
		m.withClass += instance.Weight()

		// Determine misclassification cost
		if m.costMatrix.NotNil {
			if predictedClass < 0 {
				// For missing predictions, we assume the worst possible cost.
				// This is pretty harsh.
				// Perhaps we could take the negative of the cost of a correct
				// prediction (-m_CostMatrix.getElement(actualClass,actualClass)),
				// although often this will be zero

				m.totalCost += instance.Weight() * m.costMatrix.GetMaxCost(actualClass, instance)
			} else {
				m.totalCost += instance.Weight() * m.costMatrix.GetElement(actualClass, int(predictedClass), instance)
			}
		}

		// Update counts when no class was predicted
		if predictedClass < 0 {
			m.unclassified += instance.Weight()
			return
		}

		predictedProb := math.Max(MIN_SF_PROB, predictedDistribution[actualClass])
		priorProb := math.Max(MIN_SF_PROB, m.classPriors[actualClass]/m.classPriorsSum)
		if predictedProb >= priorProb {
			m.sumKBInfo += (math.Log2(predictedProb) - math.Log2(priorProb)) * instance.Weight()
		} else {
			m.sumKBInfo -= (math.Log2(1.0-predictedProb) - math.Log2(1.0-priorProb)) * instance.Weight()
		}

		m.sumSchemeEntropy -= math.Log2(predictedProb) * instance.Weight()
		m.sumPriorEntropy -= math.Log2(priorProb) * instance.Weight()

		m.updateNumericScores(predictedDistribution, m.makeDistribution(instance.ClassValue(data.ClassIndex())), instance.Weight())

		// Update other stats
		m.confusionMatrix[actualClass][predictedClass] += instance.Weight()
		if predictedClass != actualClass {
			m.incorrect += instance.Weight()
		} else {
			m.correct += instance.Weight()
		}
	} else {
		m.missingClass += instance.Weight()
	}

}

func (m *Evaluation) updateStatsForPredictor(predictedValue float64, instance datas.Instance, data *datas.Instances) {
	if !instance.ClassMissing(data.ClassIndex()) {

		// Update stats
		m.withClass += instance.Weight()
		if instance.IsMissingValue(int(predictedValue)) {
			m.unclassified += instance.Weight()
			return
		}
		m.sumClass += instance.Weight() * instance.ClassValue(data.ClassIndex())
		m.sumSqrClass += instance.Weight() * instance.ClassValue(data.ClassIndex()) * instance.ClassValue(data.ClassIndex())
		m.sumClassPredicted += instance.Weight() * instance.ClassValue(data.ClassIndex()) * predictedValue
		m.sumPredicted += instance.Weight() * predictedValue
		m.sumSqrPredicted += instance.Weight() * predictedValue * predictedValue
	}
}

// Convert a single prediction into a probability distribution with all zero
// probabilities except the predicted value which has probability 1.0.
func (m *Evaluation) makeDistribution(predictedClass float64) []float64 {
	result := make([]float64, m.numClasses)
	if math.IsNaN(predictedClass) {
		return result
	}
	if m.classIsNominal {
		result[int(predictedClass)] = 1.0
	} else {
		result[0] = predictedClass
	}
	return result
}

// Update the numeric accuracy measures. For numeric classes, the accuracy is
// between the actual and predicted class values. For nominal classes, the
// accuracy is between the actual and predicted class probabilities.

func (m *Evaluation) updateNumericScores(predicted, actual []float64, weight float64) {
	var diff float64
	sumErr, sumAbsErr, sumSqrErr, sumPriorAbsErr, sumPriorSqrErr := 0.0, 0.0, 0.0, 0.0, 0.0
	for i := 0; i < m.numClasses; i++ {
		diff = predicted[i] - actual[i]
		sumErr += diff
		sumAbsErr += math.Abs(diff)
		sumSqrErr += diff * diff
		diff = (m.classPriors[i] / m.classPriorsSum) - actual[i]
		sumPriorAbsErr += math.Abs(diff)
		sumPriorSqrErr += diff * diff
	}
	m.sumErr += weight * sumErr / float64(m.numClasses)
	m.sumAbsErr += weight * sumAbsErr / float64(m.numClasses)
	m.sumSqrErr += weight * sumSqrErr / float64(m.numClasses)
	m.sumPriorAbsErr += weight * sumPriorAbsErr / float64(m.numClasses)
	m.sumPriorSqrErr += weight * sumPriorSqrErr / float64(m.numClasses)
}

// Update the cumulative record of classification margins
func (m *Evaluation) updateMargins(predictedDistribution []float64, actualClass int, weight float64, data *datas.Instances) {
	probActual := predictedDistribution[actualClass]
	probNext := 0.0

	for i := 0; i < data.NumClasses(); i++ {
		if i != actualClass && predictedDistribution[i] > probNext {
			probNext = predictedDistribution[i]
		}
	}
	margin := probActual - probNext
	bin := int((margin + 1.0) / 2.0 * MarginResolution)
	m.marginCounts[bin] += weight
}

func (m *Evaluation) ToSummaryString(title string, printComplexityStatistics bool) string {
	var text string

	if printComplexityStatistics && m.noPriors {
		printComplexityStatistics = false
		println("Priors disabled, cannot print complexity statistics!")
	}

	fmt.Println(m.withClass)
	fmt.Println(m.classIsNominal)

	text += title + "\n"
	if m.withClass > 0 {
		if m.classIsNominal {
			text += "Correctly Classified Instances     "
			text += strconv.FormatInt(int64(m.correct), 10) + "     " + strconv.FormatFloat(100*m.correct/m.withClass, 'f', 4, 64) + " %\n"
			text += "Incorrectly Classified Instances   "
			text += strconv.FormatInt(int64(m.incorrect), 10) + "     " + strconv.FormatFloat(100*m.incorrect/m.withClass, 'f', 4, 64) + " %\n"
			text += "Kappa statistic                    "
			text += strconv.FormatFloat(m.kappa(), 'f', 4, 64) + " %\n"
		}
	}
	return text
}

func (m *Evaluation) kappa() float64 {
	sumRows := make([]float64, len(m.confusionMatrix))
	sumColumns := make([]float64, len(m.confusionMatrix))
	sumOfWeights := 0.0
	for i := range m.confusionMatrix {
		for j := range m.confusionMatrix[i] {
			sumRows[i] += m.confusionMatrix[i][j]
			sumColumns[j] += m.confusionMatrix[i][j]
			sumOfWeights += m.confusionMatrix[i][j]
		}
	}
	correct, chanceAgreement := 0.0, 0.0
	for i := range m.confusionMatrix {
		chanceAgreement += (sumRows[i] * sumColumns[i])
		correct += m.confusionMatrix[i][i]
	}
	chanceAgreement /= (sumOfWeights * sumOfWeights)
	correct /= sumOfWeights

	if chanceAgreement < 1 {
		return (correct - chanceAgreement) / (1 - chanceAgreement)
	} else {
		return 1
	}
}

func (m *Evaluation) ToClassDetailsString(title string) string {
	if !m.classIsNominal {
		panic("Evaluation: No confusion matrix possible!")
	}

	text := title + "\n               TP Rate   FP Rate" + "   Precision   Recall" + "  F-Measure  Class\n"
	for i := 0; i < m.numClasses; i++ {
		text += "               " + strconv.FormatFloat(m.truePositiveRate(i), 'f', 3, 64) + "      "
		text += strconv.FormatFloat(m.falsePositiveRate(i), 'f', 3, 64) + "      "
		text += strconv.FormatFloat(m.precision(i), 'f', 3, 64) + "      "
		text += strconv.FormatFloat(m.recall(i), 'f', 3, 64) + "     "
		text += strconv.FormatFloat(m.fMeasure(i), 'f', 3, 64) + "   "
		text += m.classNames[i] + "   \n"
	}

	text += "Weighted Avg.  " + strconv.FormatFloat(m.weightedTruePositiveRate(), 'f', 3, 64)
	text += "      " + strconv.FormatFloat(m.weightedFalsePositiveRate(), 'f', 3, 64)
	text += "      " + strconv.FormatFloat(m.weightedPrecision(), 'f', 3, 64)
	text += "      " + strconv.FormatFloat(m.weightedRecall(), 'f', 3, 64)
	text += "     " + strconv.FormatFloat(m.weightedFMeasure(), 'f', 3, 64)
	text += "\n"
	return text
}

func (m *Evaluation) ToMatrixString(title string) string {

	text := ""

	IDChars := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
		'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
	var IDWidth int
	fractional := false

	if !m.classIsNominal {
		panic("Evaluation: No confusion matrix possible! Class not nominal.")
	}

	// Find the maximum value in the matrix
	// and check for fractional display requirement
	maxval := 0.0
	for i := 0; i < m.numClasses; i++ {
		for j := 0; j < m.numClasses; j++ {
			current := m.confusionMatrix[i][j]
			if current < 0 {
				current *= -10
			}
			if current > maxval {
				maxval = current
			}
			fract := current - math_Rint(current)
			if !fractional && ((math.Log(fract) / math.Log(10)) >= -2) {
				fractional = true
			}
		}
	}

	f := func() float64 {
		if fractional {
			return 3
		} else {
			return 0
		}
	}
	IDWidth = int(1 + math.Max(math.Log(maxval)/math.Log(10)+f(), math.Log(float64(m.numClasses))/math.Log(float64(len(IDChars)))))
	text += title + "\n"
	for i := 0; i < m.numClasses; i++ {
		if fractional {
			text += fmt.Sprint(" ", num2ShortID(i, IDChars, IDWidth-3), "   ")
		} else {
			text += fmt.Sprint(" ", num2ShortID(i, IDChars, IDWidth))
		}
	}
	text += "   <-- classified as\n"
	for i := 0; i < m.numClasses; i++ {
		for j := 0; j < m.numClasses; j++ {
			f := func() int {
				if fractional {
					return 2
				} else {
					return 0
				}
			}
			text += fmt.Sprint(" ", strconv.FormatFloat(m.confusionMatrix[i][j], 'f', f(), 64))
		}
		text += fmt.Sprint(" | ", num2ShortID(i, IDChars, IDWidth), " = ", m.classNames[i], "\n")
	}
	return text
}

func num2ShortID(num int, IDChars []byte, IDWidth int) string {
	ID := make([]byte, IDWidth)
	var i int

	for i = IDWidth - 1; i >= 0; i-- {
		ID[i] = IDChars[num%len(IDChars)]
		num = num/len(IDChars) - 1
		if num < 0 {
			break
		}
	}
	for i--; i >= 0; i-- {
		ID[i] = ' '
	}
	b := bytes.NewBuffer(ID)
	return b.String()
}

func math_Rint(a float64) float64 {
	twoToThe52 := uint64(1) << uint64(52)
	sign := math.Copysign(1.0, a)
	a = math.Abs(a)

	if a < float64(twoToThe52) {
		a = (float64(twoToThe52) + a) - float64(twoToThe52)
	}
	return sign * a
}

func (m *Evaluation) truePositiveRate(classIndex int) float64 {
	correct, total := 0.0, 0.0
	for j := 0; j < m.numClasses; j++ {
		if j == classIndex {
			correct += m.confusionMatrix[classIndex][j]
		}
		total += m.confusionMatrix[classIndex][j]
	}
	if total == 0 {
		return 0
	}
	return correct / total
}

func (m *Evaluation) falsePositiveRate(classIndex int) float64 {
	incorrect, total := 0.0, 0.0
	for i := 0; i < m.numClasses; i++ {
		if i != classIndex {
			for j := 0; j < m.numClasses; j++ {
				if j == classIndex {
					incorrect += m.confusionMatrix[i][j]
				}
				total += m.confusionMatrix[i][j]
			}
		}
	}
	if total == 0 {
		return 0
	}
	return incorrect / total
}

func (m *Evaluation) precision(classIndex int) float64 {
	correct, total := 0.0, 0.0
	for i := 0; i < m.numClasses; i++ {
		if i == classIndex {
			correct += m.confusionMatrix[i][classIndex]
		}
		total += m.confusionMatrix[i][classIndex]
	}
	if total == 0 {
		return 0
	}
	return correct / total
}

func (m *Evaluation) recall(classIndex int) float64 {
	return m.truePositiveRate(classIndex)
}

func (m *Evaluation) fMeasure(classIndex int) float64 {
	precision := m.precision(classIndex)
	recall := m.recall(classIndex)
	if (precision + recall) == 0 {
		return 0
	}
	return 2 * precision * recall / (precision + recall)
}

func (m *Evaluation) weightedTruePositiveRate() float64 {
	classCounts := make([]float64, m.numClasses)
	classCountSum := 0.0

	for i := 0; i < m.numClasses; i++ {
		for j := 0; j < m.numClasses; j++ {
			classCounts[i] += m.confusionMatrix[i][j]
		}
		classCountSum += classCounts[i]
	}

	truePosTotal := 0.0
	for i := 0; i < m.numClasses; i++ {
		temp := m.truePositiveRate(i)
		truePosTotal += (temp * classCounts[i])
	}
	return truePosTotal / classCountSum
}

func (m *Evaluation) weightedFalsePositiveRate() float64 {
	classCounts := make([]float64, m.numClasses)
	classCountSum := 0.0

	for i := 0; i < m.numClasses; i++ {
		for j := 0; j < m.numClasses; j++ {
			classCounts[i] += m.confusionMatrix[i][j]
		}
		classCountSum += classCounts[i]
	}

	falsePosTotal := 0.0
	for i := 0; i < m.numClasses; i++ {
		temp := m.falsePositiveRate(i)
		falsePosTotal += (temp * classCounts[i])
	}
	return falsePosTotal / classCountSum
}

func (m *Evaluation) weightedPrecision() float64 {
	classCounts := make([]float64, m.numClasses)
	classCountSum := 0.0

	for i := 0; i < m.numClasses; i++ {
		for j := 0; j < m.numClasses; j++ {
			classCounts[i] += m.confusionMatrix[i][j]
		}
		classCountSum += classCounts[i]
	}

	precisionTotal := 0.0
	for i := 0; i < m.numClasses; i++ {
		temp := m.precision(i)
		precisionTotal += (temp * classCounts[i])
	}
	return precisionTotal / classCountSum
}

func (m *Evaluation) weightedRecall() float64 {
	return m.weightedTruePositiveRate()
}

func (m *Evaluation) weightedFMeasure() float64 {
	classCounts := make([]float64, m.numClasses)
	classCountSum := 0.0

	for i := 0; i < m.numClasses; i++ {
		for j := 0; j < m.numClasses; j++ {
			classCounts[i] += m.confusionMatrix[i][j]
		}
		classCountSum += classCounts[i]
	}

	fMeasureTotal := 0.0
	for i := 0; i < m.numClasses; i++ {
		temp := m.fMeasure(i)
		fMeasureTotal += (temp * classCounts[i])
	}
	return fMeasureTotal / classCountSum
}
