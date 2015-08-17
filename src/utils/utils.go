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

//Sorts a given array of doubles in ascending order and returns an array of
//integers with the positions of the elements of the original array in the
//sorted array. SEE: This sort is not stable anymore look for better implementations
func SortFloat(array []float64) []int {
	initialIndex(len(array))
	//fmt.Println(len(index), "index")
	if len(array) > 1 {
		array = replaceMissingValueWithMaxFloat64(array)
		quickSort(array, 0, len(array)-1)
	}
	return index
}

//Initial index, filled with values from 0 to size - 1.
func initialIndex(size int) {
	index = make([]int, size)
	for i := range index {
		index[i] = i
	}
}

//Replaces all "missing values" in the given array of float64 values with math.MaxFloat64
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
 *
 * @param array the array of doubles to be sorted
 * @param index the index into the array of doubles
 * @param left the first index of the subset to be sorted
 * @param right the last index of the subset to be sorted
 */
// @ requires 0 <= first && first <= right && right < array.length;
// @ requires (\forall int i; 0 <= i && i < index.length; 0 <= index[i] &&
// index[i] < array.length);
// @ requires array != index;
// assignable index;
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
