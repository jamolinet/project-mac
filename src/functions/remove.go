package functions

import (
	"github.com/project-mac/src/data"
)

type Remove struct {
	NotNil             bool
	SelectedCols       []int //the selected columns
	SelectedAttributes []int //the selected attributes' indexes, the ones we will keep
	InvertSel          bool  //whether to use invert selection or not, the selected attributes will be kept if true
	OutputFormat       data.Instances
}

func NewRemove() Remove {
	var r Remove
	r.InvertSel = true
	r.NotNil = true
	return r
}

func (r *Remove) IsNotNil() bool {
	return r.NotNil
}

// Start execution of the filter
func (r *Remove) Exec(instances data.Instances) {
	r.SetInputFormat(instances)
	for _, instance := range instances.Instances() {
		if r.OutputFormat.NumAttributes() == 0 {
			continue
		}
		vals := make([]float64, r.OutputFormat.NumAttributes())
		for i, current := range r.SelectedAttributes {
			vals[i] = instance.Value(current)
		}
		//Instance is always sparse
		inst := data.NewSparseInstance(instance.Weight(), vals, r.OutputFormat.Attributes())
		r.OutputFormat.Add(inst)
	}
}

// Sets the format of the input and output instances
func (r *Remove) SetInputFormat(instInfo data.Instances) {
	r.GetSelectedAttributes(len(instInfo.Attributes()))
	attributes := make([]data.Attribute, 0)
	outputClass := -1
	for _, current := range r.SelectedAttributes {
		if instInfo.ClassIndex() == current {
			outputClass = len(attributes)
		}
		keep := *instInfo.Attribute(current)
		//fmt.Println(keep.Name())
		attributes = append(attributes, keep)
	}
	//fmt.Println(len(attributes), "attributes", "\n", outputClass, "outputClass")
	r.OutputFormat = data.NewInstancesWithClassIndex(outputClass)
	r.OutputFormat.SetAttributes(attributes)
	r.OutputFormat.SetDatasetName(instInfo.DatasetName())
}

func (r *Remove) SetSelectedColumns(cols []int) {
	r.SelectedCols = cols
}

func (r *Remove) SetInvertSelection(flag bool) {
	r.InvertSel = flag
}

func (r *Remove) GetSelectedAttributes(numAttributes int) {
	if r.InvertSel {
		r.SelectedAttributes = r.SelectedCols
	} else {
		//a very costly implementation, **must be changed in the future**
		for j := range r.SelectedCols {
			for i := 0; i < numAttributes; i++ {
				if j != i {
					contains := func() bool {
						for _, k := range r.SelectedAttributes {
							if i == k {
								return true
							}
						}
						return false
					}
					if !contains() {
						r.SelectedAttributes = append(r.SelectedAttributes, i)
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
	if r.OutputFormat.NumAttributes() == 0 {
		//nothing is done
		return instance
	}
	vals := make([]float64, r.OutputFormat.NumAttributes())
	for i, current := range r.SelectedAttributes {
		vals[i] = instance.Value(current)
	}
	//Instance is always sparse
	inst := data.NewSparseInstance(instance.Weight(), vals, r.OutputFormat.Attributes())
	return inst
}
