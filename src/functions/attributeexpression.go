package functions

import (
	"github.com/cosn/collections/stack"
	datas "github.com/project-mac/src/data"
	"math"
	"strconv"
	"strings"
)

const (
	/** Supported operators. l = log, b = abs, c = cos, e = exp, s = sqrt,
	  f = floor, h = ceil, r = rint, t = tan, n = sin */
	OPERATORS = "+-*/()^lbcesfhrtn"
	/** Unary functions. l = log, b = abs, c = cos, e = exp, s = sqrt,
	  f = floor, h = ceil, r = rint, t = tan, n = sin */
	UNARY_FUNCTIONS = "lbcesfhrtn"
)

type AttributeOperand struct {
	attributeIndex int64
	negative       bool
}

func NewAttributeOperand(operand string, sign bool) AttributeOperand {
	var ao AttributeOperand
	ao.attributeIndex, _ = strconv.ParseInt(operand[1:], 10, 32)
	ao.negative = sign
	return ao
}

type NumericOperand struct {
	numericConst float64
}

func NewNumericOperand(operand string, sign bool) NumericOperand {
	var no NumericOperand
	no.numericConst, _ = strconv.ParseFloat(operand, 32)
	if sign {
		no.numericConst *= -1.0
	}
	return no
}

type Operator struct {
	operator rune
}

func NewOperator(opp rune) Operator {
	var o Operator
	if isOperator(opp) {
		panic("Unrecognized operator:" + string(opp))
	}
	o.operator = opp
	return o
}

//Apply this operator to the supplied arguments

func (m *Operator) applyOperator(first, second float64) float64 {
	switch m.operator {
	case '+':
		return (first + second)
	case '-':
		return (first - second)
	case '*':
		return (first * second)
	case '/':
		return (first / second)
	case '^':
		return math.Pow(first, second)
	}
	return math.NaN()
}

//Apply this operator (function) to the supplied argument
func (m *Operator) applyFunction(value float64) float64 {
	switch m.operator {
	case 'l':
		return math.Log(value)
	case 'b':
		return math.Abs(value)
	case 'c':
		return math.Cos(value)
	case 'e':
		return math.Exp(value)
	case 's':
		return math.Sqrt(value)
	case 'f':
		return math.Floor(value)
	case 'h':
		return math.Ceil(value)
	case 'r':
		return math_Rint(value)
	case 't':
		return math.Tan(value)
	case 'n':
		return math.Sin(value)
	}
	return math.NaN()
}

func isOperator(tok rune) bool {
	if strings.IndexRune(OPERATORS, tok) == -1 {
		return false
	}
	return true
}

func isUnaryFunction(tok rune) bool {
	if strings.IndexRune(UNARY_FUNCTIONS, tok) == -1 {
		return false
	}
	return true
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

type AttributeExpression struct {
	Operator
	AttributeOperand
	NumericOperand
	operatorStack    stack.S
	originalInfix    string
	postFixExpVector []interface{}
	signMod          bool
	previousTok      string
}

func NewAttributeExpression() AttributeExpression {
	return *new(AttributeExpression)
}

// Handles the processing of an infix operand to postfix
func (m *AttributeExpression) handleOperand(tok string) {
	if strings.IndexRune(tok, 'a') != -1 {
		m.postFixExpVector = append(m.postFixExpVector, NewAttributeOperand(tok, m.signMod))
	} else {
		// should be a numeric constant
		m.postFixExpVector = append(m.postFixExpVector, NewNumericOperand(tok, m.signMod))
	}
	m.signMod = false
}

// Handles the processing of an infix operator to postfix
func (m *AttributeExpression) handleOperator(tok string) {
	push := true

	tokchar := rune(tok[0])
	if tokchar == ')' {
		popop := " "
		do := true
		for do {
			popop = m.operatorStack.Pop().(string)
			if popop[0] != '(' {
				m.postFixExpVector = append(m.postFixExpVector, NewOperator(rune(popop[0])))
			}
			do = popop[0] != '('
		}
	} else {
		infixToc := infixPriority(rune(tok[0]))
		for !m.operatorStack.IsEmpty() && stackPriority(rune(m.operatorStack.Peek().(string)[0])) >= infixToc {

			// try an catch double operators and see if the current one can
			// be interpreted as the sign of an upcoming number
			if len(m.previousTok) == 1 && isOperator(rune(m.previousTok[0])) && m.previousTok[0] != ')' {
				if tok[0] == '-' {
					m.signMod = true
				} else {
					m.signMod = false
				}
				push = false
				break
			} else {
				popop := m.operatorStack.Pop().(string)
				m.postFixExpVector = append(m.postFixExpVector, NewOperator(rune(popop[0])))
			}
		}
		if len(m.postFixExpVector) == 0 {
			if tok[0] == '-' {
				m.signMod = true
				push = false
			}
		}
		if push {
			m.operatorStack.Push(tok)
		}
	}
}

//Evaluate the expression using the supplied Instance.
//   * Assumes that the infix expression has been converted to
//   * postfix and stored in m_postFixExpVector
func (m *AttributeExpression) EvaluateExpression(instance datas.Instance) float64 {
	vals := make([]float64, instance.NumAttributes()+1)
	for i := 0; i < instance.NumAttributes(); i++ {
		if instance.IsMissingValue(i) {
			vals[i] = instance.MissingValue
		} else {
			vals[i] = instance.Value(i)
		}
	}
	m.EvaluateExpressionFloat64(vals)
	return vals[len(vals)-1]
}

// Evaluate the expression using the supplied array of attribute values
func (m *AttributeExpression) EvaluateExpressionFloat64(vals []float64) {
	var operands stack.S
	
	for _,nextob := range m.postFixExpVector {
		if nextob_, ok := nextob.(NumericOperand); ok {
			operands.Push(nextob_.numericConst)
		} else if nextob_, ok := nextob.(AttributeOperand); ok {
			value := vals[nextob_.attributeIndex]
			if nextob_.negative {
				value = -value
			}
			operands.Push(value)
		} else if nextob_, ok := nextob.(Operator); ok {
			op := nextob_.operator
			if isUnaryFunction(op) {
				operand := operands.Pop().(float64)
				result := nextob_.applyFunction(operand)
				operands.Push(result)
			} else {
				second := operands.Pop().(float64)
				first := operands.Pop().(float64)
				result := nextob_.applyOperator(first,second)
				operands.Push(result)
			}
		} else {
			panic("Unknown object in postfix vector!")
		}
	}
	
	if operands.Len() != -1 {
		panic("Problem applying function")
	} 
	
	result := operands.Pop().(float64)
	if math.IsNaN(result) || math.IsInf(result,0) {
		vals[len(vals)-1] = math.NaN()
	} else {
		vals[len(vals)-1] = result
	}
}

//Unfinished: weka.core.AttributeExpression.java, Line 341
func (m *AttributeExpression) ConvertInfixToPostfix(infixExp string) {
	m.originalInfix = infixExp

	infixExp = strings.Replace(infixExp, " ", "", -1)
	infixExp = strings.Replace(infixExp, "log", "l", -1)
	infixExp = strings.Replace(infixExp, "abs", "b", -1)
	infixExp = strings.Replace(infixExp, "cos", "c", -1)
	infixExp = strings.Replace(infixExp, "exp", "e", -1)
	infixExp = strings.Replace(infixExp, "sqrt", "s", -1)
	infixExp = strings.Replace(infixExp, "floor", "f", -1)
	infixExp = strings.Replace(infixExp, "ceil", "h", -1)
	infixExp = strings.Replace(infixExp, "rint", "r", -1)
	infixExp = strings.Replace(infixExp, "tan", "t", -1)
	infixExp = strings.Replace(infixExp, "sin", "n", -1)
	
	tokenizer := stringTokenizer(infixExp)
	m.postFixExpVector = make([]interface{},0)
	for _,tok := range tokenizer {
		
		if len(tok) > 1 {
			m.handleOperand(tok)
		} else {
			// probably an operator, but could be a single char operand
			if isOperator(rune(tok[0])) {
				m.handleOperator(tok)
			} else {
				// should be a numeric constant
				m.handleOperand(tok)
			}
		}
		m.previousTok = tok
	}
	for !m.operatorStack.IsEmpty() {
		popop := m.operatorStack.Pop().(string)
		if popop[0] == '(' || popop[0] == ')' {
			panic("Mis-matched parenthesis!")
		}
		m.postFixExpVector = append(m.postFixExpVector, NewOperator(rune(popop[0])))
	}
}

func stringTokenizer(s string) []string {
	for _, char := range s {
		for _, o := range OPERATORS {
			if char == o {
				s = strings.Replace(s, string(char), "*", -1)
				break
			}
		}
	}
	return strings.Split(s, "*")
}

// Return the infix priority of an operator
func infixPriority(opp rune) int {
	switch opp {
	case 'l':
	case 'b':
	case 'c':
	case 'e':
	case 's':
	case 'f':
	case 'h':
	case 'r':
	case 't':
	case 'n':
		return 3
	case '^':
		return 2
	case '*':
		return 2
	case '/':
		return 2
	case '+':
		return 1
	case '-':
		return 1
	case '(':
		return 4
	case ')':
		return 0
	default:
		panic("Unrecognized operator:" + string(opp))
	}
	return -1
}

// Return the stack priority of an operator
func stackPriority(opp rune) int {
	switch opp {
	case 'l':
	case 'b':
	case 'c':
	case 'e':
	case 's':
	case 'f':
	case 'h':
	case 'r':
	case 't':
	case 'n':
		return 3
	case '^':
		return 2
	case '*':
		return 2
	case '/':
		return 2
	case '+':
		return 1
	case '-':
		return 1
	case '(':
		return 0
	case ')':
		return -1
	default:
		panic("Unrecognized operator:" + string(opp))
	}
	return -1
}

