package functions

import (
	"github.com/project-mac/src/data"
	//"fmt"
)

type NumericToBinary struct {
	input  data.Instances
	output data.Instances
}

func NewNumericToBinary() NumericToBinary {
	var ntb NumericToBinary
	return ntb
}

func (ntb *NumericToBinary) Exec(instances data.Instances) {
	ntb.SetInput(instances)
	tmp := make([]data.Instance,len(instances.Instances()))
	for i, instance := range instances.Instances() {
		tmp[i] = ntb.convertInstance(instance)
	}
	ntb.output.SetInstances(tmp)
}

func (ntb *NumericToBinary) convertInstance(instance data.Instance) data.Instance {
	inst := data.NewInstance()
	vals := make([]float64, len(instance.RealValues()))
	newIndexes := make([]int, len(instance.RealValues()))
	for j := range instance.RealValues() {
		att := ntb.input.Attribute(instance.Index(j))
		if att.Type() != data.NUMERIC || instance.Index(j) == ntb.input.ClassIndex() {
			//fmt.Println(ntb.input.ClassIndex())
			vals[j] = instance.ValueSparse(j)
		} else {
			if instance.IsMissingValue(j) {
				//fmt.Println("DSAD")
				vals[j] = instance.ValueSparse(j)
			} else {
				//fmt.Println("DSAD---")
				vals[j] = 1
			}
		}
		newIndexes[j] = instance.Index(j)
	}
	inst.SetWeight(instance.Weight())
	inst.SetRealValues(vals)
	inst.SetIndices(newIndexes)
	return inst
}

func (ntb *NumericToBinary) SetInput(data data.Instances) {
	ntb.input = data
	ntb.SetOutput()
}

func (ntb *NumericToBinary) SetOutput() {
	newAtts := make([]data.Attribute, 0)
	newClassIndex := ntb.input.ClassIndex()
	out := data.NewInstances()
	vals := make([]string, 2)

	// Compute new attributes
	for j, att := range ntb.input.Attributes() {
		if j == newClassIndex || att.Type() != data.NUMERIC {
			newAtts = append(newAtts,att)
		} else {
			attributeName := att.Name() + "_binarize"
			vals[0] = "0"
			vals[1] = "1"
			nAtt := data.NewAttribute()
			nAtt.SetName(attributeName)
			nAtt.SetType(data.NOMINAL)
			nAtt.SetValues(vals)
			nAtt.SetHasFixedBounds(true)
			hash := make(map[string]int, 2)
			hash[vals[0]] = 0
			hash[vals[1]] = 1
			nAtt.SetValuesIndexes(hash)
			newAtts = append(newAtts, nAtt)
		}
	}
	out.SetAttributes(newAtts)
	out.SetDatasetName(ntb.input.DatasetName())
	out.SetClassIndex(newClassIndex)
	ntb.output = out
}

func (ntb *NumericToBinary) Output() data.Instances {
	return ntb.output
}
