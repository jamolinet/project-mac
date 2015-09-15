package main

import (
	"fmt"
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/functions"
	"github.com/project-mac/src/smo"
	//	"github.com/project-mac/src/utils"
	"math/rand"
	//"time"
)

func main() {
	//now := time.Now()
	instances := data.NewInstancesWithClassIndex(0)
	err := instances.ParseFile("C:\\Users\\Yuri\\Documents\\Go\\src\\github.com\\project-mac\\src\\main\\_AppsLemmas.arff")
	if err != nil {
		panic(err.Error())
	}
//	for _,i:=range instances.Instances() {
//			fmt.Println(i.RealValues(), "instance.RealValues() main_smo")	
//		}
	//for _, attr := range instances.Attributes() {
		//fmt.Println(attr.Name(), attr.Type())
		//		for idx, val := range attr.Values() {
		//			fmt.Printf("idx: %d val: %s", idx, val)
		//			fmt.Println()
		//		}
	//}

	//		for _, inst := range instances.Instances() {
	//			fmt.Println(inst.Values())
	//			fmt.Println(inst.RealValues())
	//		}
//	stwv := functions.NewStringToWordVectorInst(instances)
//	stwv.SetIDF_Transformation(true)
//	stwv.SetTF_Transformation(true)
//	stwv.SetWordsToKeep(15)
//	stwv.SetPerClass(false)
//	stwv.SetNormalize(false)
//	processed := stwv.Exec()
	//processed = stwv.Exec()
	//processed.ClassIndex()
	//	for _, attr := range processed.Attributes() {
	//		fmt.Println(attr.Name(), attr.Type())
	//		for idx, val := range attr.Values() {
	//			fmt.Printf("idx: %d val: %s", idx, val)
	//			fmt.Println()
	//		}
	//	}
	//fmt.Println(processed.Attributes()[0].Values())
	//	for _, inst := range processed.Instances() {
	//		fmt.Println(inst.Values())
	//		fmt.Println(inst.RealValues())
	//		fmt.Println(inst.Indices())
	//	}
//	ig := functions.NewInfoGain()
//	ranker := functions.NewRanker()
//	ranker.SetThreshold(0.0)
//	ranker.SetNumToSelect(-1)
//	//ranker.SetRange("1,3")
//	ig.SetBinarize(true)
//	as := functions.NewAttributeSelection()
//	as.SetEvaluator(ig)
//	as.SetSearchMethod(ranker)
//	as.StartSelection(processed)
//	processed = as.Output()
//	for _, inst := range processed.Instances() {
//		fmt.Println(inst.Values())
//		fmt.Println(inst.RealValues())
//		fmt.Println(inst.Indices())
//	}
	//ig.BuildEvaluator(processed)
	//i := utils.SortFloat([]float64{0, 0, 0, 0, 0, 0, 0.8904916402194916, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.7793498372920848, 0, 0, 0, 0, 0.7793498372920848})
	//fmt.Println(i)
	//data.ExportToArffFileSparse(processed, "", "")
	//now1 := time.Now()
	//fmt.Println("Total execution time: ", now1.Sub(now).Seconds(), "second(s)")
	evaluator := smo.NewEvaluation(instances)
//	if &evaluator == nil {
//		println("nothing")
//	}
//	for _,p :=range processed.Instances() {
//		fmt.Println(p.Weight())
//	}
	random := rand.New(rand.NewSource(1))
	kernel := smo.NewPolyKernel()
	kernel.SetExponent(1)
	kernel.SetCacheSize(250007)
	_smo := smo.NewSMO(kernel)
	_smo.SetC(1)
	_smo.SetTolerance(0.0010)
	_smo.SetEps(1.0E-12)
	_smo.SetNormalize(true)
	_smo.SetNumFolds(1)
	_smo.SetSeed(1)
	classifier := smo.NewClassifier(_smo)
	vector := functions.NewStringToWordVector()
	vector.SetIDF_Transformation(false)
	vector.SetTF_Transformation(true)
	vector.SetWordsToKeep(15)
	vector.SetPerClass(false)
	vector.SetNormalize(true)
	ig := functions.NewInfoGain()
	ranker := functions.NewRanker()
	ranker.SetThreshold(0.0)
	ranker.SetNumToSelect(-1)
	//ranker.SetRange("1,3")
	ig.SetBinarize(true)
	as := functions.NewAttributeSelection()
	as.SetEvaluator(ig)
	as.SetSearchMethod(ranker)
	classifier.SetStringToWordVector(vector)
	classifier.SetAS(as)
	evaluator.CrossValidateModel(classifier, instances,3, random)
	fmt.Println(evaluator.ToSummaryString("\nResults\n======\n",false))
	fmt.Println(evaluator.ToMatrixString("=== Confusion Matrix ===\n"))
	fmt.Println("Is done!!!!!!!!")
}
