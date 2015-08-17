package main

import (
	"fmt"
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/functions"
)

func main() {
	instances := data.NewInstancesWithClassIndex(0)
	instances.ParseFile("C:\\Users\\Yuri\\workspace\\SMO\\src\\main\\_AppsLemmas.arff")
//	for _, attr := range instances.Attributes() {
//		fmt.Println(attr.Name(), attr.Type())
//		for idx, val := range attr.Values() {
//			fmt.Printf("idx: %d val: %s", idx, val)
//			fmt.Println()
//		}
//	}
//
//	for _, inst := range instances.Instances() {
//		fmt.Println(inst.Values())
//		fmt.Println(inst.RealValues())
//	}
	stwv := functions.NewStringToWordVectorInst(instances)
	stwv.SetIDF_Transformation(true)
	stwv.SetTF_Transformation(true)
	stwv.SetWordsToKeep(15)
	stwv.SetPerClass(false)
	stwv.SetNormalize(false)
	processed := stwv.Exec()
	//processed = stwv.Exec()
	processed.ClassIndex()
//	for _, attr := range processed.Attributes() {
//		fmt.Println(attr.Name(), attr.Type())
//		for idx, val := range attr.Values() {
//			fmt.Printf("idx: %d val: %s", idx, val)
//			fmt.Println()
//		}
//	}
	//fmt.Println(processed.Attributes()[0].Values())
	for _, inst := range processed.Instances() {
		fmt.Println(inst.Values())
		fmt.Println(inst.RealValues())
		fmt.Println(inst.Indices())
	}
}
