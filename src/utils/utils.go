package utils

import (
	//	"fmt"
	"math"
)

const (
	SMALL = 1e-6
)

var index []int

//Greater than
func Gr(a float64, b float64) bool {
	return a-b > SMALL
}

//Smaller than
func Sm(a float64, b float64) bool {
	return b-a > SMALL
}

//Greater or Equal than
func GrOrEq(a float64, b float64) bool {
	return (b-a < SMALL) || (a >= b)
}

//Equal to
func Eq(a float64, b float64) bool {
	return (a == b) || (a-b < SMALL) && (b-a < SMALL)
}

// Sorts a given array of integers in ascending order and returns an array of
// integers with the positions of the elements of the original array in the
// sorted array. The sort is stable. (Equal elements remain in their original
//  order.)

func SortInt(array []int) []int {
	//Always reset global variable index
	index = nil

	index = make([]int, len(array))
	newIndex := make([]int, len(array))
	var helpIndex []int
	var numEqual int

	for i := range array {
		index[i] = i
	}
	quickSortINT(array, index, 0, len(array)-1)

	//Make sort stable
	i := 0
	for i < len(index) {
		numEqual = 1
		for j := i + 1; j < len(index) && (array[index[i]] == array[index[j]]); j++ {
			numEqual++
		}
		if numEqual > 1 {
			helpIndex = make([]int, numEqual)
			for j := range helpIndex {
				helpIndex[j] = i + j
			}
			quickSortINT(index, helpIndex, 0, numEqual-1)
			for j := 0; j < numEqual; j++ {
				newIndex[i+j] = index[helpIndex[j]]
			}
			i += numEqual
		} else {
			newIndex[i] = index[i]
			i++
		}
	}
	return newIndex
}

func quickSortINT(array, index []int, left, right int) {
	if left < right {
		middle := partitionINT(array, index, left, right)
		quickSortINT(array, index, left, middle)
		quickSortINT(array, index, middle+1, right)
	}
}

func partitionINT(array, index []int, l, r int) int {
	pivot := float64(array[index[(l+r)/2]])
	var help int
	for l < r {
		for array[index[l]] < int(pivot) && l < r {
			l++
		}
		for array[index[r]] > int(pivot) && l < r {
			r--
		}
		if l < r {
			help = index[l]
			index[l] = index[r]
			index[r] = help
			l++
			r--
		}
	}
	if l == r && array[index[r]] > int(pivot) {
		r--
	}
	return r
}

// Sorts a given array of doubles in ascending order and returns an array of
// integers with the positions of the elements of the original array in the
// sorted array. SEE: This sort is not stable anymore look for better implementations
func SortFloat(array []float64) []int {
	//Always reset global variable index
	index = nil
	initialIndex(len(array))
	//fmt.Println(len(index), "index")
	if len(array) > 1 {
		array = replaceMissingValueWithMaxFloat64(array)
		quickSort(array, 0, len(array)-1)
	}
	return index
}

// Initial index, filled with values from 0 to size - 1.
func initialIndex(size int) {
	index = make([]int, size)
	for i := range index {
		index[i] = i
	}
}

// Replaces all "missing values" in the given array of float64 values with math.MaxFloat64
func replaceMissingValueWithMaxFloat64(array []float64) []float64 {
	//fmt.Println(array, "array array")
	for i := range array {
		if math.IsNaN(array[i]) {
			array[i] = math.MaxFloat64
		}
	}
	return array
}

/**
 * Implements quicksort with median-of-three method and explicit sort for
 * problems of size three or less.

 */
func quickSort(array []float64 /*index []int,*/, left, right int) {
	diff := right - left
	switch diff {
	case 0:
		//do nothing
	case 1:
		// Swap two elements if necessary
		conditionalSwap(array, left, right)
	case 2:
		// Just need to sort three elements
		conditionalSwap(array, left, left+1)
		conditionalSwap(array, left, right)
		conditionalSwap(array, left+1, right)
	default:
		// Establish pivot
		var pivotLocation, center int
		pivotLocation = sortLeftRightAndCenter(array, left, right)
		//fmt.Println(pivotLocation)
		// Move pivot to the right, partition, and restore pivot
		swap(pivotLocation, right-1)
		center = partition(array, left, right, array[index[right-1]])
		//fmt.Println(center, "center")
		swap(center, right-1)
		// Sort recursively
		quickSort(array, left, center-1)
		quickSort(array, center+1, right)
	}
}

//Conditional swap for quick sort.
func conditionalSwap(array []float64 /*index *[]int,*/, left, right int) {
	if array[index[left]] > array[index[right]] {
		help := index[left]
		index[left] = index[right]
		index[right] = help
	}
}

//Sorts left, right, and center elements only, returns resulting center as
//pivot.
func sortLeftRightAndCenter(array []float64 /*index *[]int,*/, l, r int) int {
	c := (l + r) / 2
	conditionalSwap(array, l, c)
	conditionalSwap(array, l, r)
	conditionalSwap(array, c, r)
	return c
}

//Swaps two elements in the given integer array.
func swap( /*index *[]int,*/ l, r int) {
	help := index[l]
	index[l] = index[r]
	index[r] = help
}

//Partitions the instances around a pivot. Used by quicksort and
//kthSmallestValue.
func partition(array []float64 /*index *[]int,*/, l, r int, pivot float64) int {
	r--
	for true {
		{
			l++
			for array[index[l]] < pivot {
				l++
			}
		}
		{
			r--
			for array[index[r]] > pivot {
				r--
			}
		}
		if l >= r {
			return l
		}
		swap(l, r)
	}
	return 0
}

func MaxIndex(floats []float64) int {
	maximmun, maxIndex := 0.0, 0

	for i, d := range floats {
		if i == 0 || d > maximmun {
			maxIndex = i
			maximmun = d
		}
	}
	return maxIndex
}

func MaxIndexInts(floats []int) int {
	maximmun, maxIndex := 0, 0

	for i, d := range floats {
		if i == 0 || d > maximmun {
			maxIndex = i
			maximmun = d
		}
	}
	return maxIndex
}

func Normalize(doubles *[]float64, sum float64) {
	d := *doubles
	if math.IsNaN(sum) {
		panic("Can't normalize array. Sum is NaN.")
	}
	if sum == 0 {
		panic("Can't normalize array. Sum is zero.")
	}
	for i:= range d {
		d[i] /= sum
	}
	*doubles = d
}

