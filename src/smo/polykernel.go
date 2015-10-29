package smo

import (
	"fmt"
	"github.com/project-mac/src/data"
	"math"
)

type PolyKernel struct {
	/* From PolyKernel.java */
	LowerOrder bool
	Exponent   float64
	/* From CachedKernel.java */
	KernelEvals, CacheHints, CacheSize_ int
	Storage                            []float64
	Keys                               []int64

	KernelMatrix         [][]float64
	NumInsts, CacheSlots int
	/* From Kernel.java */
	Data            data.Instances
	ChecksTurnedOff bool
}

// Default constructor
func NewPolyKernel() PolyKernel {
	var pk PolyKernel
	pk.Exponent = 1
	pk.LowerOrder = false
	pk.CacheSize_ = 250007
	pk.CacheSlots = 4
	pk.ChecksTurnedOff = false
	return pk
}

// New instance of PolyKernel
func NewPolyKernelWithParams(data data.Instances, CacheSize_ int, Exponent float64, LowerOrder bool) PolyKernel {
	var pk PolyKernel
	pk.ChecksTurnedOff = false
	pk.CacheSlots = 4
	pk.SetCacheSize(CacheSize_)
	pk.SetExponent(Exponent)
	pk.SetLowerOrder(LowerOrder)
	pk.BuildKernel(data)
	return pk
}

// Frees the cache used by the kernel
func (pk *PolyKernel) Clean() {
	if pk.Exponent == 1 {
		pk.Data = data.NewInstances()
		pk.Storage = nil
		pk.Keys = nil
		pk.KernelMatrix = nil
	}
}

// Builds the kernel with the given Data. Initializes the kernel cache.
// The actual size of the cache in bytes is (64 * CacheSize_).
func (pk *PolyKernel) BuildKernel(Data data.Instances) {
	if !pk.ChecksTurnedOff {
	}
	pk.InitVars(Data)
}

// Initializes variables
func (pk *PolyKernel) InitVars(Data data.Instances) {
	pk.Data = Data

	pk.KernelEvals = 0
	pk.CacheHints = 0
	pk.NumInsts = len(Data.Instances())

	if pk.CacheSize_ > 0 {
		//Use LRU Cache
		pk.Storage = make([]float64, pk.CacheSize_*pk.CacheSlots)
		pk.Keys = make([]int64, pk.CacheSize_*pk.CacheSlots)
	} else {
		pk.Storage = nil
		pk.KernelMatrix = nil
		pk.Keys = nil
	}
}

// Implements the abstract function of Kernel using the cache. This method
// uses the Evaluate() method to do the actual dot product.
func (pk *PolyKernel) Eval(id1, id2 int, inst1 data.Instance) float64 {

	result := 0.0
	key := int64(-1)
	location := -1
	// we can only cache if we know the indexes and caching is not
	// disbled (m_cacheSize == -1)
	if id1 >= 0 && pk.CacheSize_ != -1 {
		//Use full cache?
		if pk.CacheSize_ == 0 {
			if pk.KernelMatrix == nil {
				pk.KernelMatrix = make([][]float64, len(pk.Data.Instances()))
				for i := range pk.KernelMatrix {
					pk.KernelMatrix[i] = make([]float64, i+1)
					for j := 0; j <= 1; j++ {
						pk.KernelEvals++
						pk.KernelMatrix[i][j] = pk.Evaluate(i, j, *pk.Data.Instance(i))
					}
				}
			}
			pk.CacheHints++
			if id1 > id2 {
				result = pk.KernelMatrix[id1][id2]
			} else {
				result = pk.KernelMatrix[id2][id1]
			}
			return result
		}
		// Use LRU cache
		if id1 > id2 {
			key = int64(id1) + int64(id2*pk.NumInsts)
		} else {
			key = int64(id2) + int64(id1*pk.NumInsts)
		}
		location = int(key % int64(pk.CacheSize_) * int64(pk.CacheSlots))
		loc := location
		for i := 0; i < pk.CacheSlots; i++ {
			thiskey := pk.Keys[loc]
			if thiskey == 0 {
				break // empty slot, so break out of loop early
			}
			if thiskey == (key + 1) {
				pk.CacheHints++
				// move entry to front of cache (LRU) by swapping
				// only if it's not already at the front of cache
				if i > 0 {
					tmps := pk.Storage[loc]
					pk.Storage[loc] = pk.Storage[location]
					pk.Keys[loc] = pk.Keys[location]
					pk.Storage[location] = tmps
					pk.Keys[location] = thiskey
					return tmps
				} else {
					return pk.Storage[loc]
				}
			}
			loc++
		}
	}
	result = pk.Evaluate(id1, id2, inst1)

	pk.KernelEvals++
	// store result in cache
	if key != -1 && pk.CacheSize_ != -1 {
		// move all cache slots forward one array index
		// to make room for the new entry
		tmpKeys := pk.Keys[location:location+pk.CacheSlots]
		tmpStorage := pk.Storage[location:location+pk.CacheSlots]
			for i := 1; i <= pk.CacheSlots; i++ {
				pk.Keys[location+i] = tmpKeys[i-1]
				pk.Storage[location+i] = tmpStorage[i-1]
			}
		//		copy(tmpKeys[location+1:], pk.Keys[location:pk.CacheSlots])
		//		copy(tmpStorage[location+1:], pk.Storage[location:pk.CacheSlots])
		//copy(pk.Keys[location+1:], pk.Keys[location:pk.CacheSlots]) //Later see if is ok if not then add +1
		//copy(pk.Storage[location+1:], pk.Storage[location:pk.CacheSlots])//Later see if is ok if not then add +1
		pk.Storage[location] = result
		pk.Keys[location] = key + 1
	}
	return result
}

// Evaluates the kernel
func (pk *PolyKernel) Evaluate(id1, id2 int, inst1 data.Instance) float64 {
	var result float64
	if id1 == id2 {
		result = pk.DotProd(inst1, inst1)
	} else {
		result = pk.DotProd(inst1, *pk.Data.Instance(id2))
	}
	//Use lower order terms?
	if pk.LowerOrder {
		result += 1
	}
	if pk.Exponent != 1 {
		result = math.Pow(result, pk.Exponent)
	}
	return result
}

// Calculates a dot product between two instances
func (pk *PolyKernel) DotProd(inst1, inst2 data.Instance) float64 {
	result := 0.0

	// A fast dot product
	n1 := len(inst1.RealValues())
	n2 := len(inst2.RealValues())
	classIndex := pk.Data.ClassIndex()
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
		pk.CacheSize_ = value
		pk.Clean()
	} else {
		fmt.Printf("Cache size cannot be less than -1, provided: %i", value)
	}
}

// Returns the number of time Eval has been called.
func (pk *PolyKernel) NumEvals() int {
	return pk.KernelEvals
}

// Returns the number of cache hits on dot products.
func (pk *PolyKernel) NumCacheHints() int {
	return pk.CacheHints
}

// Gets the size of the cache
func (pk *PolyKernel) CacheSize() int {
	return pk.CacheSize_
}

func (pk *PolyKernel) SetLowerOrder(set bool) {
	pk.LowerOrder = set
}

func (pk *PolyKernel) SetExponent(set float64) {
	pk.Exponent = set
}

func (pk *PolyKernel) SetChecksTurnedOff(set bool) {
	pk.ChecksTurnedOff = set
}
