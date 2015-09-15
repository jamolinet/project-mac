package smo

import (
	"math"
)

type NominalPrediction struct {
	distribution []float64
	actual       float64
	predicted    float64
	weight       float64
}

func NewNominalPrediction(actual float64, distribution []float64) NominalPrediction  {
	var np NominalPrediction
	np.actual = actual
	copy(np.distribution, distribution)
	np.weight = 1
	np.updatePredicted()
	return np
}

func NewNominalPredictionWeight(actual float64, distribution []float64, weight float64) NominalPrediction {
	var np NominalPrediction
	np.actual = actual
	copy(np.distribution, distribution)
	np.weight = weight
	np.updatePredicted()
	return np
}

func (m *NominalPrediction) updatePredicted() {

	predictedClass := -1.0
	bestProb := 0.0
	for i := 0; i < len(m.distribution); i++ {
		if m.distribution[i] > bestProb {
			predictedClass = float64(i)
			bestProb = m.distribution[i]
		}
	}
	
	if predictedClass != 1 {
		m.predicted = predictedClass
	} else {
		m.predicted = math.NaN()
	}
}
