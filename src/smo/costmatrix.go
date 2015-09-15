package smo

import (
	datas "github.com/project-mac/src/data"
	"github.com/project-mac/src/functions"
	"math"
)

type CostMatrix struct {
	size   int
	matrix [][]interface{}
	NotNil bool
}

func NewCostMatrix(numClasses int) CostMatrix {
	var cm CostMatrix
	cm.size = numClasses
	cm.NotNil = true
	return cm
}

func (m *CostMatrix) initialize() {
	m.matrix = make([][]interface{}, m.size)
	for i := range m.matrix {
		m.matrix[i] = make([]interface{}, m.size)
	}
	for i := range m.matrix {
		for j := range m.matrix {
			m.SetCell(i, j, func() interface{} {
				if i == j {
					return 0.0
				} else {
					return 1.0
				}
			})
		}
	}
}

func (m *CostMatrix) SetCell(rowIndex, columnIndex int, value interface{}) {
	m.matrix[rowIndex][columnIndex] = value
}

func (m *CostMatrix) GetCell(rowIndex, columnIndex int) interface{} {
	return m.matrix[rowIndex][columnIndex]
}

func (m *CostMatrix) Size() int {
	return m.size
}

func (m *CostMatrix) GetElement(rowIndex, columnIndex int, inst datas.Instance) float64 {
	if _,ok := m.matrix[rowIndex][columnIndex].(float64); ok {
		return m.matrix[rowIndex][columnIndex].(float64)
	} else if _,ok := m.matrix[rowIndex][columnIndex].(string); ok {
		m.replaceStrings()
	}
	temp,_ := m.matrix[rowIndex][columnIndex].(functions.AttributeExpression)
	return temp.EvaluateExpression(inst)
}

func (m *CostMatrix) GetMaxCostOnlyVals(classVal int) float64 {
	maxCost := math.Inf(-1)

	for i := 0; i < m.size; i++ {
		element := m.GetCell(classVal, i)
		if _, ok := element.(float64); !ok {
			panic("Can't use non-fixed costs when getting max cost.")
		}
		cost, _ := element.(float64)
		if cost > maxCost {
			maxCost = cost
		}
	}
	return maxCost
}

func (m *CostMatrix) GetMaxCost(classVal int, inst datas.Instance) float64 {
	if !m.replaceStrings() {
		return m.GetMaxCostOnlyVals(classVal)
	}

	maxCost := math.Inf(-1)
	var cost float64
	for i := 0; i < m.size; i++ {
		element := m.GetCell(classVal, i)
		if _, ok := element.(float64); !ok {
			temp, _ := element.(functions.AttributeExpression)
			cost = temp.EvaluateExpression(inst)
		} else {
			cost, _ = element.(float64)
		}
		if cost > maxCost {
			maxCost = cost
		}
	}
	return maxCost
}

func (m *CostMatrix) replaceStrings() bool {
	nonFloat64 := false

	for i := 0; i < m.size; i++ {
		for j := 0; j < m.size; j++ {
			if cell, ok := m.GetCell(i, j).(string); ok {
				temp := functions.NewAttributeExpression()
				temp.ConvertInfixToPostfix(cell)
				m.SetCell(i, j, temp)
				nonFloat64 = true
			} else if _, ok := m.GetCell(i, j).(functions.AttributeExpression); ok {
				nonFloat64 = true
			}
		}
	}
	return nonFloat64
}
