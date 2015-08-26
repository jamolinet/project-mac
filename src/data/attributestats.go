package data

import (
	"github.com/project-mac/src/utils"
)

type AttributeStats struct {
	/*
		The number of int-like values
		The number of real-like values (i.e. have a fractional part)
		The number of missing values
		The number of distinct values
		The number of values that only appear once
		The total number of values (i.e. number of instances)
	*/
	IntCount, RealCount, MissingCount,
	DistinctCount, UniqueCount, TotalCount int
	//Counts of each nominal value
	NominalCounts []int
	numericStats  Stats
}

func NewAttributeStats() AttributeStats {
	var as AttributeStats
	as.IntCount, as.RealCount, as.MissingCount, as.DistinctCount, as.UniqueCount, as.TotalCount =
		0, 0, 0, 0, 0, 0
	return as
}

// Updates the counters for one more observed distinct value
func (as *AttributeStats) AddDistinct(value float64, count int) {
	if count > 0 {
		if count == 1 {
			as.UniqueCount++
		}
		if utils.Eq(value, float64(int(value))) {
			as.IntCount += count
		} else {
			as.RealCount += count
		}
		if as.NominalCounts != nil {
			as.NominalCounts[int(value)] = count
		}
		//if as.n Implement if necessary Stats for NumericStats
	}
	as.DistinctCount++
}
