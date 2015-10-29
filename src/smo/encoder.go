package smo

import (
	"encoding/json"
	"os"
	"github.com/pquerna/ffjson/ffjson"
)

func ToJSON(filePath, fileName string, class Classifier) {
	encodedFile, err := os.Create(filePath + "\\" + fileName)
	if err != nil {
		panic(err.Error())
	}
	defer encodedFile.Close()
	encoder := ffjson.NewEncoder(encodedFile)
	e := encoder.Encode(class)
	if  e != nil {
		panic(e.Error())
	}
}

func ToValueFromJSON(file string) Classifier {
	encodedFile, err := os.Open(file)
	if err != nil {
		panic(err.Error())
	}
	defer encodedFile.Close()
	decoder := json.NewDecoder(encodedFile)
	var smo Classifier
	decoder.Decode(&smo)
	return smo
}

func ToJSONSMO(filePath, fileName string, class SMO) {
	encodedFile, err := os.Create(filePath + "\\" + fileName)
	if err != nil {
		panic(err.Error())
	}
	defer encodedFile.Close()
	encoder := ffjson.NewEncoder(encodedFile)
	e := encoder.Encode(class)
	if  e != nil {
		panic(e.Error())
	}
}

func ToValueFromJSONSMO(file string) SMO {
	encodedFile, err := os.Open(file)
	if err != nil {
		panic(err.Error())
	}
	defer encodedFile.Close()
	decoder := json.NewDecoder(encodedFile)
	var smo SMO
	decoder.Decode(&smo)
	return smo
}