package smo

import (
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/functions"
	"github.com/gonum/matrix/mat64"
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
	maxIts int
	structure data.Instances
	debug bool
}

func NewLogistic() Logistic {
	var l Logistic
	l.ridge = 1e-8
	l.maxIts = -1
	
	return l
}
