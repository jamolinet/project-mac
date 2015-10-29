package smo

import (
	//"fmt"
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/functions"
	"github.com/project-mac/src/utils"
	"math"
	"math/rand"
)

const (
	//Filter: Normalize training data
	FILTER_NORMALIZE = 0
	//Filter: Standardize training data
	FILTER_STANDARDIZE = 1
	//Filter: No normalization/standardization
	FILTER_NONE = 2
)

// Implements John Platt's sequential minimal optimization algorithm for training a support vector classifier.
type SMO struct {
	//The binary classifier(s)
	Classifiers [][]BinarySMO

	//The complexity parameter
	C_C float64

	//Epsilon for rounding
	Eps float64

	//Tolerance for accuracy of result
	Tol float64

	//Whether to normalize/standardize/neither
	FilterType int

	/** Remains other attributes to declare*/
	/**********/ /////
	//The Filter used to make attributes numeric
	NominalToBinary functions.NominalToBinary

	//The Filter used to standardize/normalize all values
	FilterS functions.Standardize
	FilterN functions.Normalize
	Filter  functions.Filter

	//The Filter used to get rid of Missing values
	Missing functions.ReplaceMissingValues

	//The class index from the training data
	ClassIndex int

	//The class attribute
	classAttribute data.Attribute

	//whether the Kernel_ is a linear one
	kernelIsLinear bool

	//	Turn off all checks and conversions? Turning them off assumes
	//      that data is purely numeric, doesn't contain any Missing values,
	//      and has a nominal class. Turning them off also means that
	//      no header information will be stored if the machine is linear.
	//      Finally, it also assumes that no instance has a weight equal to 0
	checksTurnedOff bool

	//Precision constant for updating sets
	Del float64

	//Whether logistic models are to be fit
	FitLogisticModels bool

	//The number of folds for the internal cross-validation
	numFolds int

	//The random number seed
	randomSeed int

	//the Kernel_ to use
	Kernel_ PolyKernel
}

func NewSMO(Kernel_ PolyKernel) SMO {
	var smo SMO
	smo.Classifiers = nil
	smo.C_C = 1.0
	smo.Eps = 1.0e-12
	smo.Tol = 1.0e-3
	smo.FilterType = FILTER_NORMALIZE
	smo.ClassIndex = -1
	smo.kernelIsLinear = true
	smo.Del = -math.MaxFloat64
	smo.FitLogisticModels = false
	smo.numFolds = -1
	smo.randomSeed = 1
	smo.Kernel_ = Kernel_
	return smo
}

func (m *SMO) SetEps(Eps float64) {
	m.Eps = Eps
}

func (m *SMO) SetC(c float64) {
	m.C_C = c
}

func (m *SMO) SetTolerance(Tol float64) {
	m.Tol = Tol
}

func (m *SMO) SetSeed(seed int) {
	m.randomSeed = seed
}

func (m *SMO) SetNumFolds(folds int) {
	m.numFolds = folds
}

func (m *SMO) SetFitLogistic(l bool) {
	m.FitLogisticModels = l
}

func (m *SMO) SetNormalize(n bool) {
	if n {
		m.FilterType = FILTER_NORMALIZE
	} else {
		m.FilterType = FILTER_STANDARDIZE
	}
}

func (m *SMO) TurnChecksOn() {
	m.checksTurnedOff = false
}

func (m *SMO) TurnChecksOff() {
	m.checksTurnedOff = true
}

func (m *SMO) BuildClassifier(insts data.Instances) {
	//fmt.Println(len(insts.Instances()))
	if !m.checksTurnedOff {
		//Remove instances with Missing value

		insts.DeleteWithMissingClass()
		//fmt.Println(len(insts.Instances()), "Missing class")

		/* Removes all the instances with weight equal to 0.
		   MUST be done since condition (8) of Keerthi's paper
		   is made with the assertion Ci > 0 (See equation (3a). */
		data := data.NewInstancesWithInst(insts, insts.NumInstances())
		//fmt.Println(len(data.Instances()), "Missing class")
		for _, instance := range insts.Instances() {
			if instance.Weight() > 0 {
				data.Add(instance)
			}
		}
		if data.NumInstances() == 0 {
			panic("No training instances left after removing " +
				"instances with weight 0!")
		}
		insts = data //all ok so far
	}

	if !m.checksTurnedOff {
		m.Missing = functions.NewReplacingMissingValues()
		m.Missing.SetInputFormat(insts)
		m.Missing.Exec(insts)
		m.Missing.BatchFinished()
		insts = m.Missing.OutputAll()
	} //ok so far

	//It can handle numeric data, so:
	onlyNumeric := true
	if !m.checksTurnedOff {
		for i, att := range insts.Attributes() {
			if i != insts.ClassIndex() {
				if !att.IsNumeric() {
					onlyNumeric = false
					break
				}
			}
		}
	}
	//fmt.Println(onlyNumeric,"numeric")
	if !onlyNumeric {
		m.NominalToBinary = functions.NewNominalToBinary()
		m.NominalToBinary.SetInputFormat(insts)
		m.NominalToBinary.Exec(insts)
		insts = m.NominalToBinary.OutputAll()
	} else {
		m.NominalToBinary.IsNil = ""
	}

	//fmt.Println(insts.NumInstances(), "len")

	if m.FilterType == FILTER_STANDARDIZE {
		m.FilterS = functions.NewStandardize()
		m.FilterS.SetInputFormat(insts)
		m.FilterS.Exec(insts)
		insts = m.FilterS.OutputAll()
	} else if m.FilterType == FILTER_NORMALIZE {
		m.FilterN = functions.NewNormalize()
		m.FilterN.SetInputFormat(insts)
		m.FilterN.Exec(insts)
		insts = m.FilterN.OutputAll()
	}
	//	for _, in := range insts.Instances() {
	//		fmt.Println(in.RealValues(), in.NumAttributesTest(), "--1")
	//		fmt.Println(in.Indices(), "--2")
	//	}
	m.ClassIndex = insts.ClassIndex()
	m.classAttribute = insts.ClassAttribute()
	//This Kernel_ will always be linear in this case
	m.kernelIsLinear = true

	// Generate subsets representing each class
	subsets := make([]data.Instances, insts.NumClasses())
	for i := range subsets {
		subsets[i] = data.NewInstancesWithInst(insts, insts.NumInstances())
	}

	for _, inst := range insts.Instances() {
		subsets[int(inst.ClassValue(insts.ClassIndex()))].Add(inst)
	}
				//ok so far
	// Build the binary Classifiers
	m.Classifiers = make([][]BinarySMO, insts.NumClasses())
	for i := range m.Classifiers {
		m.Classifiers[i] = make([]BinarySMO, insts.NumClasses())
	}

	for i := 0; i < insts.NumClasses(); i++ {
		for j := i + 1; j < insts.NumClasses(); j++ {
			m.Classifiers[i][j] = NewBinarySMO()
			m.Classifiers[i][j].SetSMO(m)
			m.Classifiers[i][j].SetKernel(m.Kernel_)
			_data := data.NewInstancesWithInst(insts, insts.NumInstances())
			for k := 0; k < subsets[i].NumInstances(); k++ {
				_data.Add(subsets[i].InstanceNoPtr(k))
			}
			for k := 0; k < subsets[j].NumInstances(); k++ {
				_data.Add(subsets[j].InstanceNoPtr(k))
			}
			_data.Randomize(m.randomSeed)
			//ok so far	
			m.Classifiers[i][j].buildClassifier(_data, i, j, m.FitLogisticModels, m.numFolds, m.randomSeed)
		}
	}
	//fmt.Println("new fold")
}

// Estimates class probabilities for given instance.
//The numClasses parameter is mandatory
func (m *SMO) DistributionForInstance(inst data.Instance, numClasses int) []float64 {

	//Filter instance
//	if !m.checksTurnedOff {
//		m.Missing.Input(inst)
//		//fmt.Println(m.Missing.ModesAndMeans())
//		m.Missing.BatchFinished()
//		inst = m.Missing.Output()
//	}
		
	if m.NominalToBinary.IsNil == "no" {
		//fmt.Println("m.NominalToBinary.IsNil ==")
		m.NominalToBinary.Input(inst)
		//No need to call batchFinished(), the queue never gonna be empty
		inst = m.NominalToBinary.Output()
	}

	if m.FilterS.NotNil() {
		m.FilterS.Input(inst)
		m.FilterS.BatchFinished()
		inst = m.FilterS.Output()
	} else if m.FilterN.NotNil() {
		m.FilterN.Input(inst)
		m.FilterN.BatchFinished()
		inst = m.FilterN.Output()
	}
	
	
	if !m.FitLogisticModels {
		result := make([]float64, numClasses)
		for i := 0; i < numClasses; i++ {
			for j := i + 1; j < numClasses; j++ {
				if m.Classifiers[i][j].Alpha != nil || m.Classifiers[i][j].SparseWeights != nil {
					output := m.Classifiers[i][j].SVMOutput(-1, inst)
					if output > 0 {
						result[j] += 1
					} else {
						result[i] += 1
					}
				}
			}
		}
		sum := 0.0
		for i := 0; i < len(result); i++ {
			sum += result[i]
		}
		utils.Normalize(&result, sum)
		//fmt.Println(result, "distribution after normalize",sum)
		return result
	} else {
		// We only need to do pairwise coupling if there are more
		// then two classes.
		if numClasses == 2 {
			newInst := make([]float64, 2)
			newInst[0] = m.Classifiers[0][1].SVMOutput(-1, inst)
			newInst[1] = math.NaN()
			//newInst[1] = inst.MissingValue
			return m.Classifiers[0][1].logistic.DistributionForInstance(data.NewInstanceWeightValues(1, newInst))
		}
		r, n := make([][]float64, numClasses), make([][]float64, numClasses)
		for i := range r {
			n[i] = make([]float64, numClasses)
			r[i] = make([]float64, numClasses)
		}
		for i := 0; i < numClasses; i++ {
			for j := i + 1; j < numClasses; j++ {
				if m.Classifiers[i][j].Alpha != nil || m.Classifiers[i][j].SparseWeights != nil {
					newInst := make([]float64, 2)
					newInst[0] = m.Classifiers[0][1].SVMOutput(-1, inst)
					newInst[1] = math.NaN()
					//newInst[1] = inst.MissingValue
					r[i][j] = m.Classifiers[i][j].logistic.DistributionForInstance(data.NewInstanceWeightValues(1, newInst))[0]
					n[i][j] = m.Classifiers[i][j].sumOfWeights
				}
			}
		}
		return PairWiseCoupling(n, r)
	}
}

func (m *SMO) ObtainsVotes(inst data.Instance, numClasses int) []int {

	// Filter instance
	if !m.checksTurnedOff {
		m.Missing.Input(inst)
		m.Missing.BatchFinished()
		inst = m.Missing.Output()
	}

	m.NominalToBinary.Input(inst)
	//No need to call batchFinished(), the queue never gonna be empty
	inst = m.NominalToBinary.Output()

	m.Filter.Input(inst)
	m.Filter.BatchFinished()
	inst = m.Filter.Output()

	votes := make([]int, numClasses)
	for i := range votes {
		for j := i + 1; j < numClasses; j++ {
			output := m.Classifiers[i][j].SVMOutput(-1, inst)
			if output > 0 {
				votes[j] += 1
			} else {
				votes[i] += 1
			}
		}
	}
	return votes
}

//From the MultiClassClassifier class
func PairWiseCoupling(n, r [][]float64) []float64 {
	// Initialize p and u array
	p := make([]float64, len(r))
	for i := range p {
		p[i] = 1.0 / float64(len(p))
	}
	u := make([][]float64, len(r))
	for i := range u {
		u[i] = make([]float64, len(r))
		for j := range u[i] {
			u[i][j] = 0.5
		}
	}

	//firstSum doesn't change
	firstSum := make([]float64, len(p))
	for i := range p {
		for j := i + 1; j < len(p); j++ {
			firstSum[i] += n[i][j] * r[i][j]
			firstSum[j] += n[i][j] * (1 - r[i][j])
		}
	}

	// Iterate until convergence
	changed := true
	for changed {
		changed = false
		secondSum := make([]float64, len(p))
		for i := range p {
			for j := i + 1; j < len(p); j++ {
				secondSum[i] += n[i][j] * u[i][j]
				secondSum[j] += n[i][j] * (1 - u[i][j])
			}
		}
		for i := range p {
			if (firstSum[i] == 0) || (secondSum[i] == 0) {
				if p[i] > 0 {
					changed = true
				}
				p[i] = 0
			} else {
				factor := firstSum[i] / secondSum[i]
				pOld := p[i]
				p[i] *= factor
				if math.Abs(pOld-p[i]) > 1.0e-3 {
					changed = true
				}
			}
		}
		sum := 0.0
		for _, d := range p {
			sum += d
		}
		utils.Normalize(&p, sum)
		for i := range r {
			for j := i + 1; j < len(r); j++ {
				u[i][j] = p[i] / (p[i] + p[j])
			}
		}
	}
	return p
}

// Type for building a binary support vector machine.
type BinarySMO struct {
	//The Lagrange multipliers
	Alpha []float64
	//The thresholds
	b, bLow, bUp float64
	//The indices for m_bLow and m_bUp
	iLow, iUp int
	//The training data
	data data.Instances
	//Weight vector for linear machine
	Weights []float64
	//Variables to hold weight vector in sparse form
	SparseWeights []float64
	SparseIndices []int
	//Kernel to use /*Always PolyKernel*/
	Kernel_ PolyKernel
	//The transformed class values
	class []float64
	//The current set of Errors for all non-bound examples
	Errors []float64
	//The five different sets used by the algorithm
	/** {i: 0 < Alpha[i] < C} */
	/**  {i: class[i] = 1, Alpha[i] = 0} */
	/**  {i: class[i] = -1, Alpha[i] =C} */
	/** {i: class[i] = 1, Alpha[i] = C} */
	/**  {i: class[i] = -1, Alpha[i] = 0} */
	i0, i1, i2, i3, i4 SMOSet
	//The set of support vectors
	// {i: 0 < Alpha[i]}
	supportVectors SMOSet
	//Stores the weight of the training instances
	sumOfWeights float64
	logistic     Logistic
	smo *SMO
	ClassIndex int
}

func NewBinarySMO() BinarySMO {
	return *new(BinarySMO)
}

//Call after NewBinarySMO(), always
func (bsmo *BinarySMO) SetSMO(smo *SMO) {
	bsmo.smo = smo
	bsmo.ClassIndex = smo.ClassIndex
}

//Fits logistic regression model to SVM outputs analogue
//     * to John Platt's method.
func (m *BinarySMO) fitLogistic(insts data.Instances, cl1, cl2, numFolds int, random *rand.Rand) {

	// Create header of instances object
	atts := make([]data.Attribute, 2)
	atts = append(atts, data.NewAttributeWithName("pred"))
	attVals := make([]string, 2)
	attVals = append(attVals, insts.ClassAttribute().Value(cl1))
	attVals = append(attVals, insts.ClassAttribute().Value(cl2))
	atts = append(atts, data.NewAttributeWithName("class"))
	_data := data.NewInstancesNameAttCap("data", atts, insts.NumInstances())
	_data.SetClassIndex(1)

	// Collect data for fitting the logistic model
	if numFolds <= 0 {

		// Use training data
		for _, inst := range insts.Instances() {
			vals := make([]float64, 2)
			vals[0] = m.SVMOutput(-1, inst)
			if inst.ClassValue(insts.ClassIndex()) == float64(cl2) {
				vals[1] = 1
			}
			_data.Add(data.NewInstanceWeightValues(inst.Weight(), vals))
		}
	} else {

		// Check whether number of folds too large
		if numFolds > insts.NumInstances() {
			numFolds = insts.NumInstances()
		}

		// Make copy of instances because we will shuffle them around
		insts = data.NewInstancesWithInst(insts, insts.NumInstances())

		// Perform three-fold cross-validation to collect
		// unbiased predictions
		insts.Randomizes(random)
		insts.Stratify(numFolds)
		for i := 0; i < numFolds; i++ {
			train := insts.TrainCVRand(numFolds, i, random)
			smo := NewBinarySMO()
			smo.SetKernel(m.Kernel_)
			smo.buildClassifier(train, cl1, cl2, false, -1, -1)
			test := insts.TestCV(numFolds, i)
			for j := 0; j < test.NumInstances(); j++ {
				vals := make([]float64, 2)
				vals[0] = m.SVMOutput(-1, test.InstanceNoPtr(j))
				if test.Instance(j).ClassValue(insts.ClassIndex()) == float64(cl2) { //if it doesn't work then use test.Instance(j).ClassValue(test.ClassIndex())
					vals[1] = 1
				}
				_data.Add(data.NewInstanceWeightValues(test.Instance(j).Weight(), vals))
			}
		}
	}

	// Build logistic regression modelF
	m.logistic = NewLogistic()
	m.logistic.BuildClassifier(_data)
}

// Method for building the binary classifier
func (bsmo *BinarySMO) buildClassifier(insts data.Instances, cl1, cl2 int, fitLogistic bool, numFolds int, randomSeed int) {
	//Initialize some variables
	bsmo.bUp, bsmo.bLow, bsmo.b = -1, 1, 0
	bsmo.Alpha, bsmo.Weights, bsmo.Errors = nil, nil, nil
	bsmo.SparseWeights, bsmo.SparseIndices = nil, nil

	//Store the sum of Weights
	bsmo.sumOfWeights = insts.SumOfWeights()
	//Set class values
	bsmo.class = make([]float64, insts.NumInstances())
	bsmo.iUp, bsmo.iLow = -1, -1
	for i := range bsmo.class {
		if int(insts.Instance(i).ClassValue(insts.ClassIndex())) == cl1 {
			bsmo.class[i], bsmo.iLow = -1, i
		} else if int(insts.Instance(i).ClassValue(insts.ClassIndex())) == cl2 {
			bsmo.class[i], bsmo.iUp = 1, i
		} else {
			panic("This should never happen.")
		}
	}

	//Check whether one or both classes are Missing
	if bsmo.iUp == -1 || bsmo.iLow == -1 {
		if bsmo.iUp != -1 {
			bsmo.b = -1
		} else if bsmo.iLow != -1 {
			bsmo.b = 1
		} else {
			bsmo.class = nil
			return
		}
		//Is linear always
		//if bsmo.kernelIsLinear {
		bsmo.SparseWeights = make([]float64, 0, 0)
		bsmo.SparseIndices = make([]int, 0, 0)
		bsmo.class = nil
		//		} else {
		//			bsmo.supportVectors = NewSMOSet(0)
		//			bsmo.Alpha = make([]float64, 0, 0)
		//			bsmo.class = make([]float64, 0, 0)
		//		}

		//Fit sigmoind if requested
		if fitLogistic {
			bsmo.fitLogistic(insts, cl1, cl2, numFolds, rand.New(rand.NewSource(int64(randomSeed))))
		}
		return
	}
	
	//Set the reference to the data
	bsmo.data = insts

	//If machine is linear, reserve space for Weights
	//Is linear
	//if bsmo.kernelIsLinear {
	bsmo.Weights = make([]float64, bsmo.data.NumAttributes())
	//	} else {
	//		bsmo.Weights = nil
	//	}

	//Initialize Alpha array to zero
	bsmo.Alpha = make([]float64, bsmo.data.NumInstances())
	
	//Initialize sets
	bsmo.supportVectors = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i0 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i1 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i2 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i3 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i4 = NewSMOSet(bsmo.data.NumInstances())

	//Clean out some instances variables
	bsmo.SparseWeights = nil
	bsmo.SparseIndices = nil

	//Init Kernel_
	bsmo.Kernel_.BuildKernel(bsmo.data)

	//Initialize error cache
	bsmo.Errors = make([]float64, bsmo.data.NumInstances())
	bsmo.Errors[bsmo.iLow], bsmo.Errors[bsmo.iUp] = 1, -1

	//Build up i1 and i4
	for i, class := range bsmo.class {
		if class == 1 {
			bsmo.i1.Insert(i)
		} else {
			bsmo.i4.Insert(i)
		}
	}

	//Loop to find all the support vectors
	numChanged := 0
	examineAll := true
	for numChanged > 0 || examineAll {
		numChanged = 0
		if examineAll {
			for i := range bsmo.Alpha {
				if bsmo.ExamineExample(i) {
					numChanged++
				}
			}
		} else {

			//This code implements Modification 1 from Keerthi et al.'s paper
			for i, Alpha := range bsmo.Alpha {
				if Alpha > 0 && Alpha < bsmo.smo.C_C*bsmo.data.Instance(i).Weight() {
					if bsmo.ExamineExample(i) {
						numChanged++
					}

					//Is optimality on unbound vectors obtained?
					if bsmo.bUp > bsmo.bLow-2*bsmo.smo.Tol {
						numChanged = 0
						break
					}
				}
			}

		}

		if examineAll {
			examineAll = false
		} else if numChanged == 0 {
			examineAll = true
		}
	}

	//Set threshold
	bsmo.b = (bsmo.bLow + bsmo.bUp) / 2.0

	//Save memory
	bsmo.Kernel_.Clean()

	bsmo.Errors = nil
	bsmo.i0, bsmo.i1, bsmo.i2, bsmo.i3, bsmo.i4 = *new(SMOSet) , *new(SMOSet), *new(SMOSet), *new(SMOSet), *new(SMOSet)

	//If machine is linear, delete training data and store weight vector in sparse format
	//Is Linear
	//if bsmo.kernelIsLinear {
	//Don't need to store the set of support vectors
	bsmo.supportVectors =  *new(SMOSet)

	//Don't need to store the class values either
	bsmo.class = nil

	//Clean out training data
	if !bsmo.smo.checksTurnedOff {
		bsmo.data = data.NewInstancesWithInst(bsmo.data, 0)
	} else {
		//bsmo.data =nill
	}

	//Convert weight vectors
	SparseWeights := make([]float64, len(bsmo.Weights))
	SparseIndices := make([]int, len(bsmo.Weights))
	counter := 0
	for i, weight := range bsmo.Weights {
		if weight != 0 {
			SparseWeights[counter] = weight
			SparseIndices[counter] = i
			counter++
		}
	}
	bsmo.SparseWeights = make([]float64, counter)
	bsmo.SparseIndices = make([]int, counter)
	copy(bsmo.SparseWeights, SparseWeights[:counter])
	copy(bsmo.SparseIndices, SparseIndices[:counter])
	//Clean out Weights vector
	bsmo.Weights = nil
	//Don't need the alphas in the linear case
	bsmo.Alpha = nil
	//}
	//ok so far
	//Fit sigmoid if requested
	if fitLogistic {
		bsmo.fitLogistic(insts, cl1, cl2, numFolds, rand.New(rand.NewSource(int64(randomSeed))))
	}
}

// Examines instance
func (this *BinarySMO) ExamineExample(i2 int) bool {
	var y2, F2 float64
	i1 := -1
	
	y2 = this.class[i2]
	if this.i0.Contains(i2) {
		F2 = this.Errors[i2]
	} else {
		F2 = this.SVMOutput(i2, this.data.InstanceNoPtr(i2)) + this.b - y2
		this.Errors[i2] = F2

		//Update thresholds
		if this.i1.Contains(i2) || this.i2.Contains(i2) && F2 < this.bUp {
			this.bUp, this.iUp = F2, i2
		} else if this.i3.Contains(i2) || this.i4.Contains(i2) && F2 > this.bLow {
			this.bLow, this.iLow = F2, i2
		}
	}
	// Check optimality using current bLow and bUp and, if
	// violated, find an index i1 to do joint optimization
	// with i2...
	optimal := true
	if this.i0.Contains(i2) || this.i1.Contains(i2) || this.i2.Contains(i2) {
		if this.bLow-F2 > 2*this.smo.Tol {
			optimal, i1 = false, this.iLow
		}
	}
	if this.i0.Contains(i2) || this.i3.Contains(i2) || this.i4.Contains(i2) {
		if F2-this.bUp > 2*this.smo.Tol {
			optimal, i1 = false, this.iUp
		}
	}
	if optimal {
		return false
	}
	//For i2 unbound choose the better i1...
	if this.i0.Contains(i2) {
		if this.bLow-F2 > F2-this.bUp {
			i1 = this.iLow
		} else {
			i1 = this.iUp
		}
	}
	if i1 == -1 {
		panic("This should never happen.")
	}

	return this.takeStep(i1, i2, F2)
}

// Method solving for the Lagrange multipliers for two instances
func (this *BinarySMO) takeStep(i1, i2 int, F2 float64) bool {
	//fmt.Println(i1,i2,"i1 i2")
	var alph1, alph2, y1, y2, F1, s, L, H, k11, k12, k22, eta, a1, a2, f1, f2, v1, v2, Lobj, Hobj float64
	C1 := this.smo.C_C * this.data.Instance(i1).Weight()
	C2 := this.smo.C_C * this.data.Instance(i2).Weight()

	//Don't do anything if the two instances are the same
	if i1 == i2 {
		return false
	}

	//Initialize variables
	alph1, alph2 = this.Alpha[i1], this.Alpha[i2]
	y1, y2 = this.class[i1], this.class[i2]
	F1 = this.Errors[i1]
	s = y1 * y2
//	ok so far
	//Find the constraints on a2
	if y1 != y2 {
		L = math.Max(0, alph2-alph1)
		H = math.Min(C2, C1+alph2-alph1)
	} else {
		L = math.Max(0, alph1+alph2-C1)
		H = math.Min(C2, alph1+alph2)
	}
	
	if L >= H {
		return false
	}

	//Compute second derivative of the objective function
	k11 = this.Kernel_.Eval(i1, i1, this.data.InstanceNoPtr(i1))
	k12 = this.Kernel_.Eval(i1, i2, this.data.InstanceNoPtr(i1))
	k22 = this.Kernel_.Eval(i2, i2, this.data.InstanceNoPtr(i2))
	eta = 2*k12 - k11 - k22

	//Check if second derivative is negative
	if eta < 0 {

		//Compute unconstrained maximun
		a2 = alph2 - y2*(F1-F2)/eta

		//Compute constrained maximun
		if a2 < L {
			a2 = L
		} else if a2 > H {
			a2 = H
		}
	} else {

		//Look at endpoints of diagonal
		f1 = this.SVMOutput(i1, this.data.InstanceNoPtr(i1))
		f2 = this.SVMOutput(i2, this.data.InstanceNoPtr(i2))
		v1 = f1 + this.b - y1*alph1*k11 - y2*alph2*k12
		v2 = f2 + this.b - y1*alph1*k12 - y2*alph2*k22
		gamma := alph1 + s*alph2
		Lobj = (gamma - s*L) + L - 0.5*k11*(gamma-s*L)*(gamma-s*L) -
			0.5*k22*L*L - s*k12*(gamma-s*L)*L -
			y1*(gamma-s*L)*v1 - y2*L*v2
		Hobj = (gamma - s*H) + H - 0.5*k11*(gamma-s*H)*(gamma-s*H) -
			0.5*k22*H*H - s*k12*(gamma-s*H)*H -
			y1*(gamma-s*H)*v1 - y2*H*v2
		if Lobj > Hobj+this.smo.Eps {
			a2 = L
		} else if Lobj < Hobj-this.smo.Eps {
			a2 = H
		} else {
			a2 = alph2
		}
	}
	//ok so far
	if math.Abs(a2-alph2) < this.smo.Eps*(a2+alph2+this.smo.Eps) {
		return false
	}

	//To prevent precision problems
	if a2 > C2-this.smo.Del*C2 {
		a2 = C2
	} else if a2 <= this.smo.Del*C2 {
		a2 = 0
	}

	//Recompute a1
	a1 = alph1 + s*(alph2-a2)
	//fmt.Println(a1,"a1")

	//To prevent precision problems
	if a1 > C1-this.smo.Del*C1 {
		a1 = C1
	} else if a1 <= this.smo.Del*C1 {
		a1 = 0
	}

	//Update sets
	if a1 > 0 {
		this.supportVectors.Insert(i1)
	} else {
		this.supportVectors.Delete(i1)
	}
	if a1 > 0 && a1 < C1 {
		this.i0.Insert(i1)
	} else {
		this.i0.Delete(i1)
	}
	if (y1 == 1) && (a1 == 0) {
		this.i1.Insert(i1)
	} else {
		this.i1.Delete(i1)
	}
	if (y1 == -1) && (a1 == C1) {
		this.i2.Insert(i1)
	} else {
		this.i2.Delete(i1)
	}
	if (y1 == 1) && (a1 == C1) {
		this.i3.Insert(i1)
	} else {
		this.i3.Delete(i1)
	}
	if (y1 == -1) && (a1 == 0) {
		this.i4.Insert(i1)
	} else {
		this.i4.Delete(i1)
	}
	if a2 > 0 {
		this.supportVectors.Insert(i2)
	} else {
		this.supportVectors.Delete(i2)
	}
	if (a2 > 0) && (a2 < C2) {
		this.i0.Insert(i2)
	} else {
		this.i0.Delete(i2)
	}
	if (y2 == 1) && (a2 == 0) {
		this.i1.Insert(i2)
	} else {
		this.i1.Delete(i2)
	}
	if (y2 == -1) && (a2 == C2) {
		this.i2.Insert(i2)
	} else {
		this.i2.Delete(i2)
	}
	if (y2 == 1) && (a2 == C2) {
		this.i3.Insert(i2)
	} else {
		this.i3.Delete(i2)
	}
	if (y2 == -1) && (a2 == 0) {
		this.i4.Insert(i2)
	} else {
		this.i4.Delete(i2)
	}

	//Update weight vector to reflect change a1 and a2, if linear SVM
	//the Kernel_ is linear
	//if this.kernelIsLinear {
	inst1 := this.data.Instance(i1)
	for p1 := range inst1.RealValues() {
		if inst1.Index(p1) != this.data.ClassIndex() {
			this.Weights[inst1.Index(p1)] += y1 * (a1 - alph1) * inst1.ValueSparse(p1)
		}
	}
	inst2 := this.data.Instance(i2)
	for p2 := range inst2.RealValues() {
		if inst2.Index(p2) != this.data.ClassIndex() {
			this.Weights[inst2.Index(p2)] += y2 * (a2 - alph2) * inst2.ValueSparse(p2)
		}
	}
	//}

	//Update error cache using new Langrage multipliers
	for j := this.i0.GetNext(-1); j != -1; j = this.i0.GetNext(j) {
		if j != i1 && j != i2 {
			this.Errors[j] +=
				y1*(a1-alph1)*this.Kernel_.Eval(i1, j, this.data.InstanceNoPtr(i1)) +
					y2*(a2-alph2)*this.Kernel_.Eval(i2, j, this.data.InstanceNoPtr(i2))
		}
	}

	//Update error cache for i1 and i2
	this.Errors[i1] += y1*(a1-alph1)*k11 + y2*(a2-alph2)*k12
	this.Errors[i2] += y1*(a1-alph1)*k12 + y2*(a2-alph2)*k22

//ok so far

	//Update array with Langrage multipliers
	this.Alpha[i1] = a1
	this.Alpha[i2] = a2

	//Update thresholds
	this.bLow, this.bUp = -math.MaxFloat64, math.MaxFloat64
	this.iLow, this.iUp = -1, -1
	for j := this.i0.GetNext(-1); j != -1; j = this.i0.GetNext(j) {
		if this.Errors[j] < this.bUp {
			this.bUp, this.iUp = this.Errors[j], j
		}
		if this.Errors[j] > this.bLow {
			this.bLow, this.iLow = this.Errors[j], j
		}
	}
	if !this.i0.Contains(i1) {
		if this.i3.Contains(i1) || this.i4.Contains(i1) {
			if this.Errors[i1] > this.bLow {
				this.bLow, this.iLow = this.Errors[i1], i1
			}
		} else {
			if this.Errors[i1] < this.bUp {
				this.bUp, this.iUp = this.Errors[i1], i1
			}
		}
	}
	if !this.i0.Contains(i2) {
		if this.i3.Contains(i2) || this.i4.Contains(i2) {
			if this.Errors[i2] > this.bLow {
				this.bLow, this.iLow = this.Errors[i2], i2
			}
		} else {
			if this.Errors[i2] < this.bUp {
				this.bUp, this.iUp = this.Errors[i2], i2
			}
		}
	}
	if this.iLow == -1 || this.iUp == -1 {
		panic("This should never happen.")
	}
//ok so far
	//Made some progress
	return true
}

func (this *BinarySMO) SVMOutput(index int, inst data.Instance) float64 {
	
	result := 0.0
	//Is this a linear machine? Is always linear
	//if this.kernelIsLinear {

	//Is weight vector stored in sparse format

	if this.SparseWeights == nil {
		for p := range inst.RealValues() {
			if inst.Index(p) != this.smo.ClassIndex {
				result += this.Weights[inst.Index(p)] * inst.ValueSparse(p)
			}
		}
	} else {
		n1, n2 := len(inst.RealValues()), len(this.SparseWeights)
		for p1, p2 := 0, 0; p1 < n1 && p2 < n2; {
			ind1, ind2 := inst.Index(p1), this.SparseIndices[p2]
			if ind1 == ind2 {
				if ind1 != this.ClassIndex {
					result += inst.ValueSparse(p1) * this.SparseWeights[p2]
				}
				p1++
				p2++
			} else if ind1 > ind2 {
				p2++
			} else {
				p1++
			}
		}
	}
	//	} else {
	//		for i := this.supportVectors.GetNext(-1); i != -1; i = this.supportVectors.GetNext(i) {
	//			result += this.class[i] * this.Alpha[i] * this.Kernel_.Eval(index, i, inst)
	//		}
	//	}
	result -= this.b
	return result
}

// Quick and dirty check whether the quadratic programming problem is solved.
func (bsmo *BinarySMO) CheckClassifier() {
	sum := 0.0
	for i, alp := range bsmo.Alpha {
		if alp > 0 {
			sum += bsmo.class[i] / alp
		}
	}
	println("Sum of y(i) * Alpha(i): ", sum)

	for i := 0; i < len(bsmo.Alpha); i++ {
		output := bsmo.SVMOutput(i, *bsmo.data.Instance(i))
		if utils.Eq(bsmo.Alpha[i], 0) {
			if utils.Sm(bsmo.class[i]*output, 1) {
				println("KKT condition 1 violated: ", bsmo.class[i]*output)
			}
		}
		if utils.Gr(bsmo.Alpha[i], 0) && utils.Sm(bsmo.Alpha[i], bsmo.smo.C_C*bsmo.data.Instance(i).Weight()) {
			if !utils.Eq(bsmo.class[i]*output, 1) {
				println("KKT condition 2 violated: ", bsmo.class[i]*output)
			}
		}
		if utils.Eq(bsmo.Alpha[i], bsmo.smo.C_C*bsmo.data.Instance(i).Weight()) {
			if utils.Gr(bsmo.class[i]*output, 1) {
				println("KKT condition 3 violated: ", bsmo.class[i]*output)
			}
		}
	}

}

func (bsmo *BinarySMO) SetKernel(value PolyKernel) {
	bsmo.Kernel_ = value
}

func (bsmo *BinarySMO) Kernel() PolyKernel {
	return bsmo.Kernel_

}
