package main

import (
	"fmt"
	"github.com/project-mac/src/data"
	"github.com/project-mac/src/functions"
	"github.com/project-mac/src/smo"
	"github.com/project-mac/src/utils"
	//	"github.com/project-mac/src/utils"
	"math/rand"
	"math"
	//"time"
)

func main() {

	instances := data.NewInstancesWithClassIndex(0)
	err := instances.ParseFile("C:\\Users\\Yuri\\Documents\\Go\\src\\github.com\\project-mac\\src\\main\\AppsLemmas.arff")
	if err != nil {
		panic(err.Error())
	}
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
	classifier := smo.NewClassifier(&_smo)
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
	ig.SetBinarize(false)
	as := functions.NewAttributeSelection()
	as.SetEvaluator(ig)
	as.SetSearchMethod(ranker)
	classifier.SetStringToWordVector(vector)
	classifier.SetAS(as)
	/*class := */ evaluator.CrossValidateModel(classifier, instances,5, random)
	fmt.Println(evaluator.ToSummaryString("\nResults\n======\n", false))
	fmt.Println(evaluator.ToClassDetailsString("=== Detailed Accuracy By Class ===\n"))
	fmt.Println(evaluator.ToMatrixString("=== Confusion Matrix ===\n"))
	fmt.Println("IS DONE!!!!!!!!")

	//loadSMO(_smo)
}

func loadSMO(_smo smo.SMO) {
	smo.ToJSONSMO("D:\\", "enc.txt", _smo)
	insts1 := data.NewInstancesWithClassIndex(0)
	insts1.ParseFile("C:\\Users\\Yuri\\Documents\\Go\\src\\github.com\\project-mac\\src\\main\\n1.arff")
	vector1 := functions.NewStringToWordVectorInst(insts1)
	vector1.SetIDF_Transformation(false)
	vector1.SetTF_Transformation(true)
	vector1.SetWordsToKeep(15)
	vector1.SetPerClass(false)
	vector1.SetNormalize(true)
	insts := vector1.Exec()
	ig := functions.NewInfoGain()
	ranker := functions.NewRanker()
	ranker.SetThreshold(0.0)
	ranker.SetNumToSelect(-1)
	ig.SetBinarize(false)
	as := functions.NewAttributeSelection()
	as.SetEvaluator(ig)
	as.SetSearchMethod(ranker)
	as.StartSelection(insts)
	insts = as.Output()
	fmt.Println(insts.Attributes_)
	model := smo.ToValueFromJSONSMO("D:\\enc.txt")
	eval := smo.NewEvaluation(insts1)
	for i, instance := range insts.Instances_ {
		dist := model.DistributionForInstance(instance, insts.NumClasses())
				fmt.Println(utils.MaxIndex(dist))
		//		fmt.Println(insts.Attributes_[0].Value(index))
		eval.UpdateStatsForClassfier(dist, insts1.InstanceNoPtr(i),&insts1)
	}
	fmt.Println(eval.ToSummaryString("\nResults\n======\n", false))
	fmt.Println(eval.ToMatrixString("=== Confusion Matrix ===\n"))
	fmt.Println(math.Lgamma(5))

}
