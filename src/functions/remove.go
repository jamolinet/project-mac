package functions

import (
	"github.com/project-mac/src/data"
)

type Remove struct {
	notNil             bool
	selectedCols       []int //the selected columns
	selectedAttributes []int //the selected attributes' indexes, the ones we will keep
	invertSel          bool  //whether to use invert selection or not, the selected attributes will be kept if true
	outputFormat       data.Instances
}

func NewRemove() Remove {
	var r Remove
	r.invertSel = true
	r.notNil = true
	return r
}

func (r *Remove) IsNotNil() bool {
	return r.notNil
}

// Start execution of the filter
func (r *Remove) Exec(instances data.Instances) {
	r.SetInputFormat(instances)
	for _, instance := range instances.Instances() {
		if r.outputFormat.NumAttributes() == 0 {
			continue
		}
		vals := make([]float64, r.outputFormat.NumAttributes())
		for i, current := range r.selectedAttributes {
			vals[i] = instance.Value(current)
		}
		//Instance is always sparse
		inst := data.NewSparseInstance(instance.Weight(), vals, r.outputFormat.Attributes())
		r.outputFormat.Add(inst)
	}
}

// Sets the format of the input and output instances
func (r *Remove) SetInputFormat(instInfo data.Instances) {
	r.getSelectedAttributes(len(instInfo.Attributes()))
	attributes := make([]data.Attribute, 0)
	outputClass := -1
	for _, current := range r.selectedAttributes {
		if instInfo.ClassIndex() == current {
			outputClass = len(attributes)
		}
		keep := *instInfo.Attribute(current)
		//fmt.Println(keep.Name())
		attributes = append(attributes, keep)
	}
	//fmt.Println(len(attributes), "attributes", "\n", outputClass, "outputClass")
	r.outputFormat = data.NewInstancesWithClassIndex(outputClass)
	r.outputFormat.SetAttributes(attributes)
	r.outputFormat.SetDatasetName(instInfo.DatasetName())
}

func (r *Remove) SetSelectedColumns(cols []int) {
	r.selectedCols = cols
}

func (r *Remove) SetInvertSelection(flag bool) {
	r.invertSel = flag
}

func (r *Remove) getSelectedAttributes(numAttributes int) {
	if r.invertSel {
		r.selectedAttributes = r.selectedCols
	} else {
		//a very costly implementation, **must be changed in the future**
		for j := range r.selectedCols {
			for i := 0; i < numAttributes; i++ {
				if j != i {
					contains := func() bool {
						for _, k := range r.selectedAttributes {
							if i == k {
								return true
							}
						}
						return false
					}
					if !contains() {
						r.selectedAttributes = append(r.selectedAttributes, i)
					}
				} else {
					break
				}
			}
		}
	}
}

// This method does the function of calling in weka convertInstance(Instance) and then output()
// due in this implementation does not exists the m_OutputQueue value
func (r *Remove) ConvertAndReturn(instance data.Instance) data.Instance {
	if r.outputFormat.NumAttributes() == 0 {
		//nothing is done
		return instance
	}
	vals := make([]float64, r.outputFormat.NumAttributes())
	for i, current := range r.selectedAttributes {
		vals[i] = instance.Value(current)
	}
	//Instance is always sparse
	inst := data.NewSparseInstance(instance.Weight(), vals, r.outputFormat.Attributes())
	return inst
}
