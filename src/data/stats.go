package data

import (
	"math"
)

type Stats struct {
	/*
		The number of values seen
		The sum of values seen
		The sum of values squared seen
		The std deviation of values at the last calculateDerived() call
		The mean of values at the last calculateDerived() call
		The minimum value seen, or Double.NaN if no values seen
		The maximum value seen, or Double.NaN if no values seen
	*/
	Count, Sum, SumSq, StdDev, Mean,
	Min, Max float64
}

func NewStats() Stats {
	var s Stats
	s.Count, s.Sum, s.SumSq = 0, 0, 0
	s.StdDev, s.Mean, s.Min, s.Max = math.NaN(), math.NaN(), math.NaN(), math.NaN()
	return s
}

// Adds a value to the observed values
func (this Stats) Add(value float64) {
	this.AddNtimes(value, 1)
}

// Adds a value that has been seen n times to the observed values
func (this Stats) AddNtimes(value, n float64) {
	this.Sum += value * n
	this.SumSq += value * value * n
	this.Count += n
	if math.IsNaN(this.Min) {
		this.Min, this.Max = value, value
	} else if value < this.Min {
		this.Min = value
	} else if value > this.Max {
		this.Max = value
	}
}

// Removes a value to the observed values (no checking is done
// that the value being removed was actually added)
func (this Stats) Substract(value float64) {
	this.SubstractNtimes(value, 1)
}

// Subtracts a value that has been seen n times from the observed values
func (this Stats) SubstractNtimes(value, n float64) {
	this.Sum -= value * n
	this.SumSq -= value * value * n
	this.Count -= n
}

// Tells the object to calculate any statistics that don't have their
// values automatically updated during add. Currently updates the mean
// and standard deviation.
func (this Stats) CalculateDerived() {
	this.Mean = math.NaN()
	this.StdDev = math.NaN()
	if this.Count > 0 {
		this.Mean = this.Sum / this.Count
		this.StdDev = math.Inf(1)
		if this.Count > 1 {
			this.StdDev = this.SumSq - (this.Sum*this.Sum)/this.Count
			this.StdDev /= (this.Count - 1)
			if this.StdDev < 0 {
				this.StdDev = 0
			}
			this.StdDev = math.Sqrt(this.StdDev)
		}
	}
}
