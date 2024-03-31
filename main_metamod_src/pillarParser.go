package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func writePillarScriptedEffect(inDir string, outDir string, outName string) {
	innoLib := readPillarFiles(inDir)

	outFile, err := os.Create(filepath.Join(outDir, outName))
	if err != nil {
		fmt.Println("Failed to create new file: " + filepath.Join(outDir, outName))
		time.Sleep(100)
		log.Fatal(err)
	}
	writeHeader(outFile)
	_, err = outFile.WriteString("demd_add_pillar_bonuses = {\n\n")
	for _, inno := range innoLib {
		if len(inno.bonuses) > 0 {
			_, err = outFile.WriteString("\tif = {\n")
			_, err = outFile.WriteString("\t\tlimit = { has_cultural_pillar = " + inno.name + " }\n")
			for _, bonus := range inno.bonuses {
				_, err = outFile.WriteString("\t\tchange_variable = { name = " + bonus.resource +  " add = " + fmt.Sprintf("%.2f", bonus.amount) + " }\n")
			}

			_, err = outFile.WriteString("\t}\n\n")
		}
	}
	_, err = outFile.WriteString("}")
}

func readPillarFiles(inDir string) []inno {
	fileInfo, err := ioutil.ReadDir(inDir)
	if err != nil {
		fmt.Println("failed to read directory: " + inDir)
		log.Fatal(err)
	}

	innoLib := make([]inno,0)

	for i := 0; i < len(fileInfo); i++ {
		innos := readPillarFile(filepath.Join(inDir,fileInfo[i].Name()))
		for _, inno := range innos {
			innoLib = append(innoLib, inno)
		}
	}
	return innoLib
}

// Return instance file contents
func readPillarFile(inFile string) map[string]inno {
	thisFile, err := os.Open(inFile)
	if err != nil {
		fmt.Println("Failed to open file: " + inFile)
		log.Fatal(err)
	}
	// Initialize scanner
	scanner := bufio.NewScanner(thisFile)

	innovations := make(map[string]inno)

	bracketCounter := 0
	currentInnoName := "placeholder"

	for scanner.Scan() {
		line := cleanLine(scanner.Text())

		fields := strings.Fields(line)

		// note entrance into new innovation line
		if len(fields) > 0 && bracketCounter == 0 &&
			(strings.Split(fields[0], "_")[0] == "martial" || strings.Split(fields[0], "_")[0] == "ethos" ||
				strings.Split(fields[0], "_")[0] == "heritage" || strings.Split(fields[0], "_")[0] == "language") {
			currentInnoName = fields[0]
		}

		// note relevant bonus
		if len(fields) > 2 && strings.Contains(fields[0],"demd_meta") {
			fragments := strings.Split(fields[0], "_")
			length := len(fragments)

			// extract value
			value, _ := strconv.ParseFloat(fragments[length-1], 64)


			// extract resource resource
			resource := fragments[2]
			for i := 3; i < length-2; i++ {
				resource += "_" + fragments[i]
			}

			if fragments[length-2] == "mult" {
				value = value / 100.0
				resource += "_cultural_production_mult"
			} else if fragments[length-2] == "multneg" {
				value = - value / 100.0
				resource += "_cultural_production_mult"
			}
			var newBonus bonus
			newBonus.resource = resource
			newBonus.amount = value

			if _, ok := innovations[currentInnoName]; ok {
				// inno exists in map
				innovations[currentInnoName].bonuses[newBonus.resource] = newBonus
			} else {
				// inno not yet in map
				var newInno inno
				newInno.name = currentInnoName
				newInno.bonuses = make(map[string]bonus)
				innovations[currentInnoName] = newInno
				innovations[currentInnoName].bonuses[newBonus.resource] = newBonus
			}

		}

		// note bracket depth
		negBrackets := strings.Count(line, "}")
		bracketCounter -= negBrackets
		posBrackets := strings.Count(line, "{")
		bracketCounter += posBrackets
	}
	thisFile.Close()
	return innovations
}

