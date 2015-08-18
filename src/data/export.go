package data

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	//	"syscall"
)

func ExportToArffFileSparse(data Instances, fileName string, filePath string) {
	//Select the necessary information
	name := data.DatasetName()
	attributes := data.Attributes()
	instances := data.Instances()
	//If fileName is undefined so give a particular name to the file
	if fileName == "" {
		fileName = strings.Replace(strings.Replace(name+"_"+strconv.FormatInt(rand.Int63(), 10)+"_"+time.Now().String(), "-", "_", -1), ":", "_", -1)
	}
	check := func(err error) {
		if err != nil {
			panic(fmt.Errorf("The file with name: %s, in path: %s cannot be created! \n Error throw: %s", fileName, filePath, err.Error()))
		}
	}
	var file *os.File
	var err error
	if filePath == "" {
		file, err = os.Create(fileName + ".arff")
		check(err)
	} else {
		file, err = os.Create(filePath + fileName + ".arff")
		check(err)
	}

	writer := bufio.NewWriter(file)
	defer file.Close()
	fmt.Println(file.Name())
	writer.WriteString("@relation " + name)
	writer.WriteString("\n\n")
	attrsDefinition := ""
	for _, attr := range attributes {
		attrsDefinition += "@attribute "
		attrsDefinition += attr.Name() + " "
		if attr.Type() == NUMERIC {
			attrsDefinition += "numeric "
			if attr.HasFixedBounds() {
				attrsDefinition += fmt.Sprint("[", attr.Min(), ",", attr.Max())
			}
		} else if attr.Type() == NOMINAL {
			tmp := strings.Replace(fmt.Sprint(attr.Values()), "[", "{", 1)
			tmp = strings.Replace(tmp, " ", ",", -1)
			attrsDefinition += fmt.Sprint(strings.Replace(tmp, "]", "}", 1))
		}
		attrsDefinition += "\n"
	}
	instDeclaration := ""
	for _, inst := range instances {
		tmp := "{"
		for i, val := range inst.RealValues() {
			idx := inst.Indices()[i]
			if attributes[idx].Type() == NOMINAL {
				tmp += fmt.Sprint(idx, " ", attributes[idx].Values()[int(val)], ",")
			} else if attributes[idx].Type() == NUMERIC {
				tmp += fmt.Sprint(idx, val, ",")
			}
		}
		tmp = strings.TrimSuffix(tmp, ",")
		tmp += "}"
		instDeclaration += "\n" + tmp
	}
	writer.WriteString(attrsDefinition)
	writer.WriteString("\n@data")
	writer.WriteString(instDeclaration)
	writer.Flush()
}
