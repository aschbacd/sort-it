package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/fatih/color"
)

// *************** PRINT *************** //

func printBanner() {
	fmt.Print(
		` _______  _______  ______    _______         ___   _______ 
|       ||       ||    _ |  |       |       |   | |       |
|  _____||   _   ||   | ||  |_     _| ____  |   | |_     _|
| |_____ |  | |  ||   |_||_   |   |  |____| |   |   |   |  
|_____  ||  |_|  ||    __  |  |   |         |   |   |   |  
 _____| ||       ||   |  | |  |   |         |   |   |   |  
|_______||_______||___|  |_|  |___|         |___|   |___|`)

	fmt.Print("\n\n                  - YOUR DATA ORGANIZER -\n\n")
}

func printMessage(message, mode string, overrideLast, skipNewLine bool) {
	// Colors
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Set options
	messageTemplate := "[ %s ] %s\n"

	if overrideLast && skipNewLine {
		messageTemplate = "\r[ %s ] %s"
	} else if overrideLast && !skipNewLine {
		messageTemplate = "\r[ %s ] %s\n"
	}

	// Print message
	switch mode {
	case "info":
		fmt.Printf(messageTemplate, cyan("INFO"), message)
	case "warning":
		fmt.Printf(messageTemplate, yellow("WARNING"), message)
	case "error":
		fmt.Printf(messageTemplate, red("ERROR"), message)
	}
}

func printHelp() {
	// Print help
	println("Usage: sort-it [OPTION]... [SOURCE-FOLDER] [DESTINATION-FOLDER]")
	println("Scan source folder recursively and sort files according to supplied options.")
	println("Example: sort-it ./source/ ./destination/")
	println("\nOptions:")
	println("  --copy-duplicates                        full sort, copy duplicates into subfolder of destination")
	println("  --duplicates-only                        don't check file type (mulitmedia)")
	println("  --duplicates-only --copy-duplicates      don't check file type (multimedia) and copy duplicates")
	println("  --multimedia-only                        only sort files of type multimedia, ignore other file types")
	println("  --multimedia-only --copy-duplicates      only sort files of type multimedia, ignore other file types, and copy duplicates")
	os.Exit(1)
}

// *************** CHECKS *************** //

func checkFolder(path string) bool {
	if _, err := os.Stat(path); err != nil {
		// Invalid folder
		return false
	}

	return true
}

// *************** OTHER UTILITIES *************** //

func getFileCount(sourcePath string) int {
	count := 0

	err := filepath.Walk(sourcePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				count++
				printMessage("Getting files: "+strconv.Itoa(count), "info", true, true)
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	println("")

	return count
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func writeDuplicateFiles(sourceFolder, destinationFolder, duplicatesFolder string, duplicates map[string]*DuplicateCollection) {
	// Set destination path
	destinationPath := path.Join(destinationFolder, "Duplicates")
	err := os.MkdirAll(destinationPath, 0777)
	if err != nil {
		printMessage(err.Error(), "error", false, false)
		os.Exit(1)
	}

	// Write html file
	htmlString := "<!DOCTYPE html><html lang='en'><head><meta charset='UTF-8'/><title>Duplicates</title></head><body><h1>Duplicates</h1><ul>"
	sortedDuplicates := []*DuplicateCollection{}

	for _, file := range duplicates {
		if len(file.Duplicates) > 0 {
			sortedDuplicates = append(sortedDuplicates, file)
			htmlString = htmlString + "<li><p><a href='file:///" + path.Join(sourceFolder, file.Path) + "' target='_blank'>" + file.Path + " (" + file.Hash + ")</a></p><ul>"
			for _, duplicate := range file.Duplicates {
				htmlString = htmlString + "<li><p><a href='file:///" + path.Join(duplicatesFolder, duplicate) + "' target='_blank'>" + duplicate + "</a></p></li>"
			}
			htmlString = htmlString + "</ul></li>"
		}
	}

	htmlString = htmlString + "</ul></body></html>"

	err = ioutil.WriteFile(path.Join(destinationPath, "sort-it_duplicates.html"), []byte(htmlString), 0777)
	if err != nil {
		printMessage(err.Error(), "error", false, false)
		os.Exit(1)
	}

	// Write json file
	jsonString, err := json.MarshalIndent(sortedDuplicates, "", "    ")
	if err != nil {
		printMessage(err.Error(), "error", false, false)
		os.Exit(1)
	}

	err = ioutil.WriteFile(path.Join(destinationPath, "sort-it_duplicates.json"), jsonString, 0777)
	if err != nil {
		printMessage(err.Error(), "error", false, false)
		os.Exit(1)
	}

}
