package smo

import (
	"fmt"
	"github.com/project-mac/src/data"
	"math"
)

type PolyKernel struct {
	/* From PolyKernel.java */
	lowerOrder bool
	exponent   float64
	/* From CachedKernel.java */
	kernelEvals, cacheHints, cacheSize int
	storage                            []float64
	keys                               []int64

	kernelMatrix         [][]float64
	numInsts, cacheSlots int
	/* From Kernel.java */
	data            data.Instances
	checksTurnedOff bool
}

// Default constructor
func NewPolyKernel() PolyKernel {
	var pk PolyKernel
	pk.exponent = 1
	pk.lowerOrder = false
	pk.cacheSize = 250007
	pk.cacheSlots = 4
	pk.checksTurnedOff = false
	return pk
}

// New instance of PolyKernel
func NewPolyKernelWithParams(data data.Instances, cacheSize int, exponent float64, lowerOrder bool) PolyKernel {
	var pk PolyKernel
	pk.checksTurnedOff = false
	pk.cacheSlots = 4
	pk.SetCacheSize(cacheSize)
	pk.SetExponent(exponent)
	pk.SetLowerOrder(lowerOrder)
	pk.BuildKernel(data)
	return pk
}

// Frees the cache used by the kernel
func (pk *PolyKernel) Clean() {
	if pk.exponent == 1 {
		pk.data = data.NewInstances()
		pk.storage = nil
		pk.keys = nil
		pk.kernelMatrix = nil
	}
}

// Builds the kernel with the given data. Initializes the kernel cache.
// The actual size of the cache in bytes is (64 * cacheSize).
func (pk *PolyKernel) BuildKernel(data data.Instances) {
	if !pk.checksTurnedOff {
	}
	pk.initVars(data)
}

// Initializes variables
func (pk *PolyKernel) initVars(data data.Instances) {
	pk.data = data

	pk.kernelEvals = 0
	pk.cacheHints = 0
	pk.numInsts = len(data.Instances())

	if pk.cacheSize > 0 {
		//Use LRU Cache
		pk.storage = make([]float64, pk.cacheSize*pk.cacheSlots)
		pk.keys = make([]int64, pk.cacheSize*pk.cacheSlots)
	} else {
		pk.storage = nil
		pk.kernelMatrix = nil
		pk.keys = nil
	}
}

// Implements the abstract function of Kernel using the cache. This method
// uses the Evaluate() method to do the actual dot product.
func (pk *PolyKernel) Eval(id1, id2 int, inst1 data.Instance) float64 {
	//fmt.Println(id1, "id1", id2,"id2")
	result := 0.0
	key := int64(-1)
	location := -1
	// we can only cache if we know the indexes and caching is not
	// disbled (m_cacheSize == -1)
	if id1 >= 0 && pk.cacheSize != -1 {
		//Use full cache?
		if pk.cacheSize == 0 {
			if pk.kernelMatrix == nil {
				pk.kernelMatrix = make([][]float64, len(pk.data.Instances()))
				for i := range pk.kernelMatrix {
					pk.kernelMatrix[i] = make([]float64, i+1)
					for j := 0; j <= 1; j++ {
						pk.kernelEvals++
						pk.kernelMatrix[i][j] = pk.Evaluate(i, j, *pk.data.Instance(i))
					}
				}
			}
			pk.cacheHints++
			if id1 > id2 {
				result = pk.kernelMatrix[id1][id2]
			} else {
				result = pk.kernelMatrix[id2][id1]
			}
			return result
		}
		// Use LRU cache
		if id1 > id2 {
			key = int64(id1) + int64(id2*pk.numInsts)
		} else {
			key = int64(id2) + int64(id1*pk.numInsts)
		}
		location = int(key % int64(pk.cacheSize) * int64(pk.cacheSlots))
		loc := location
		for i := 0; i < pk.cacheSlots; i++ {
			thiskey := pk.keys[loc]
			if thiskey == 0 {
				break // empty slot, so break out of loop early
			}
			if thiskey == (key + 1) {
				pk.cacheHints++
				// move entry to front of cache (LRU) by swapping
				// only if it's not already at the front of cache
				if i > 0 {
					tmps := pk.storage[loc]
					pk.storage[loc] = pk.storage[location]
					pk.keys[loc] = pk.keys[location]
					pk.storage[location] = tmps
					pk.keys[location] = thiskey
					return tmps
				} else {
					return pk.storage[loc]
				}
			}
			loc++
		}
	}
	result = pk.Evaluate(id1, id2, inst1)

	pk.kernelEvals++
	// store result in cache
	if key != -1 && pk.cacheSize != -1 {
		// move all cache slots forward one array index
		// to make room for the new entry
		tmpKeys := pk.keys[location:location+pk.cacheSlots]
		tmpStorage := pk.storage[location:location+pk.cacheSlots]
			for i := 1; i <= pk.cacheSlots; i++ {
//				fmt.Println(i, location, location+i, i-1)
				pk.keys[location+i] = tmpKeys[i-1]
				pk.storage[location+i] = tmpStorage[i-1]
			}
		//		copy(tmpKeys[location+1:], pk.keys[location:pk.cacheSlots])
		//		copy(tmpStorage[location+1:], pk.storage[location:pk.cacheSlots])
		//copy(pk.keys[location+1:], pk.keys[location:pk.cacheSlots]) //Later see if is ok if not then add +1
		//copy(pk.storage[location+1:], pk.storage[location:pk.cacheSlots])//Later see if is ok if not then add +1
		pk.storage[location] = result
		pk.keys[location] = key + 1
		//		pk.storage= tmpStorage
		//		pk.keys = tmpKeys
	}
	return result
}

// Evaluates the kernel
func (pk *PolyKernel) Evaluate(id1, id2 int, inst1 data.Instance) float64 {
	var result float64
	if id1 == id2 {
		result = pk.dotProd(inst1, inst1)
	} else {
		//fmt.Println(id2, "id2")
		//fmt.Println(len(pk.data.Instances()))
		result = pk.dotProd(inst1, *pk.data.Instance(id2))
	}
	//Use lower order terms?
	if pk.lowerOrder {
		result += 1
	}
	if pk.exponent != 1 {
		result = math.Pow(result, pk.exponent)
	}
	return result
}

// Calculates a dot product between two instances
func (pk *PolyKernel) dotProd(inst1, inst2 data.Instance) float64 {
	result := 0.0

	// A fast dot product
	n1 := len(inst1.RealValues())
	n2 := len(inst2.RealValues())
	classIndex := pk.data.ClassIndex()
	for p1, p2 := 0, 0; p1 < n1 && p2 < n2; {
		ind1 := inst1.Index(p1)
		ind2 := inst2.Index(p2)
		if ind1 == ind2 {
			if ind1 != classIndex {
				result += inst1.ValueSparse(p1) * inst2.ValueSparse(p2)
			}
			p1++
			p2++
		} else if ind1 > ind2 {
			p2++
		} else {
			p1++
		}
	}
	return result
}

// Sets the size of the cache to use (a prime number)
func (pk *PolyKernel) SetCacheSize(value int) {
	if value >= -1 {
		pk.cacheSize = value
		pk.Clean()
	} else {
		fmt.Printf("Cache size cannot be less than -1, provided: %i", value)
	}
}

// Returns the number of time Eval has been called.
func (pk *PolyKernel) NumEvals() int {
	return pk.kernelEvals
}

// Returns the number of cache hits on dot products.
func (pk *PolyKernel) NumCacheHints() int {
	return pk.cacheHints
}

// Gets the size of the cache
func (pk *PolyKernel) CacheSize() int {
	return pk.cacheSize
}

func (pk *PolyKernel) SetLowerOrder(set bool) {
	pk.lowerOrder = set
}

func (pk *PolyKernel) SetExponent(set float64) {
	pk.exponent = set
}

func (pk *PolyKernel) SetChecksTurnedOff(set bool) {
	pk.checksTurnedOff = set
}
