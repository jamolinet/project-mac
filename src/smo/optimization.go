package smo

import (
	"github.com/project-mac/src/utils"
	"math"
)

const (
	ALF    = 1.0e-4
	BETA   = 0.9
	TOLX   = 1.0e-6
	STPMX  = 100.0
	MAXITS = 200
)

type Optimization struct {
	debug bool
	//function value
	//G'*p
	f, slope float64
	//Used when iteration overflow occurs
	x []int
	//Test if zero step in lnsrch
	isZeroStep    bool
	epsilon, zero float64
}

func NewOptimization() Optimization {
	var o Optimization
	o.isZeroStep = false
	o.debug = false
	o.epsilon = 1
	for 1.0+o.epsilon > 1 {
		o.epsilon /= 2.0
	}
	o.epsilon *= 2.0
	o.zero = math.Sqrt(o.epsilon)
	return o
}

func (m *Optimization) Lnsrch(xold, gradient, direct []float64, stpmax float64, isFixed []bool, nwsBounds [][]float64, wsBdsIndx []int) []float64 {

	var i, k float64
	lenght := len(xold)
	fixedOne := -1           //idx of variable to be fixed
	var alam, alamin float64 //lambda to be found, and its lower bound

	//For convergence and bound test
	var temp, test float64
	alpha := math.Inf(1)
	fold := m.f
	//For cubic interpolation
	var a, b float64
	alama := 0.0
	disc, maxalam := 0.0, 1.0
	var rhs1, rhs2, tmplam float64
	var sum float64
	x := make([]float64, lenght) //New variable values

	//Scale step
	for sum, i := 0.0, 0; i < lenght; i++ {
		if !isFixed[i] { //For fixed variables, direction = 0
			sum += math.Pow(direct[i], 2)
		}
	}

	sum = math.Sqrt(sum)
	if m.debug {
		println("print later some statistics")
	}
	if sum > stpmax {
		for i := 0; i < lenght; i++ {
			if !isFixed[i] {
				direct[i] *= stpmax / sum
			}
		}
	} else {
		maxalam = stpmax / sum
	}

	//Compute intial rate of decrease, g'*d
	m.slope = 0.0
	for i := 0; i < lenght; i++ {
		x[i] = xold[i]
		if !isFixed[i] {
			m.slope += gradient[i] * direct[i]
		}
	}

	if m.debug {
		println("print later some more statistics")
	}

	//Slope too small
	if math.Abs(m.slope) <= m.zero {
		if m.debug {
			println("Gradient and direction orthogonal -- " +
				"Min. found with current fixed variables" +
				" (or all variables fixed). Try to release" +
				" some variables now.")
		}
		return x
	}

	//Err: slope > 0
	if m.slope > m.zero {
		println("print later some more statistics inside a loop")
		panic("g'*p positive!")
	}

	//Compute LAMBDAmin and upper bound of lambda--alpha
	test = 0.0
	for i := 0; i < lenght; i++ {
		if !isFixed[i] { //No need for fixed variables
			temp = math.Abs(direct[i]) / math.Max(math.Abs(x[i]), 1.0)
			if temp > test {
				test = temp
			}

		}
	}

	if test > m.zero { //Not converge
		alamin = TOLX / test
	} else {
		if m.debug {
			println("print later some statistics")
		}
		return x
	}

	//Check whether any non-working-set bounds are "binding"
	for i := 0; i < lenght; i++ {
		if !isFixed[i] { //No need for fixed variables
			var alpi float64
			if direct[i] < -m.epsilon && !math.IsNaN(nwsBounds[0][i]) { //Not feasible
				alpi = (nwsBounds[0][i] - xold[i]) / direct[i]
				if alpi <= m.zero { //Zero
					if m.debug {
						println("print later some statistics")
					}
					x[i] = nwsBounds[0][i]
					isFixed[i] = true //Fix this variable
					alpha = 0
					nwsBounds[0][i] = math.NaN() //Add cons. to working set
					wsBdsIndx = append(wsBdsIndx, i)
				} else if alpha > alpi { //Fix one variable in on iteration
					alpha = alpi
					fixedOne = i
				}
			} else if direct[i] > m.epsilon && !math.IsNaN(nwsBounds[1][i]) { //Not feasible
				alpi = (nwsBounds[1][i] - xold[i]) / direct[i]
				if alpi <= m.zero { //Zero
					if m.debug {
						println("print later some statistics")
					}
					x[i] = nwsBounds[1][i]
					isFixed[i] = true //Fix this variable
					alpha = 0
					nwsBounds[1][i] = math.NaN() //Add cons. to working set
					wsBdsIndx = append(wsBdsIndx, i)
				} else if alpha > alpi { //Fix one variable in on iteration
					alpha = alpi
					fixedOne = i
				}
			}
		}
	}

	if m.debug {
		println("print later some statistics")
	}
		
	/* TODO: Continue from line 481 Optimization.java*/	

	return x
}

func (m *Optimization) equals(a, b []int) bool {
	if (a == nil || b == nil) || (len(a) != len(b)) {
		return false
	}
	sorta, sortb := utils.SortInt(a), utils.SortInt(b)
	for j := 0; j < len(a); j++ {
		if a[sorta[j]] != b[sortb[j]] {
			return false
		}
	}
	return true
}

//type DynamicIntArray struct {
//	objects            []int
//	size               int
//	capacityIncrement  int
//	capacityMultiplier int
//}
//
//func NewDynamicIntArray(capacity int) {
//	var dia DynamicIntArray
//	dia.objects = make([]int, capacity)
//	dia.capacityIncrement = 1
//	dia.capacityMultiplier = 2
//}
