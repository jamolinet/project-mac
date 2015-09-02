package smo

import (
	"fmt"
	"github.com/gonum/matrix/mat64"
	"github.com/project-mac/src/utils"
	"math"
)

const (
	
)

type Optimizer interface {
	objectiveFunction([]float64) float64
	evaluateGradient([]float64) []float64
	evaluateHessian([]float64,int) []float64
}

type Optimization struct {
	debug bool
	//function value
	//G'*p
	f, slope float64
	//Used when iteration overflow occurs
	x []float64
	//Test if zero step in lnsrch
	isZeroStep    bool
	epsilon, zero float64
	ALF float64  
	BETA float64 
	TOLX float64  
	STPMX  int
	MAXITS int
	Optimizer
}

func NewOptimization(opt Optimizer) Optimization {
	var o Optimization
	o.Optimizer = opt
	o.isZeroStep = false
	o.debug = false
	o.epsilon = 1
	o.ALF    = 1.0e-4
	o.BETA   = 0.9
	o.TOLX   = 1.0e-6
	o.STPMX  = 100.0
	o.MAXITS = 200
	for 1.0+o.epsilon > 1 {
		o.epsilon /= 2.0
	}
	o.epsilon *= 2.0
	o.zero = math.Sqrt(o.epsilon)
	return o
}

//Find a new point x in the direction p from a point xold at which the
//     * value of the function has decreased sufficiently, the positive
//     * definiteness of B matrix (approximation of the inverse of the Hessian) is
//     * preserved and no bound constraints are violated. Details see "Numerical
//     * Methods for Unconstrained Optimization and Nonlinear Equations". "Numeric
//     * Recipes in C" was also consulted.
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
		alamin = m.TOLX / test
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

		m.f = m.objectiveFunction(x) //Compute fnew
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
			m.f = m.objectiveFunction(x)
			if math.IsNaN(m.f) {
				panic("Objective function value is NaN!")
			}
			initF = math.Inf(1)
		}

		if m.debug {
			println("print later some statistics")
		}

		if m.f <= fold+m.ALF*alam*m.slope { //Alpah condition: sufficient function decrease
			if m.debug {
				println("print later some statistics")
			}
			newGrad = m.evaluateGradient(x)
			newSlope = 0.0
			for i := 0; i < lenght; i++ {
				if !isFixed[i] {
					newSlope += newGrad[i] * direct[i]
				}
			}
			if newSlope >= m.BETA*m.slope { //Beta condition: ensure pos. defnty
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
				for !(alam >= upper || m.f > fold+m.ALF*alam*m.slope) {
					lo = alam
					flo = m.f
					alam *= 2.0
					if alam >= upper { //Avoid rounding errors
						alam = upper
					}

					for i := 0; i < lenght; i++ {
						if !isFixed[i] {
							x[i] = xold[i] + alam*direct[i]
						}
					}
					m.f = m.objectiveFunction(x)
					if math.IsNaN(m.f) {
						panic("Objective function value is NaN!")
					}
					newGrad = m.evaluateGradient(x)
					newSlope = 0.0
					for i := 0; i < lenght; i++ {
						if !isFixed[i] {
							newSlope += newGrad[i] * direct[i]
						}
					}
					if newSlope >= m.BETA*m.slope {
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

	for newSlope < m.BETA*m.slope && ldiff >= alamin {
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
		m.f = m.objectiveFunction(x)
		if math.IsNaN(m.f) {
			panic("Objective function value is NaN!")
		}

		if m.f > fold+m.ALF*alam*m.slope {
			//Alpha condition fails, shrink lambda_upper
			ldiff = lincr
			fhi = m.f
		} else { //Alpha condition holds
			newGrad = m.evaluateGradient(x)
			newSlope = 0.0
			for i := 0; i < lenght; i++ {
				if !isFixed[i] {
					newSlope += newGrad[i] * direct[i]
				}
			}
			if newSlope < m.BETA*m.slope {
				//Beta condition fails, shrink lambda_lower
				lo = alam
				ldiff -= lincr
				flo = m.f
			}
		}
	}

	if newSlope < m.BETA*m.slope { //Cannot satisfy beta condition, take lo
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

// Main algorithm. Descriptions see "Practical Optimization"
func (m *Optimization) findArgmin(initX []float64, constraints [][]float64) []float64 {
	l := len(initX)

	//Initially all variables are free, all bounds are constraints of
	//non-working-set constraints
	isFixed := make([]bool, l)
	nwsBounds := make([][]float64, 2)
	nwsBounds[0] = make([]float64, l)
	nwsBounds[1] = make([]float64, l)
	//Record indice of fixed variables, simply for efficiency
	wsBdsIndx := make([]int, len(constraints))
	//Vectors used to record the variable indices to be freed
	var toFree, oldToFree []int

	//Initial value of obj. function, gradient and inverse of the Hessian
	m.f = m.objectiveFunction(initX)
	if math.IsNaN(m.f) {
		panic("Objective function value is NaN!")
	}

	sum := 0.0
	grad := m.evaluateGradient(initX)
	var oldGrad, oldX []float64
	deltaGrad := make([]float64, l)
	deltaX := make([]float64, l)
	direct := make([]float64, l)
	x := make([]float64, l)
	L := mat64.NewDense(l, l, nil) //Lower triangle of Cholesky factor
	D := make([]float64, l)        //Diagonal of Cholesky factor
	for i := 0; i < l; i++ {
		L.SetRow(i, make([]float64, l))
		L.Set(i, i, 1.0)
		D[i] = 1.0
		direct[i] = -grad[i]
		sum += grad[i] * grad[i]
		x[i] = initX[i]
		nwsBounds[0][i] = constraints[0][i]
		nwsBounds[1][i] = constraints[1][i]
		isFixed[i] = false
	}
	stpmax := float64(m.STPMX) * math.Max(math.Sqrt(sum), float64(l))

	//iterates:
	for step := 0; step < m.MAXITS; step++ {
		if m.debug {
			println("print later some statistics")
		}
		//Try at most one feasible newton step, i.e. 0<lamda<=alpha
		oldX = x
		oldGrad = grad

		//Also update grad
		if m.debug {
			println("Line search...")
		}
		m.isZeroStep = false
		x = m.Lnsrch(x, grad, direct, stpmax, isFixed, nwsBounds, wsBdsIndx)
		if m.debug {
			println("Line search finished.")
		}

		if m.isZeroStep { //Zero step, simply delete rows/cols of D and L
			for f := 0; f < len(wsBdsIndx); f++ {
				idx := wsBdsIndx[f]
				L.SetRow(idx, make([]float64, l))
				L.SetCol(idx, make([]float64, l))
				D[idx] = 0.0
			}
			grad = m.evaluateGradient(x)
			step--
		} else {
			//Check converge on x
			finish := false
			test := 0.0
			for h := 0; h < l; h++ {
				deltaX[h] = x[h] - oldX[h]
				tmp := math.Abs(deltaX[h]) / math.Max(math.Abs(x[h]), 1.0)
				if tmp > test {
					test = tmp
				}
			}
			if test < m.zero {
				if m.debug {
					println("print later some statistics")
				}
				finish = true
			}
			//Check zero gradient
			grad = m.evaluateGradient(x)
			test = 0.0
			denom, dxSq, dgSq, newlyBounded := 0.0, 0.0, 0.0, 0.0
			for g := 0; g < l; g++ {
				if !isFixed[g] {
					deltaGrad[g] = grad[g] - oldGrad[g]
					// Calculate the denominators
					denom += deltaX[g] * deltaGrad[g]
					dxSq += deltaX[g] * deltaX[g]
					dgSq += deltaGrad[g] * deltaGrad[g]
				} else { //Only newly bounded variables will be non-zero
					newlyBounded += deltaX[g] * (grad[g] - oldGrad[g])
				}

				// Note: CANNOT use projected gradient for testing
				// convergence because of newly bounded variables
				tmp := math.Abs(grad[g]) * math.Max(math.Abs(direct[g]), 1.0) / math.Max(math.Abs(m.f), 1.0)
				if tmp > test {
					test = tmp
				}
			}
			if test < m.zero {
				if m.debug {
					println("print later some statistics")
				}
				finish = true
			}

			// dg'*dx could be < 0 using inexact Lnsrch
			if m.debug {
				println("dg'*dx=", (denom + newlyBounded))
			}
			//dg'*dx = 0
			if math.Abs(denom+newlyBounded) < m.zero {
				finish = true
			}

			size := len(wsBdsIndx)
			isUpdate := true //Whether to update BFGS formula
			//Converge: check whether release any current constraints
			if finish {
				if m.debug {
					println("print later some statistics")
				}

				if toFree != nil {
					oldToFree = toFree[0:]
				}
				toFree = make([]int, len(wsBdsIndx))

				for m1 := size - 1; m1 >= 0; m1-- {
					index := wsBdsIndx[m1]
					hessian := m.evaluateHessian(x, index)
					deltaL := 0.0
					if hessian != nil {
						for mm := 0; mm < len(hessian); mm++ {
							if !isFixed[mm] { //Free variable
								deltaL += hessian[mm] * direct[mm]
							}
						}
					}

					// First and second order Lagrangian multiplier estimate
					// If user didn't provide Hessian, use first-order only
					var L1, L2 float64
					if x[index] >= constraints[1][index] { //Upper bound
						L1 = -grad[index]
					} else if x[index] <= constraints[0][index] { // Lower bound
						L1 = grad[index]
					} else {
						panic(fmt.Errorf("x[%i] not fixed on the"+" bounds where it should have been!", index))
					}
					//L2 = L1 + deltaL
					L2 = L1 + deltaL
					if m.debug {
						println("print later some statistics")
					}
					//Check validity of Lagrangian multiplier estimate
					isConverge := (2.0 * math.Abs(deltaL)) < math.Min(math.Abs(L1), math.Abs(L2))
					if (L1*L2 > 0.0) && isConverge { //Same sign and converge: valid
						if L2 < 0.0 { // Negative Lagrangian: feasible
							toFree = append(toFree, index)
							copy(wsBdsIndx[m1:], wsBdsIndx[m1+1:])
							//wsBdsIndx = delete(wsBdsIndx, m)
							finish = false // Not optimal, cannot finish
						}
					}

					// Although hardly happen, better check it
					// If the first-order Lagrangian multiplier estimate is wrong,
					// avoid zigzagging
					if hessian == nil && toFree != nil && m.equals(toFree, oldToFree) {
						finish = true
					}
				}

				if finish { //Min. found
					if m.debug {
						println("Min found.")
					}
					m.f = m.objectiveFunction(x)
					if math.IsNaN(m.f) {
						panic("Objective function value is NaN!")
						return x
					}
				}

				//Free some variables
				for mmm := 0; mmm < len(toFree); mmm++ {
					freeIndx := toFree[mmm]
					isFixed[freeIndx] = false                    //Free this variable
					if x[freeIndx] <= constraints[0][freeIndx] { // Lower bound
						nwsBounds[0][freeIndx] = constraints[0][freeIndx]
						if m.debug {
							println("print later some statistics")
						}
					} else { // Upper bound
						nwsBounds[1][freeIndx] = constraints[1][freeIndx]
						if m.debug {
							println("print later some statistics")
						}
					}
					L.Set(freeIndx, freeIndx, 1.0)
					D[freeIndx] = 1.0
					isUpdate = false
				}
			}

			if denom < math.Max(m.zero*math.Sqrt(dxSq)*math.Sqrt(dgSq), m.zero) {
				if m.debug {
					println("dg'*dx negative!")
				}
				isUpdate = false //Do not update
			}
			//If Hessian will be positive definite, update it
			if isUpdate {
				//modify once: dg*dg'/(dg'*dx)
				coeff := 1.0 / denom // 1/(dg'*dx)
				m.updateCholeskyFactor(L, D, deltaGrad, coeff, isFixed)

				//modify twice: g*g'/(g'*p)
				coeff = 1.0 / m.slope // 1/(g'*p)
				m.updateCholeskyFactor(L, D, oldGrad, coeff, isFixed)

			}
		}

		//Find new direction
		LD := mat64.NewDense(l, l, nil) //L*D
		b := make([]float64, l)

		for k := 0; k < l; k++ {
			if !isFixed[k] {
				b[k] = -grad[k]
			} else {
				b[k] = 0.0
			}

			for j := k; j < l; j++ { // Lower triangle
				if !isFixed[j] && !isFixed[k] {
					LD.Set(j, k, L.At(j, k)*D[k])
				}
			}
		}

		//Solve (LD)*y = -g, where y=L'*direct
		LDIR := SolveTriangle(*LD, b, true, isFixed)
		LD = nil

		for n := 0; n < len(LDIR); n++ {
			if math.IsNaN(LDIR[n]) {
				panic("Error")
			}
		}

		//Solve L'*direct = y
		direct = SolveTriangle(*L, LDIR, false, isFixed)
		for n := 0; n < len(direct); n++ {
			if math.IsNaN(direct[n]) {
				panic("direct is NaN!")
			}
		}
	}

	if m.debug {
		println("Cannot find minimum -- too many iterations!")
	}

	m.x = x

	return nil

}


func (m *Optimization) updateCholeskyFactor(L *mat64.Dense, D []float64, v []float64, coeff float64, isFixed []bool) {
	var t, p, b float64
	n := len(v)
	vp := make([]float64, n)
	for i := 0; i < len(v); i++ {
		if !isFixed[i] {
			vp[i] = v[i]
		} else {
			vp[i] = 0.0
		}
	}

	if coeff > 0.0 {
		t = coeff
		for j := 0; j < n; j++ {
			if isFixed[j] {
				continue
			}

			p = vp[j]
			d := D[j]
			dbarj := d + t*p*p
			D[j] = dbarj

			b = p * t / dbarj
			t *= d / dbarj
			for r := j + 1; r < n; r++ {
				if !isFixed[r] {
					l := L.At(r, j)
					vp[r] -= p * l
					L.Set(r, j, l+b*vp[r])
				} else {
					L.Set(r, j, 0.0)
				}
			}
		}
	} else {
		P := SolveTriangle(*L, v, true, isFixed)
		t = 0.0
		for i := 0; i < n; i++ {
			if !isFixed[i] {
				t += P[i] * P[i] / D[i]
			}
		}

		sqrt := 1.0 + coeff*t
		if sqrt < 0 {
			sqrt = 0.0
		} else {
			sqrt = math.Sqrt(sqrt)
		}

		alpha, sigma := coeff, coeff/(1.0+sqrt)
		var rho, theta float64

		for j := 0; j < n; j++ {
			if isFixed[j] {
				continue
			}

			d := D[j]
			p = P[j] * P[j] / d
			theta = 1.0 + sigma*p
			t -= p
			if t < 0.0 {
				t = 0.0 // Rounding error
			}
			plus := sigma * sigma * p * t
			if (j < n-1) && (plus <= m.zero) {
				plus = m.zero // Avoid rounding error
			}
			rho = theta*theta + plus
			D[j] = rho * d

			if math.IsNaN(D[j]) {
				panic(fmt.Errorf("print error"))
			}

			b = alpha * P[j] / (rho * d)
			alpha /= rho
			rho = math.Sqrt(rho)
			//sigmaOld := sigma
			sigma *= (1.0 + rho) / (rho * (theta + rho))
			if (j < n-1) && (math.IsNaN(sigma) || math.IsInf(sigma, 0)) {
				panic(fmt.Errorf("print error"))
			}

			for r := j + 1; r < n; r++ {
				if !isFixed[r] {
					l := L.At(r, j)
					vp[r] -= P[j] * l
					L.Set(r, j, l+b*vp[r])
				} else {
					L.Set(r, j, 0.0)
				}
			}
		}
	}
}

// Solve the linear equation of TX=B where T is a triangle matrix It can be
// solved using back/forward substitution, with O(N^2) complexity.
func SolveTriangle(t mat64.Dense, b []float64, isLower bool, isZero []bool) []float64 {
	n := len(b)
	result := make([]float64, n)
	if isZero == nil {
		isZero = make([]bool, n)
	}

	if isLower { //lower triangle, fordward-substitution
		j := 0
		for j < n && isZero[j] {
			result[j] = 0
			j++
		} // go to the first row

		if j < n {
			result[j] = b[j] / t.At(j, j)

			for ; j < n; j++ {
				if !isZero[j] {
					numerator := b[j]
					for k := 0; k < j; k++ {
						numerator -= t.At(j, k) * result[k]
					}
					result[j] = numerator / t.At(j, j)
				} else {
					result[j] = 0
				}
			}
		}
	} else { // Upper triangle, back-substitution
		j := n - 1
		for (j >= 0) && isZero[j] {
			result[j] = 0.0
			j--
		} // go to the last row
		if j >= 0 {
			result[j] = b[j] / t.At(j, j)

			for ; j >= 0; j-- {
				if !isZero[j] {
					numerator := b[j]
					for k := j + 1; k < n; k++ {
						numerator -= t.At(k, j) * result[k]
					}
					result[j] = numerator / t.At(j, j)
				} else {
					result[j] = 0.0
				}
			}
		}
	}
	return result
}

//func (m *Optimization) evaluateHessian(x []float64, index int) []float64 {
//	return nil
//}

func (m *Optimization) SetMaxIteration(i int) {
	m.MAXITS = i
}

func (m *Optimization) SetDebug(debug bool) {
	m.debug = debug
}

func (m *Optimization) MinFunction() float64 {
	return m.f
}

func (m *Optimization) VarbValues() []float64 {
	return m.x
}
