package smo

import ()

type CostMatrix struct {
	size   int
	matrix [][]interface{}
}

func NewCostMatrix(numClasses int) CostMatrix {
	var cm CostMatrix
	cm.size = numClasses
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

func (m *CostMatrix) Size() int {
	return m.size
}