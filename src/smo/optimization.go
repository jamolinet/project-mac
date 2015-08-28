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

	var i, k int
	lenght := len(xold)
	fixedOne := -1           //idx of variable to be fixed
	var alam, alamin float64 //lambda to be found, and its lower bound

	//For convergence and bound test
	var temp, test float64
	alpha := math.Inf(1)
	fold := m.f
	//For cubic interpolation
	var a, b float64
	alam2 := 0.0
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

	if alpha <= m.zero { //Zero
		m.isZeroStep = true
		if m.debug {
			println("print later some statistics")
		}
		return x
	}

	alam = alpha //Always try full feasible newton step
	if alam > 1 {
		alam = 1
	}

	//Iteration of one newton step, if necessary, backtracking is donde
	initF := fold                                           //Initial function value
	hi, lo, newSlope, fhi, flo := alam, alam, 0.0, m.f, m.f //Variables used for beta condition
	var newGrad []float64                                   //Gradient on the new variable values
kloop:
	for k = 0; ; k++ {
		if m.debug {
			println("print later some statistics")
		}
		for i := 0; i < lenght; i++ {
			if !isFixed[i] {
				x[i] = xold[i] + alam*direct[i] //Compute xnew
				if !math.IsNaN(nwsBounds[0][i]) && x[i] < nwsBounds[0][i] {
					x[i] = nwsBounds[0][i] //Rounding error
				} else if !math.IsNaN(nwsBounds[1][i]) && x[i] > nwsBounds[1][i] {
					x[i] = nwsBounds[1][i] //Rounding error
				}
			}
		}

		/* TODO: m.f = objectiveFunction*/ //Compute fnew
		if math.IsNaN(m.f) {
			panic("Objective function value is NaN!")
		}

		for math.IsInf(m.f, 0) { //Avoid infinity
			if m.debug {
				println("print later some statistics")
			}
			alam *= 0.0 //Shrink by half
			if alam <= m.epsilon {
				if m.debug {
					println("print later some statistics")
				}
				return x
			}

			for i := 0; i < lenght; i++ {
				if !isFixed[i] {
					x[i] = xold[i] + alam*direct[i]
				}
			}
			/* TODO: m.f = objectiveFunction*/
			if math.IsNaN(m.f) {
				panic("Objective function value is NaN!")
			}
			initF = math.Inf(1)
		}

		if m.debug {
			println("print later some statistics")
		}

		if m.f <= fold+ALF*alam*m.slope { //Alpah condition: sufficient function decrease
			if m.debug {
				println("print later some statistics")
			}
			/* TODO: newGrad = evaluateGradient(x)*/
			newSlope = 0.0
			for i := 0; i < lenght; i++ {
				if !isFixed[i] {
					newSlope += newGrad[i] * direct[i]
				}
			}
			if newSlope >= BETA*m.slope { //Beta condition: ensure pos. defnty
				if m.debug {
					println("print later some statistics")
				}
				if fixedOne != -1 && alam >= alpha { //HAs bounds and over
					if direct[fixedOne] > 0 {
						x[fixedOne] = nwsBounds[1][fixedOne] //Avoid rounding error
						nwsBounds[1][fixedOne] = math.NaN()  //Add cons. to working set
					} else {
						x[fixedOne] = nwsBounds[0][fixedOne] //Avoid rouding error
						nwsBounds[0][fixedOne] = math.NaN()  //Add cons. to working set
					}

					if m.debug {
						println("print later some statistics")
					}
					isFixed[fixedOne] = true //Fix the variable
					wsBdsIndx = append(wsBdsIndx, fixedOne)
				}
				return x
			} else if k == 0 { //First time: increase alam
				//Search for the smallest value not complying with alpha condition
				upper := math.Min(alpha, maxalam)
				if m.debug {
					println("print later some statistics")
				}
				for !(alam >= upper || m.f > fold+ALF*alam*m.slope) {
					lo = alam
					flo = m.f
					alam *= 2.0
					if alam >= upper { //Avoid roundig errors
						alam = upper
					}

					for i := 0; i < lenght; i++ {
						if !isFixed[i] {
							x[i] = xold[i] + alam*direct[i]
						}
					}
					/* TODO: m.f = objectiveFunction*/
					if math.IsNaN(m.f) {
						panic("Objective function value is NaN!")
					}
					/* TODO: newGrad = evaluateGradient(x)*/
					newSlope = 0.0
					for i := 0; i < lenght; i++ {
						if !isFixed[i] {
							newSlope += newGrad[i] * direct[i]
						}
					}
					if newSlope >= BETA*m.slope {
						if m.debug {
							println("print later some statistics")
						}
						if fixedOne != -1 && alam >= alpha { //Has bounds nad over
							if direct[fixedOne] > 0 {
								x[fixedOne] = nwsBounds[1][fixedOne] //Avoid rounding errors
								nwsBounds[1][fixedOne] = math.NaN()  //Add cons. to working set
							} else {
								x[fixedOne] = nwsBounds[0][fixedOne] //Avoid rounding errors
								nwsBounds[0][fixedOne] = math.NaN()  //Add cons. to working set
							}
							if m.debug {
								println("print later some statistics")
							}
							isFixed[fixedOne] = true //Fix the variable
							wsBdsIndx = append(wsBdsIndx, fixedOne)
						}
						return x
					}
				}
				hi = alam
				fhi = m.f
				break kloop
			} else {
				if m.debug {
					println("print later some statistics")
				}
				hi, lo, flo = alam2, alam, m.f
				break kloop
			}
		} else if alam < alamin { //No feasible lambda found
			if initF < fold {
				alam = math.Min(1.0, alpha)
				for i := 0; i < lenght; i++ {
					if !isFixed[i] {
						x[i] = xold[i] + alam*direct[i] //Still take Alpha
					}
				}
				if m.debug {
					println("print later some statistics")
				}
				if fixedOne != -1 && alam >= alpha { //Has bounds and over
					if direct[fixedOne] > 0 {
						x[fixedOne] = nwsBounds[1][fixedOne] //Avoid rounding errors
						nwsBounds[1][fixedOne] = math.NaN()  //Add cons. to working set
					} else {
						x[fixedOne] = nwsBounds[0][fixedOne] //Avoid rounding errors
						nwsBounds[0][fixedOne] = math.NaN()  //Add cons. to working set
					}
					if m.debug {
						println("print later some statistics")
					}
					isFixed[fixedOne] = true //Fix the variable
					wsBdsIndx = append(wsBdsIndx, fixedOne)
				}
			} else { //Convergence on delta(x)
				for i := 0; i < lenght; i++ {
					x[i] = xold[i]
				}
				m.f = fold
				if m.debug {
					println("print later some statistics")

				}
			}
			return x
		} else { //Backtracking by polynomial interpolation
			if k == 0 { //First time backtrack: quadratic interpolation
				if !math.IsInf(initF, 0) {
					initF = m.f
				}
				//lambda = -g'(0)/(2*g''(0))
				tmplam = -0.5 * alam * m.slope / ((m.f-fold)/alam - m.slope)
			} else { //Subsequent backtrack: cubic interpolation
				rhs1 = m.f - fold - alam*m.slope
				rhs2 = fhi - fold - alam2*m.slope
				a = (rhs1/(alam*alam) - rhs2/(alam2*alam2)) / (alam - alam2)
				b = (-alam2*rhs1/(alam*alam) + alam*rhs2/(alam2*alam2)) / (alam - alam2)
				if a == 0 {
					tmplam = -m.slope / (2.0 * b)
				} else {
					disc = b*b - 3.0*a*m.slope
					if disc < 0 {
						disc = 0
					}
					numerator := -b + math.Sqrt(disc)
					if numerator >= math.MaxFloat64 {
						numerator = math.MaxFloat64
						if m.debug {
							println("print later some statistics")

						}
					}
					tmplam = numerator / (3.0 * a)
				}
				if m.debug {
					println("print later some statistics")

				}
				if tmplam > 0.5*alam {
					tmplam = 0.5 * alam //lambda <= 0.5*lambda_old
				}
			}
		}
		alam2 = alam
		fhi = m.f
		alam = math.Max(tmplam, 0.1*alam) //lambda >= 0.1*lambda_old

		if alam > alpha {
			panic("Sth. wrong in lnsrch")
			//			panic("Sth. wrong in lnsrch:"
			//                        + "Lambda infeasible!(lambda=" + strconv. alam
			//                        + ", alpha=" + alpha + ", upper=" + tmplam
			//                        + "|" + (-alpha * m.slope / (2.0 * ((m.f - fold) / alpha - m.slope)))
			//                        + ", m_f=" + m.f + ", fold=" + fold
			//                        + ", slope=" + m.slope)
		}
	} //End for k := 0; ; k++

	//Quadratic interpolation between lamda values between lo and hi.
	//If cannot find a value satisfying beta condition, use lo
	ldiff := hi - lo
	var lincr float64
	if m.debug {
		println("print later some statistics")

	}

	for newSlope < BETA*m.slope && ldiff >= alamin {
		lincr = -0.5 * newSlope * ldiff * ldiff / (fhi - flo - newSlope*ldiff)

		if m.debug {
			println("print later some statistics")
		}

		if lincr < 0.2*ldiff {
			lincr = 0.2 * ldiff
		}
		alam = lo + lincr
		if alam >= hi { //We cannot go beyond the bounds, so the best we can try is hi
			alam = hi
			lincr = ldiff
		}
		for i = 0; i < lenght; i++ {
			if !isFixed[i] {
				x[i] = xold[i] + alam*direct[i]
			}
		}
		/* TODO: m.f = objectiveFunction(x)*/
		if math.IsNaN(m.f) {
			panic("Objective function value is NaN!")
		}

		if m.f > fold+ALF*alam*m.slope {
			//Alpha condition fails, shrink lambda_upper
			ldiff = lincr
			fhi = m.f
		} else { //Alpha condition holds
			/* TODO: newGrad = evaluateGradient(x)*/
			newSlope = 0.0
			for i := 0; i < lenght; i++ {
				if !isFixed[i] {
					newSlope += newGrad[i] * direct[i]
				}
			}
			if newSlope < BETA*m.slope {
				//Beta condition fails, shrink lambda_lower
				lo = alam
				ldiff -= lincr
				flo = m.f
			}
		}
	}

	if newSlope < BETA*m.slope { //Cannot satisfy beta condition, take lo
		if m.debug {
			println("Beta condition cannot be satisfied, take alpha condition")
		}
		alam = lo
		for i = 0; i < lenght; i++ {
			if !isFixed[i] {
				x[i] = xold[i] + alam*direct[i]
			}
		}
		m.f = flo
	} else if m.debug {
		println("Both alpha and beta conditions are satisfied. alam=")
	}

	if (fixedOne != -1) && (alam >= alpha) { //Has bounds and over
		if direct[fixedOne] > 0 {
			x[fixedOne] = nwsBounds[1][fixedOne] // Avoid rounding error
			nwsBounds[1][fixedOne] = math.NaN()  //Add cons. to working set
		} else {
			x[fixedOne] = nwsBounds[0][fixedOne] // Avoid rounding error
			nwsBounds[0][fixedOne] = math.NaN()  //Add cons. to working set
		}
		if m.debug {
			println("print later some statistics")
		}
		isFixed[fixedOne] = true //Fix the variable
		wsBdsIndx = append(wsBdsIndx, fixedOne)
	}

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
