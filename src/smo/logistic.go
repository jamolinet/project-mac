package smo

import (
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/functions"
	"github.com/project-mac/src/utils"
	//"github.com/gonum/matrix/mat64"
	"fmt"
	"math"
)

type Logistic struct {
	//The coefficients (optimized parameters) of the model
	par [][]float64
	//The data saved as a matrix
	data [][]float64
	//The number of attributes in the model
	numPredictors int
	//The index of the class attribute
	classIndex int
	//The number of the class labels
	numClasses int
	//The ridge parameter
	ridge float64
	//An attribute filter
	attFilter functions.RemoveUseless
	//The filter used to make attributes numeric
	nominalToBinary functions.NominalToBinary
	//The filter used to get rid of missing values
	replaceMissingValues functions.ReplaceMissingValues
	//Log-likelihood of the searched model
	lL float64
	//The maximum number of iterations
	maxIts    int
	structure data.Instances
	debug     bool
	opt       Optimizer
}

func NewLogistic() Logistic {
	var l Logistic
	l.ridge = 1e-8
	l.maxIts = -1
	return l
}

func (m *Logistic) BuildClassifier(train data.Instances) {
	//remove instances with missing class
	train.DeleteWithMissingClass()

	//Replace missing values
	m.replaceMissingValues = functions.NewReplacingMissingValues()
	m.replaceMissingValues.SetInputFormat(train)
	m.replaceMissingValues.Exec(train)
	train = m.replaceMissingValues.OutputAll()

	//Remove useless attributes
	m.attFilter = functions.NewRemoveUseless()
	m.attFilter.SetInputFormat(train)
	m.attFilter.Exec(train)
	train = m.attFilter.Output()

	//Transform attributes
	m.nominalToBinary = functions.NewNominalToBinary()
	m.nominalToBinary.SetRange("all")
	m.nominalToBinary.SetInputFormat(train)
	m.nominalToBinary.Exec(train)
	train = m.nominalToBinary.OutputAll()

	// Save the structure for printing the model
	m.structure = data.NewInstancesWithInst(train, 0)

	//Extract data
	m.classIndex = train.ClassIndex()
	m.numClasses = train.NumClasses()

	nK := m.numClasses - 1 //Only K-1 class labels needed
	m.numPredictors = train.NumAttributes() - 1
	nR := m.numPredictors
	nC := train.NumInstances()

	m.data = make([][]float64, nC) // Data values
	for i := range m.data {
		m.data[i] = make([]float64, nR+1)
	}
	Y := make([]int, nC)            // Class labels
	xMean := make([]float64, nR+1)  // Attribute means
	xSD := make([]float64, nR+1)    // Attribute stddev's
	sY := make([]float64, nK+1)     // Number of classes
	weights := make([]float64, nC)  // Weights of instances
	totWeights := 0.0               // Total weights of the instances
	m.par = make([][]float64, nR+1) // Optimized parameter values
	for i := range m.par {
		m.par[i] = make([]float64, nK)
	}

	if m.debug {
		fmt.Println("Extracting data...")
	}

	for i := 0; i < nC; i++ {
		// initialize X[][]
		current := train.Instance(i)
		Y[i] = int(current.ClassValue(m.classIndex)) //Class value starts from 0
		weights[i] = current.Weight()                //Dealing with weights
		totWeights += weights[i]

		m.data[i][0] = 1
		j := 1
		for k := 0; k <= nR; k++ {
			if k != m.classIndex {
				x := current.Value(k)
				m.data[i][j] = x
				xMean[j] += weights[i] * x
				xSD[j] += weights[i] * x * x
				j++
			}
		}

		// Class count
		sY[Y[i]]++
	}

	if (totWeights <= 1) && (nC > 1) {
		panic("Sum of weights of instances less than 1, please reweight!")
	}

	xMean[0] = 0
	xSD[0] = 1
	for j := 1; j <= nR; j++ {
		xMean[j] = xMean[j] / totWeights
		if totWeights > 1 {
			xSD[j] = math.Sqrt(math.Abs(xSD[j]-totWeights*xMean[j]*xMean[j]) / (totWeights - 1))
		} else {
			xSD[j] = 0
		}
	}
	if m.debug {
		fmt.Println("Huge intel, line 617 Logistic.java")
	}

	///Normalize input data
	for i := 0; i < nC; i++ {
		for j := 0; j <= nR; j++ {
			if xSD[j] != 0 {
				m.data[i][j] = (m.data[i][j] - xMean[j]) / xSD[j]
			}
		}
	}

	if m.debug {
		fmt.Println("\nIteration History...")
	}

	x := make([]float64, (nR+1)*nK)
	b := make([][]float64, 2) //Boundary constraints, N/A here
	for i := range b {
		b[i] = make([]float64, len(x))
	}

	// Initialize
	for p := 0; p < nK; p++ {
		offset := p * (nR + 1)
		x[offset] = math.Log(sY[p]+1.0) - math.Log(sY[nK]+1.0) // Null model
		b[0][offset] = math.NaN()
		b[1][offset] = math.NaN()
		for q := 1; q <= nR; q++ {
			x[offset+q] = 0.0
			b[0][offset+q] = math.NaN()
			b[1][offset+q] = math.NaN()
		}
	}

	o := NewOptEng(m)
	o.SetWeights(weights)
	o.SetClassLabels(Y)
	opt := NewOptimization(&o)
	opt.SetDebug(m.debug)

	if m.maxIts == -1 { //Search until convergence
		x = opt.findArgmin(x, b)
		for x == nil {
			x = opt.VarbValues()
			if m.debug {
				fmt.Println("200 iterations finished, not enough!")
			}
			x = opt.findArgmin(x, b)
		}
		if m.debug {
			fmt.Println("-------------<Converged>--------------")
		}
	} else {
		opt.SetMaxIteration(m.maxIts)
		x = opt.findArgmin(x, b)
		if x == nil { // Not enough, but use the current value
			x = opt.VarbValues()
		}
	}
	m.lL = -opt.MinFunction() //Log-likelihood

	//Don't need data martix anymore
	m.data = nil

	// Convert coefficients back to non-normalized attribute units
	for i := 0; i < nK; i++ {
		m.par[0][i] = x[i*(nR+1)]
		for j := 1; j <= nR; j++ {
			m.par[j][i] = x[i*(nR+1)+j]
			if xSD[j] != 0 {
				m.par[j][i] /= xSD[j]
				m.par[0][i] -= m.par[j][i] * xMean[j]
			}
		}
	}
}

// Computes the distribution for a given instance.
func (m *Logistic) DistributionForInstance(instance data.Instance) []float64 {

	instance = m.replaceMissingValues.ConvertAndReturn(instance)
	instance = m.attFilter.ConvertAndReturn(instance)
	instance = m.nominalToBinary.ConvertAndReturn(instance)

	//Extract the predictor columns into an array
	instDat := make([]float64, m.numPredictors+1)
	j := 1
	instDat[0] = 1
	for k := 0; k <= m.numPredictors; k++ {
		if k != m.classIndex {
			instDat[j] = instance.Value(k)
			j++
		}
	}
	distribution := m.evaluateProbability(instDat)
	return distribution
}

// Compute the posterior distribution using optimized parameter values
// and the testing instance.
func (m *Logistic) evaluateProbability(data []float64) []float64 {
	prob := make([]float64, m.numClasses)
	v := make([]float64, m.numClasses)

	// Log-posterior before normalizing
	for j := 0; j < m.numClasses-1; j++ {
		for k := 0; k <= m.numPredictors; k++ {
			v[j] += m.par[k][j] * data[k]
		}
	}
	v[m.numClasses-1] = 0

	// Do so to avoid scaling problems
	for m1 := 0; m1 < m.numClasses; m1++ {
		sum := 0.0
		for n := 0; n < m.numClasses-1; n++ {
			sum += math.Exp(v[n] - v[m1])
		}
		prob[m1] = 1 / (sum + math.Exp(-v[m1]))
	}
	return prob
}

func (m *Logistic) SetDebug(debug bool) {
	m.debug = debug
}

func (m *Logistic) SetRidge(ridge float64) {
	m.ridge = ridge
}

func (m *Logistic) SetMaxIts(its int) {
	m.maxIts = its
}

func (m *Logistic) Debug() bool {
	return m.debug
}

func (m *Logistic) Ridge() float64 {
	return m.ridge
}

func (m *Logistic) MaxIts() int {
	return m.maxIts
}

func (m *Logistic) Coefficients() [][]float64 {
	return m.par
}

type OptEng struct {
	logistic *Logistic
	//Weights of instances in the data
	weights []float64
	//Class labels of instances
	cls []int
}

// Evaluate objective function
func (oe *OptEng) objectiveFunction(x []float64) float64 {
	nll := 0.0                           // -LogLikelihood
	dim := oe.logistic.numPredictors + 1 // Number of variables per class

	for i := 0; i < len(oe.cls); i++ { //ith instance
		exp := make([]float64, oe.logistic.numClasses-1)
		var index int
		for offset := 0; offset < oe.logistic.numClasses-1; offset++ {
			index = offset * dim
			for j := 0; j < dim; j++ {
				exp[offset] += oe.logistic.data[i][j] * x[index+j]
			}
		}
		max := exp[utils.MaxIndex(exp)]
		denom := math.Exp(-max)
		var num float64
		if oe.cls[i] == oe.logistic.numClasses-1 { // Class of this instance
			num = -max
		} else {
			num = exp[oe.cls[i]] - max
		}
		for offset := 0; offset < oe.logistic.numClasses-1; offset++ {
			denom += math.Exp(exp[offset] - max)
		}
		nll -= oe.weights[i] * (num - math.Log(denom)) // Weighted NLL
	}

	// Ridge: note that intercepts NOT included
	for offset := 0; offset < oe.logistic.numClasses-1; offset++ {
		for r := 1; r < dim; r++ {
			nll += oe.logistic.ridge * x[offset*dim+r] * x[offset*dim+r]
		}
	}
	return nll
}

// Evaluate Jacobian vector
func (oe *OptEng) evaluateGradient(x []float64) []float64 {
	grad := make([]float64, len(x))
	dim := oe.logistic.numPredictors + 1 // Number of variables per class

	for i := range oe.cls { //ith instance
		num := make([]float64, oe.logistic.numClasses-1) //numerator of [-log(1+sum(exp))]'
		index := 0
		for offset := 0; offset < oe.logistic.numClasses-1; offset++ { //Which part of x
			exp := 0.0
			index = offset * dim
			for j := 0; j < dim; j++ {
				exp += oe.logistic.data[i][j] * x[index+j]
			}
			num[offset] = exp
		}

		max := num[utils.MaxIndex(num)]
		denom := math.Exp(-max) //Denominator of [-log(1+sum(exp))]'
		for offset := 0; offset < oe.logistic.numClasses-1; offset++ {
			num[offset] = math.Exp(num[offset] - max)
			denom += num[offset]
		}
		utils.Normalize(&num, denom)

		//Update denominator of the gradient of -log(Posterior)
		firstTerm := 0.0
		for offset := 0; offset < oe.logistic.numClasses-1; offset++ { //Which part of x
			index = offset * dim
			firstTerm = oe.weights[i] * num[offset]
			for q := 0; q < dim; q++ {
				grad[index+q] += firstTerm + oe.logistic.data[i][q]
			}
		}

		if oe.cls[i] != oe.logistic.numClasses-1 { //Not the last class
			for p := 0; p < dim; p++ {
				grad[oe.cls[i]*dim+p] -= oe.weights[i] * oe.logistic.data[i][p]
			}
		}
	}
	// Ridge: note that intercepts NOT included
	for offset := 0; offset < oe.logistic.numClasses-1; offset++ {
		for r := 1; r < dim; r++ {
			grad[offset*dim+r] += 2 * oe.logistic.ridge * x[offset*dim+r]
		}
	}
	return grad
}

func (oe *OptEng) evaluateHessian(x []float64, index int) []float64 {
	return nil
}

// New OptEng
func NewOptEng(l *Logistic) OptEng {
	var oe OptEng
	oe.logistic = l
	return oe
}

func (oe *OptEng) SetWeights(w []float64) {
	oe.weights = w
}

func (oe *OptEng) SetClassLabels(cls []int) {
	oe.cls = cls
}
