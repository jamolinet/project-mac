package utils

import (
	"math"
)


func Log2Binomial(a,b float64) float64 {
	if Gr(b,a) {
		panic("Can't compute binomial coefficient.")
	}
	a1,_ := math.Lgamma(a+1)
	b1,_ := math.Lgamma(b+1)
	ab1,_ := math.Lgamma((a-b)+1)
	return (a1-b1-ab1)/math.Log(2)
}

func Log2Multinomial(a float64,bs []float64) float64 {
	sum :=0.0
	
	for i := range bs {
		if Gr(bs[i],a) {
			panic("Can't compute multinomial coefficient.")
		} else {
			ln,_ := math.Lgamma(bs[i]+1)
			sum = sum + ln 
		}
	}
	ln,_ := math.Lgamma(a+1)
	return (ln-sum)/math.Log(2)
}