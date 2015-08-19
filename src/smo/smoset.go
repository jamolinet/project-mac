package smo

import ()

type SMOSet struct {
	//The current number of elements in the set
	//The first element in the set
	number, first int
	//Indicators
	indicators []bool
	//The next element for each element
	//The previous element for each element
	next, previous []int
}

// Creates a new set of the given size
func NewSMOSet(size int) SMOSet {
	var sset SMOSet
	sset.indicators = make([]bool, size)
	sset.next, sset.previous = make([]int, size), make([]int, size)
	sset.number = 0
	sset.first = -1
	return sset
}

// Checks whether an element is in the set
func (sset *SMOSet) Contains(index int) bool {
	return sset.indicators[index]
}

// Deletes an element from the set
func (sset *SMOSet) Delete(index int) {
	if sset.indicators[index] {
		if sset.first == index {
			sset.first = sset.next[index]
		} else {
			sset.next[sset.previous[index]] = sset.next[index]
		}
		if sset.next[index] != -1 {
			sset.previous[sset.next[index]] = sset.previous[index]
		}
		sset.indicators[index] = false
		sset.number--
	}
}

// Inserts an element into the set
func (sset *SMOSet) Insert(index int) {
	if !sset.indicators[index] {
		if sset.number == 0 {
			sset.first = index
			sset.next[index] = -1
			sset.previous[index] = -1
		} else {
			sset.previous[sset.first] = index
			sset.next[index] = sset.first
			sset.previous[index] = -1
			sset.first = index
		}
		sset.indicators[index] = true
		sset.number++
	}
}

// Gets the next element in the set. -1 gets the first one
func (sset *SMOSet) GetNext(index int) int {
	if index == -1 {
		return sset.first
	} else {
		return sset.next[index]
	}
}

// Returns the number of elements in the set
func (sset *SMOSet) NumElements() int {
	return sset.number
}
