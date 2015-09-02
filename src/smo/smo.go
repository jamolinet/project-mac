package smo

import (
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/functions"
	"math"
)

const (
	//filter: Normalize training data
	FILTER_NORMALIZE = 0
	//filter: Standardize training data
	FILTER_STANDARDIZE = 1
	//filter: No normalization/standardization
	FILTER_NONE = 2
)

/* TODO: TO MAKE IT Go IDIOMATIC FRIENDLY IMPLEMENT IT LIKE A EMBEDDING FIELD*/
//type SMO struct {
//The binary classifier(s)
var classifiers [][]BinarySMO = nil

//The complexity parameter
var c_C = 1.0

//Epsilon for rounding
var eps = 1.0e-12

//Tolerance for accuracy of result
var tol = 1.0e-3

//Whether to normalize/standardize/neither
var filterType = FILTER_NORMALIZE

/** Remains other attributes to declare*/
/**********/ /////
//The filter used to make attributes numeric
var nominalToBinary functions.NominalToBinary

//The filter used to standardize/normalize all values

//The filter used to get rid of missing values
var missing functions.ReplaceMissingValues

//The class index from the training data
var classIndex = -1

//The class attribute
var classAttribute data.Attribute

//whether the kernel is a linear one
var kernelIsLinear = false

//	Turn off all checks and conversions? Turning them off assumes
//      that data is purely numeric, doesn't contain any missing values,
//      and has a nominal class. Turning them off also means that
//      no header information will be stored if the machine is linear.
//      Finally, it also assumes that no instance has a weight equal to 0
var checksTurnedOff bool

//Precision constant for updating sets
var del = -math.MaxFloat64

//Whether logistic models are to be fit
var fitLogisticModels = false

//The number of folds for the internal cross-validation
var numFolds = 1

//The random number seed
var randomSeed = 1

//the kernel to use
var kernel PolyKernel

//}

func TurnChecksOn() {
	checksTurnedOff = false
}

func TurnChecksOff() {
	checksTurnedOff = true
}

func BuildClassifier(insts data.Instances) {

	if !checksTurnedOff {
		//Remove instances with missing value
		insts.DeleteWithMissingClass()

		/* Removes all the instances with weight equal to 0.
		   MUST be done since condition (8) of Keerthi's paper
		   is made with the assertion Ci > 0 (See equation (3a). */
		data := data.NewInstancesWithInst(insts, insts.NumInstances())
		for _, instance := range insts.Instances() {
			if instance.Weight() > 0 {
				data.Add(instance)
			}
		}
		if data.NumInstances() == 0 {
			panic("No training instances left after removing " +
				"instances with weight 0!")
		}
		insts = data
	}
}

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
}

func NewBinarySMO() BinarySMO {
	return *new(BinarySMO)
}

// Method for building the binary classifier
func (bsmo *BinarySMO) buildClassifier(insts data.Instances, cl1, cl2 int, fitLogistic bool, numFolds int, randomSeed int) {

	//Initialize some variables
	bsmo.bUp, bsmo.bLow, bsmo.b = -1, 1, 0
	bsmo.alpha, bsmo.weights, bsmo.errors = nil, nil, nil
	bsmo.sparseWeights, bsmo.sparseIndices = nil, nil

	//Store the sum of weights
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

	//Check whether one or both classes are missing
	if bsmo.iUp == -1 || bsmo.iLow == -1 {
		if bsmo.iUp != -1 {
			bsmo.b = -1
		} else if bsmo.iLow != -1 {
			bsmo.b = 1
		} else {
			bsmo.class = nil
			return
		}
		if kernelIsLinear {
			bsmo.sparseWeights = make([]float64, 0, 0)
			bsmo.sparseIndices = make([]int, 0, 0)
			bsmo.class = nil
		} else {
			bsmo.supportVectors = NewSMOSet(0)
			bsmo.alpha = make([]float64, 0, 0)
			bsmo.class = make([]float64, 0, 0)
		}

		//Fit sigmoind if requested
		if fitLogistic {
			//Call method fitLogistic if implemented
		}
		return
	}

	//Set the reference to the data
	bsmo.data = insts

	//If machine is linear, reserve space for weights
	if kernelIsLinear {
		bsmo.weights = make([]float64, bsmo.data.NumAttributes())
	} else {
		bsmo.weights = nil
	}

	//Initialize alpha array to zero
	bsmo.alpha = make([]float64, bsmo.data.NumInstances())

	//Initialize sets
	bsmo.supportVectors = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i0 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i1 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i2 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i3 = NewSMOSet(bsmo.data.NumInstances())
	bsmo.i4 = NewSMOSet(bsmo.data.NumInstances())

	//CLean out some instances variables
	bsmo.sparseWeights = nil
	bsmo.sparseIndices = nil

	//Init kernel
	bsmo.kernel.BuildKernel(bsmo.data)

	//Initialize error cache
	bsmo.errors = make([]float64, bsmo.data.NumInstances())
	bsmo.errors[bsmo.iLow], bsmo.errors[bsmo.iUp] = 1, -1

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
			for i := range bsmo.alpha {
				if bsmo.ExamineExample(i) {
					numChanged++
				}
			}
		} else {

			//Tis code implements Modification 1 from Keerthi et al.'s paper
			for i, alpha := range bsmo.alpha {
				if alpha > 0 && alpha < c_C*bsmo.data.Instance(i).Weight() {
					if bsmo.ExamineExample(i) {
						numChanged++
					}

					//Is optimality on unbound vectors obtained?
					if bsmo.bUp > bsmo.bLow-2*tol {
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
	bsmo.kernel.Clean()

	bsmo.errors = nil
	// bsmo.i0, bsmo.i1, bsmo.i2, bsmo.i3, bsmo.i4 will be set to nil too

	//If machine is linear, delete training data and store weight vector in sparse format
	if kernelIsLinear {
		//Don't need to store the set of support vectors
		//bsmo.supportVectors =  nil

		//Don't need to store the class values either
		bsmo.class = nil

		//Clean out training data
		if !checksTurnedOff {
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
		bsmo.sparseWeights = make([]float64, counter)
		bsmo.sparseIndices = make([]int, counter)
		copy(bsmo.sparseWeights, sparseWeights[:counter])
		copy(bsmo.sparseIndices, sparseIndices[:counter])
		//Clean out weights vector
		bsmo.weights = nil
		//Don't need the alphas in the linear case
		bsmo.alpha = nil
	}

	//Fit sigmoid if requested
	if fitLogistic {
		//Call method fitLogistic if implemented
	}
}

// Examines instance
func (this *BinarySMO) ExamineExample(i2 int) bool {
	var y2, F2 float64
	i1 := 1

	y2 = this.class[i2]
	if this.i0.Contains(i2) {
		F2 = this.errors[i2]
	} else {
		F2 = this.SVMOutput(i2, *this.data.Instance(i2)) + this.b - y2
		this.errors[i2] = F2

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
		if this.bLow-F2 > 2*tol {
			optimal, i1 = false, this.iLow
		}
	}
	if this.i0.Contains(i2) || this.i3.Contains(i2) || this.i4.Contains(i2) {
		if this.bLow-F2 > 2*tol {
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
	var alph1, alph2, y1, y2, F1, s, L, H, k11, k12, k22, eta, a1, a2, f1, f2, v1, v2, Lobj, Hobj float64
	C1 := c_C * this.data.Instance(i1).Weight()
	C2 := c_C * this.data.Instance(i2).Weight()

	//Don't do anything if the two instances are the same
	if i1 == i2 {
		return false
	}

	//Initialize variables
	alph1, alph2 = this.alpha[i1], this.alpha[i2]
	y1, y2 = this.class[i1], this.class[i2]
	F1 = this.errors[i1]
	s = y1 * y2

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
	k11 = this.kernel.Eval(i1, i1, *this.data.Instance(i1))
	k12 = this.kernel.Eval(i1, i2, *this.data.Instance(i1))
	k22 = this.kernel.Eval(i2, i2, *this.data.Instance(i2))
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
		f1 = this.SVMOutput(i1, *this.data.Instance(i1))
		f2 = this.SVMOutput(i2, *this.data.Instance(i2))
		v1 = f1 + this.b - y1*alph1*k11 - y2*alph2*k12
		v2 = f2 + this.b - y1*alph1*k12 - y2*alph2*k22
		gamma := alph1 + s*alph2
		Lobj = (gamma - s*L) + L - 0.5*k11*(gamma-s*L)*(gamma-s*L) -
			0.5*k22*L*L - s*k12*(gamma-s*L)*L -
			y1*(gamma-s*L)*v1 - y2*L*v2
		Hobj = (gamma - s*H) + H - 0.5*k11*(gamma-s*H)*(gamma-s*H) -
			0.5*k22*H*H - s*k12*(gamma-s*H)*H -
			y1*(gamma-s*H)*v1 - y2*H*v2
		if Lobj > Hobj+eps {
			a2 = L
		} else if Lobj < Hobj-eps {
			a2 = H
		} else {
			a2 = alph2
		}
	}
	if math.Abs(a2-alph2) < eps*(a2+alph2+eps) {
		return false
	}

	//To prevent precision problems
	if a2 > C2-del*C2 {
		a2 = C2
	} else if a2 <= del*C2 {
		a2 = 0
	}

	//Recompute a1
	a1 = alph1 + s*(alph2-a2)

	//To prevent precision problems
	if a1 > C1-del*C1 {
		a1 = C1
	} else if a1 <= del*C1 {
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
	if kernelIsLinear {
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
	}

	//Update error cache using new Langrage multipliers
	for j := this.i0.GetNext(-1); j != -1; j = this.i0.GetNext(j) {
		if j != i1 && j != i2 {
			this.errors[j] +=
				y1*(a1-alph1)*this.kernel.Eval(i1, j, *this.data.Instance(i1)) +
					y2*(a2-alph2)*this.kernel.Eval(i2, j, *this.data.Instance(i2))
		}
	}

	//Update error cache for i1 and i2
	this.errors[i1] += y1*(a1-alph1)*k11 + y2*(a2-alph2)*k12
	this.errors[i2] += y1*(a1-alph1)*k12 + y2*(a2-alph2)*k22

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

	//Made some progress
	return true
}

func (this *BinarySMO) SVMOutput(index int, inst data.Instance) float64 {
	result := 0.0
	//Is this a linear machine?
	if kernelIsLinear {

		//Is weight vector stored in sparse format
		if this.sparseWeights == nil {
			//n1 := len(inst.RealValues())
			for p := range inst.RealValues() {
				if inst.Index(p) != classIndex {
					result += this.weights[inst.Index(p)] * inst.ValueSparse(p)
				}
			}
		} else {
			n1, n2 := len(inst.RealValues()), len(this.sparseWeights)
			for p1, p2 := 0, 0; p1 < n1 && p2 < n2; {
				ind1, ind2 := inst.Index(p1), this.sparseIndices[p2]
				if ind1 == ind2 {
					if ind2 != classIndex {
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
	} else {
		for i := this.supportVectors.GetNext(-1); i != -1; i = this.supportVectors.GetNext(i) {
			result += this.class[i] * this.alpha[i] * this.kernel.Eval(index, i, inst)
		}
	}
	result -= this.b
	return result
}

func (bsmo *BinarySMO) SetKernel(value PolyKernel) {
	bsmo.kernel = value
}

func (bsmo *BinarySMO) Kernel() PolyKernel {
	return bsmo.kernel

}

//-------------------------------------------------
//-------------------------------------------------
//-------------------------------------------------
//-------------------------------------------------
//-------------------------------------------------
//-------------------------------------------------
//-------------------------------------------------
//-------------------------------------------------
//-------------------------------------------------
