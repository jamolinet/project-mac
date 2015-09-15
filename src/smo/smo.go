package smo

import (
	"fmt"
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/functions"
	"github.com/project-mac/src/utils"
	"math"
	"math/rand"
)

const (
	//filter: Normalize training data
	FILTER_NORMALIZE = 0
	//filter: Standardize training data
	FILTER_STANDARDIZE = 1
	//filter: No normalization/standardization
	FILTER_NONE = 2
)

// Implements John Platt's sequential minimal optimization algorithm for training a support vector classifier.
type SMO struct {
	//The binary classifier(s)
	classifiers [][]BinarySMO

	//The complexity parameter
	c_C float64

	//Epsilon for rounding
	eps float64

	//Tolerance for accuracy of result
	tol float64

	//Whether to normalize/standardize/neither
	filterType int

	/** Remains other attributes to declare*/
	/**********/ /////
	//The filter used to make attributes numeric
	nominalToBinary functions.NominalToBinary

	//The filter used to standardize/normalize all values
	filterS functions.Standardize
	filterN functions.Normalize
	filter  functions.Filter

	//The filter used to get rid of missing values
	missing functions.ReplaceMissingValues

	//The class index from the training data
	classIndex int

	//The class attribute
	classAttribute data.Attribute

	//whether the kernel is a linear one
	kernelIsLinear bool

	//	Turn off all checks and conversions? Turning them off assumes
	//      that data is purely numeric, doesn't contain any missing values,
	//      and has a nominal class. Turning them off also means that
	//      no header information will be stored if the machine is linear.
	//      Finally, it also assumes that no instance has a weight equal to 0
	checksTurnedOff bool

	//Precision constant for updating sets
	del float64

	//Whether logistic models are to be fit
	fitLogisticModels bool

	//The number of folds for the internal cross-validation
	numFolds int

	//The random number seed
	randomSeed int

	//the kernel to use
	kernel PolyKernel
}

func NewSMO(kernel PolyKernel) SMO {
	var smo SMO
	smo.classifiers = nil
	smo.c_C = 1.0
	smo.eps = 1.0e-12
	smo.tol = 1.0e-3
	smo.filterType = FILTER_NORMALIZE
	smo.classIndex = -1
	smo.kernelIsLinear = true
	smo.del = -math.MaxFloat64
	smo.fitLogisticModels = false
	smo.numFolds = -1
	smo.randomSeed = 1
	smo.kernel = kernel
	return smo
}

func (m *SMO) SetEps(eps float64) {
	m.eps = eps
}

func (m *SMO) SetC(c float64) {
	m.c_C = c
}

func (m *SMO) SetTolerance(tol float64) {
	m.tol = tol
}

func (m *SMO) SetSeed(seed int) {
	m.randomSeed = seed
}

func (m *SMO) SetNumFolds(folds int) {
	m.numFolds = folds
}

func (m *SMO) SetFitLogistic(l bool) {
	m.fitLogisticModels = l
}

func (m *SMO) SetNormalize(n bool) {
	if n {
		m.filterType = FILTER_NORMALIZE
	} else {
		m.filterType = FILTER_STANDARDIZE
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
		//Remove instances with missing value

		insts.DeleteWithMissingClass()
		//fmt.Println(len(insts.Instances()), "missing class")

		/* Removes all the instances with weight equal to 0.
		   MUST be done since condition (8) of Keerthi's paper
		   is made with the assertion Ci > 0 (See equation (3a). */
		data := data.NewInstancesWithInst(insts, insts.NumInstances())
		//fmt.Println(len(data.Instances()), "missing class")
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
		m.missing = functions.NewReplacingMissingValues()
		m.missing.SetInputFormat(insts)
		m.missing.Exec(insts)
		m.missing.BatchFinished()
		insts = m.missing.OutputAll()
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
		m.nominalToBinary = functions.NewNominalToBinary()
		m.nominalToBinary.SetInputFormat(insts)
		m.nominalToBinary.Exec(insts)
		insts = m.nominalToBinary.OutputAll()
	} else {
		m.nominalToBinary.IsNil = ""
	}

	//fmt.Println(insts.NumInstances(), "len")

	if m.filterType == FILTER_STANDARDIZE {
		m.filter = functions.NewStandardizePtr()
		m.filter.SetInputFormat(insts)
		m.filter.Exec(insts)
		insts = m.filter.OutputAll()
	} else if m.filterType == FILTER_NORMALIZE {
		//println("normalizing")
		m.filter = functions.NewNormalizePtr()
		m.filter.SetInputFormat(insts)
		m.filter.Exec(insts)
		insts = m.filter.OutputAll()
	}
	//	for _, in := range insts.Instances() {
	//		fmt.Println(in.RealValues(), in.NumAttributesTest(), "--1")
	//		fmt.Println(in.Indices(), "--2")
	//	}
	m.classIndex = insts.ClassIndex()
	m.classAttribute = insts.ClassAttribute()
	//This kernel will always be linear in this case
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
	// Build the binary classifiers
	m.classifiers = make([][]BinarySMO, insts.NumClasses())
	for i := range m.classifiers {
		m.classifiers[i] = make([]BinarySMO, insts.NumClasses())
	}

	for i := 0; i < insts.NumClasses(); i++ {
		for j := i + 1; j < insts.NumClasses(); j++ {
			m.classifiers[i][j] = NewBinarySMO()
			m.classifiers[i][j].SetSMO(m)
			m.classifiers[i][j].SetKernel(m.kernel)
			_data := data.NewInstancesWithInst(insts, insts.NumInstances())
			for k := 0; k < subsets[i].NumInstances(); k++ {
				_data.Add(subsets[i].InstanceNoPtr(k))
			}
			for k := 0; k < subsets[j].NumInstances(); k++ {
				_data.Add(subsets[j].InstanceNoPtr(k))
			}
			_data.Randomize(m.randomSeed)
			//ok so far	
			m.classifiers[i][j].buildClassifier(_data, i, j, m.fitLogisticModels, m.numFolds, m.randomSeed)
		}
	}
	//fmt.Println("new fold")
}

// Estimates class probabilities for given instance.
//The numClasses parameter is mandatory
func (m *SMO) DistributionForInstance(inst data.Instance, numClasses int) []float64 {
	/* TODO: See how to check if a type is nil*/
//		fmt.Println(inst.Indices())	
//		fmt.Println(inst.RealValues(),"rereereee", inst.NumAttributes())
	//Filter instance
	if !m.checksTurnedOff {
		m.missing.Input(inst)
		//fmt.Println(m.missing.ModesAndMeans())
		m.missing.BatchFinished()
		inst = m.missing.Output()
	}
		
//		fmt.Println(inst.Indices())	
//		fmt.Println(inst.RealValues(),"ggggggggggg", inst.NumAttributes())
	if m.nominalToBinary.IsNil == "no" {
		fmt.Println("m.nominalToBinary.IsNil ==")
		m.nominalToBinary.Input(inst)
		//No need to call batchFinished(), the queue never gonna be empty
		inst = m.nominalToBinary.Output()
	}

	if m.filter.NotNil() {
		m.filter.Input(inst)
		m.filter.BatchFinished()
		inst = m.filter.Output()
	}
	
//	fmt.Println(inst.Indices())	
//		fmt.Println(inst.RealValues(),"nononononono", inst.NumAttributes())
	
	//fmt.Println(numClasses, "numclasses")
	if !m.fitLogisticModels {
		result := make([]float64, numClasses)
		for i := 0; i < numClasses; i++ {
			for j := i + 1; j < numClasses; j++ {
				if m.classifiers[i][j].alpha != nil || m.classifiers[i][j].sparseWeights != nil {
					//println("inininin")
					//fmt.Println(i,j)
					output := m.classifiers[i][j].SVMOutput(-1, inst)
					//fmt.Println(output, "output 09090")
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
//		fmt.Println(result, "distribution before normalize",sum)
		utils.Normalize(&result, sum)
		fmt.Println(result, "distribution after normalize",sum)
		return result
	} else {
		// We only need to do pairwise coupling if there are more
		// then two classes.
		if numClasses == 2 {
			newInst := make([]float64, 2)
			newInst[0] = m.classifiers[0][1].SVMOutput(-1, inst)
			newInst[1] = inst.MissingValue
			return m.classifiers[0][1].logistic.DistributionForInstance(data.NewInstanceWeightValues(1, newInst))
		}
		r, n := make([][]float64, numClasses), make([][]float64, numClasses)
		for i := range r {
			n[i] = make([]float64, numClasses)
			r[i] = make([]float64, numClasses)
		}
		for i := 0; i < numClasses; i++ {
			for j := i + 1; j < numClasses; j++ {
				if m.classifiers[i][j].alpha != nil || m.classifiers[i][j].sparseWeights != nil {
					newInst := make([]float64, 2)
					newInst[0] = m.classifiers[0][1].SVMOutput(-1, inst)
					newInst[1] = inst.MissingValue
					r[i][j] = m.classifiers[i][j].logistic.DistributionForInstance(data.NewInstanceWeightValues(1, newInst))[0]
					n[i][j] = m.classifiers[i][j].sumOfWeights
				}
			}
		}
		return PairWiseCoupling(n, r)
	}
}

func (m *SMO) ObtainsVotes(inst data.Instance, numClasses int) []int {

	// Filter instance
	if !m.checksTurnedOff {
		m.missing.Input(inst)
		m.missing.BatchFinished()
		inst = m.missing.Output()
	}

	m.nominalToBinary.Input(inst)
	//No need to call batchFinished(), the queue never gonna be empty
	inst = m.nominalToBinary.Output()

	m.filter.Input(inst)
	m.filter.BatchFinished()
	inst = m.filter.Output()

	votes := make([]int, numClasses)
	for i := range votes {
		for j := i + 1; j < numClasses; j++ {
			output := m.classifiers[i][j].SVMOutput(-1, inst)
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
	alpha []float64
	//The thresholds
	b, bLow, bUp float64
	//The indices for m_bLow and m_bUp
	iLow, iUp int
	//The training data
	data data.Instances
	//Weight vector for linear machine
	weights []float64
	//Variables to hold weight vector in sparse form
	sparseWeights []float64
	sparseIndices []int
	//Kernel to use /*Always PolyKernel*/
	kernel PolyKernel
	//The transformed class values
	class []float64
	//The current set of errors for all non-bound examples
	errors []float64
	//The five different sets used by the algorithm
	/** {i: 0 < alpha[i] < C} */
	/**  {i: class[i] = 1, alpha[i] = 0} */
	/**  {i: class[i] = -1, alpha[i] =C} */
	/** {i: class[i] = 1, alpha[i] = C} */
	/**  {i: class[i] = -1, alpha[i] = 0} */
	i0, i1, i2, i3, i4 SMOSet
	//The set of support vectors
	// {i: 0 < alpha[i]}
	supportVectors SMOSet
	//Stores the weight of the training instances
	sumOfWeights float64
	logistic     Logistic
	*SMO
}

func NewBinarySMO() BinarySMO {
	return *new(BinarySMO)
}

//Call after NewBinarySMO(), always
func (bsmo *BinarySMO) SetSMO(smo *SMO) {
	bsmo.SMO = smo
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
			smo.SetKernel(m.kernel)
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
	//fmt.Println("building kernel binary")
	//Initialize some variables
	bsmo.bUp, bsmo.bLow, bsmo.b = -1, 1, 0
	bsmo.alpha, bsmo.weights, bsmo.errors = nil, nil, nil
	bsmo.sparseWeights, bsmo.sparseIndices = nil, nil

	//Store the sum of weights
	bsmo.sumOfWeights = insts.SumOfWeights()
	//fmt.Println(bsmo.sumOfWeights,"bsmo.sumOfWeights", cl1, cl2)
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
	//fmt.Println("class",bsmo.class, bsmo.iLow, bsmo.iUp)

	//Check whether one or both classes are missing
	if bsmo.iUp == -1 || bsmo.iLow == -1 {
		//fmt.Println("inside bsmo.iUp")
		if bsmo.iUp != -1 {
			bsmo.b = -1
		} else if bsmo.iLow != -1 {
			bsmo.b = 1
		} else {
			bsmo.class = nil
			//fmt.Println("inside return")
			return
		}
		//Is linear always
		//if bsmo.kernelIsLinear {
		bsmo.sparseWeights = make([]float64, 0, 0)
		bsmo.sparseIndices = make([]int, 0, 0)
		bsmo.class = nil
		//		} else {
		//			bsmo.supportVectors = NewSMOSet(0)
		//			bsmo.alpha = make([]float64, 0, 0)
		//			bsmo.class = make([]float64, 0, 0)
		//		}

		//Fit sigmoind if requested
		if fitLogistic {
			bsmo.fitLogistic(insts, cl1, cl2, numFolds, rand.New(rand.NewSource(int64(randomSeed))))
		}
		//fmt.Println("inside", cl1,cl2)
		//fmt.Println(len(bsmo.sparseWeights), len(bsmo.sparseIndices), bsmo.class == nil, bsmo.b,bsmo.iLow, bsmo.iUp)
		return
	}
	
	//fmt.Println("outside", cl1, cl2, len(insts.Instances()), len(insts.Attributes())) // ok so far
	//fmt.Println(cl1,cl2)
	//Set the reference to the data
	bsmo.data = insts

	//If machine is linear, reserve space for weights
	//Is linear
	//if bsmo.kernelIsLinear {
	//fmt.Println(bsmo.data.NumAttributes(),"bsmo.data.NumAttributes()")
	bsmo.weights = make([]float64, bsmo.data.NumAttributes())
	//	} else {
	//		bsmo.weights = nil
	//	}

	//Initialize alpha array to zero
	bsmo.alpha = make([]float64, bsmo.data.NumInstances())
	
	//Initialize sets
	bsmo.supportVectors = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i0 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i1 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i2 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i3 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i4 = NewSMOSet(bsmo.data.NumInstances())

	//fmt.Println(len(bsmo.weights),len(bsmo.alpha), len(bsmo.supportVectors.indicators), len(bsmo.i4.indicators), len(bsmo.class))
	
	//CLean out some instances variables
	bsmo.sparseWeights = nil
	bsmo.sparseIndices = nil

	//Init kernel
	bsmo.kernel.BuildKernel(bsmo.data)

	//Initialize error cache
	bsmo.errors = make([]float64, bsmo.data.NumInstances())
	bsmo.errors[bsmo.iLow], bsmo.errors[bsmo.iUp] = 1, -1
	//fmt.Println(bsmo.class, "class")
	//Build up i1 and i4
	for i, class := range bsmo.class {
		if class == 1 {
			bsmo.i1.Insert(i)
		} else {
			bsmo.i4.Insert(i)
		}
	}
	//fmt.Println(bsmo.alpha, "alpha")
	//Loop to find all the support vectors
	numChanged := 0
	examineAll := true
	for numChanged > 0 || examineAll {
		numChanged = 0
		if examineAll {
			for i := range bsmo.alpha {
				if bsmo.ExamineExample(i) {
					numChanged++
				}
			}
		} else {

			//Tis code implements Modification 1 from Keerthi et al.'s paper
			for i, alpha := range bsmo.alpha {
				if alpha > 0 && alpha < bsmo.c_C*bsmo.data.Instance(i).Weight() {
					if bsmo.ExamineExample(i) {
						numChanged++
					}

					//Is optimality on unbound vectors obtained?
					if bsmo.bUp > bsmo.bLow-2*bsmo.tol {
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
	//fmt.Println("executed", bsmo.b)

	//Save memory
	bsmo.kernel.Clean()

	bsmo.errors = nil
	bsmo.i0, bsmo.i1, bsmo.i2, bsmo.i3, bsmo.i4 = *new(SMOSet) , *new(SMOSet), *new(SMOSet), *new(SMOSet), *new(SMOSet)

	//If machine is linear, delete training data and store weight vector in sparse format
	//Is Linear
	//if bsmo.kernelIsLinear {
	//Don't need to store the set of support vectors
	bsmo.supportVectors =  *new(SMOSet)

	//Don't need to store the class values either
	bsmo.class = nil

	//Clean out training data
	if !bsmo.checksTurnedOff {
		bsmo.data = data.NewInstancesWithInst(bsmo.data, 0)
	} else {
		//bsmo.data =nill
	}

	//Convert weight vectors
	sparseWeights := make([]float64, len(bsmo.weights))
	sparseIndices := make([]int, len(bsmo.weights))
	counter := 0
	for i, weight := range bsmo.weights {
		if weight != 0 {
			sparseWeights[counter] = weight
			sparseIndices[counter] = i
			counter++
		}
	}
	//fmt.Println(bsmo.weights)
	bsmo.sparseWeights = make([]float64, counter)
	bsmo.sparseIndices = make([]int, counter)
	copy(bsmo.sparseWeights, sparseWeights[:counter])
	copy(bsmo.sparseIndices, sparseIndices[:counter])
	//Clean out weights vector
	bsmo.weights = nil
	//Don't need the alphas in the linear case
	bsmo.alpha = nil
	//}
	//fmt.Println(bsmo.sparseWeights, bsmo.sparseIndices, "indexex, weights") ok so far
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
		F2 = this.errors[i2]
		//fmt.Println("see test", i2)
	} else {
		F2 = this.SVMOutput(i2, this.data.InstanceNoPtr(i2)) + this.b - y2
		this.errors[i2] = F2

		//Update thresholds
		if this.i1.Contains(i2) || this.i2.Contains(i2) && F2 < this.bUp {
			this.bUp, this.iUp = F2, i2
		} else if this.i3.Contains(i2) || this.i4.Contains(i2) && F2 > this.bLow {
			this.bLow, this.iLow = F2, i2
		}
	}
	//fmt.Println(F2,"F2----------")
	// Check optimality using current bLow and bUp and, if
	// violated, find an index i1 to do joint optimization
	// with i2...
	optimal := true
	if this.i0.Contains(i2) || this.i1.Contains(i2) || this.i2.Contains(i2) {
		if this.bLow-F2 > 2*this.tol {
			optimal, i1 = false, this.iLow
		}
	}
	if this.i0.Contains(i2) || this.i3.Contains(i2) || this.i4.Contains(i2) {
		if F2-this.bUp > 2*this.tol {
			optimal, i1 = false, this.iUp
		}
	}
	if optimal {
		return false
	}
	//fmt.Println("first i2", i2)
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
	C1 := this.c_C * this.data.Instance(i1).Weight()
	C2 := this.c_C * this.data.Instance(i2).Weight()

	//Don't do anything if the two instances are the same
	if i1 == i2 {
		return false
	}

	//Initialize variables
	alph1, alph2 = this.alpha[i1], this.alpha[i2]
	y1, y2 = this.class[i1], this.class[i2]
	F1 = this.errors[i1]
	s = y1 * y2
//	fmt.Println(alph1, alph2, y1,y2,F1,s, "hghg") ok so far
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
	k11 = this.kernel.Eval(i1, i1, this.data.InstanceNoPtr(i1))
	k12 = this.kernel.Eval(i1, i2, this.data.InstanceNoPtr(i1))
	k22 = this.kernel.Eval(i2, i2, this.data.InstanceNoPtr(i2))
	eta = 2*k12 - k11 - k22
	//fmt.Println(k11,k12,k22,eta, "fhdjhkf")
	//fmt.Println(eta,"eta")
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
		if Lobj > Hobj+this.eps {
			a2 = L
		} else if Lobj < Hobj-this.eps {
			a2 = H
		} else {
			a2 = alph2
		}
	}
	//ok so far
	if math.Abs(a2-alph2) < this.eps*(a2+alph2+this.eps) {
		return false
	}

	//To prevent precision problems
	if a2 > C2-this.del*C2 {
		a2 = C2
	} else if a2 <= this.del*C2 {
		a2 = 0
	}

	//Recompute a1
	a1 = alph1 + s*(alph2-a2)
	//fmt.Println(a1,"a1")

	//To prevent precision problems
	if a1 > C1-this.del*C1 {
		a1 = C1
	} else if a1 <= this.del*C1 {
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
	//the kernel is linear
	//if this.kernelIsLinear {
	inst1 := this.data.Instance(i1)
	for p1 := range inst1.RealValues() {
		if inst1.Index(p1) != this.data.ClassIndex() {
			this.weights[inst1.Index(p1)] += y1 * (a1 - alph1) * inst1.ValueSparse(p1)
		}
	}
	inst2 := this.data.Instance(i2)
	for p2 := range inst2.RealValues() {
		if inst2.Index(p2) != this.data.ClassIndex() {
			this.weights[inst2.Index(p2)] += y2 * (a2 - alph2) * inst2.ValueSparse(p2)
		}
	}
	//}

	//Update error cache using new Langrage multipliers
	for j := this.i0.GetNext(-1); j != -1; j = this.i0.GetNext(j) {
		if j != i1 && j != i2 {
			this.errors[j] +=
				y1*(a1-alph1)*this.kernel.Eval(i1, j, this.data.InstanceNoPtr(i1)) +
					y2*(a2-alph2)*this.kernel.Eval(i2, j, this.data.InstanceNoPtr(i2))
		}
	}

	//Update error cache for i1 and i2
	this.errors[i1] += y1*(a1-alph1)*k11 + y2*(a2-alph2)*k12
	this.errors[i2] += y1*(a1-alph1)*k12 + y2*(a2-alph2)*k22

//	fmt.Println(this.weights, "weights")
//	fmt.Println(this.errors, "errors")
//ok so far

	//Update array with Langrage multipliers
	this.alpha[i1] = a1
	this.alpha[i2] = a2

	//Update thresholds
	this.bLow, this.bUp = -math.MaxFloat64, math.MaxFloat64
	this.iLow, this.iUp = -1, -1
	for j := this.i0.GetNext(-1); j != -1; j = this.i0.GetNext(j) {
		if this.errors[j] < this.bUp {
			this.bUp, this.iUp = this.errors[j], j
		}
		if this.errors[j] > this.bLow {
			this.bLow, this.iLow = this.errors[j], j
		}
	}
	if !this.i0.Contains(i1) {
		if this.i3.Contains(i1) || this.i4.Contains(i1) {
			if this.errors[i1] > this.bLow {
				this.bLow, this.iLow = this.errors[i1], i1
			}
		} else {
			if this.errors[i1] < this.bUp {
				this.bUp, this.iUp = this.errors[i1], i1
			}
		}
	}
	if !this.i0.Contains(i2) {
		if this.i3.Contains(i2) || this.i4.Contains(i2) {
			if this.errors[i2] > this.bLow {
				this.bLow, this.iLow = this.errors[i2], i2
			}
		} else {
			if this.errors[i2] < this.bUp {
				this.bUp, this.iUp = this.errors[i2], i2
			}
		}
	}
	if this.iLow == -1 || this.iUp == -1 {
		panic("This should never happen.")
	}
	//fmt.Println(this.alpha,this.bLow,this.bUp,this.iLow,this.iUp, "maxim") //ok so far
	//Made some progress
	return true
}

func (this *BinarySMO) SVMOutput(index int, inst data.Instance) float64 {
	//fmt.Println(inst.Indices(),"SVMoutput")
	//fmt.Println(inst.RealValues(),"SVMoutput")
	//fmt.Println(this.sparseIndices, this.sparseWeights,this.b)
	result := 0.0
	//Is this a linear machine? Is always linear
	//if this.kernelIsLinear {

	//Is weight vector stored in sparse format
	//fmt.Println(len(this.sparseWeights), "sparseweights")
	if this.sparseWeights == nil {
		fmt.Println("in")
		//if this.sparseWeights == nil {
		//fmt.Println("Im here")
		//n1 := len(inst.RealValues())
		for p := range inst.RealValues() {
			//fmt.Println(p, len(this.weights),"this")
			if inst.Index(p) != this.classIndex {
				result += this.weights[inst.Index(p)] * inst.ValueSparse(p)
			}
		}
	} else {
		n1, n2 := len(inst.RealValues()), len(this.sparseWeights)
		for p1, p2 := 0, 0; p1 < n1 && p2 < n2; {
			ind1, ind2 := inst.Index(p1), this.sparseIndices[p2]
			if ind1 == ind2 {
				if ind1 != this.classIndex {
					result += inst.ValueSparse(p1) * this.sparseWeights[p2]
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
	//			result += this.class[i] * this.alpha[i] * this.kernel.Eval(index, i, inst)
	//		}
	//	}
	result -= this.b
	return result
}

// Quick and dirty check whether the quadratic programming problem is solved.
func (bsmo *BinarySMO) CheckClassifier() {
	sum := 0.0
	for i, alp := range bsmo.alpha {
		if alp > 0 {
			sum += bsmo.class[i] / alp
		}
	}
	println("Sum of y(i) * alpha(i): ", sum)

	for i := 0; i < len(bsmo.alpha); i++ {
		output := bsmo.SVMOutput(i, *bsmo.data.Instance(i))
		if utils.Eq(bsmo.alpha[i], 0) {
			if utils.Sm(bsmo.class[i]*output, 1) {
				println("KKT condition 1 violated: ", bsmo.class[i]*output)
			}
		}
		if utils.Gr(bsmo.alpha[i], 0) && utils.Sm(bsmo.alpha[i], bsmo.c_C*bsmo.data.Instance(i).Weight()) {
			if !utils.Eq(bsmo.class[i]*output, 1) {
				println("KKT condition 2 violated: ", bsmo.class[i]*output)
			}
		}
		if utils.Eq(bsmo.alpha[i], bsmo.c_C*bsmo.data.Instance(i).Weight()) {
			if utils.Gr(bsmo.class[i]*output, 1) {
				println("KKT condition 3 violated: ", bsmo.class[i]*output)
			}
		}
	}

}

func (bsmo *BinarySMO) SetKernel(value PolyKernel) {
	bsmo.kernel = value
}

func (bsmo *BinarySMO) Kernel() PolyKernel {
	return bsmo.kernel

}
