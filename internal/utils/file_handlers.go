package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
)

// File represents a file including only necessary meta data
type File struct {
	SourceFile        string `json:"SourceFile"`
	FileName          string `json:"FileName"`
	Directory         string `json:"Directory"`
	FileSize          string `json:"FileSize"`
	FileType          string `json:"FileType"`
	FileTypeExtension string `json:"FileTypeExtension"`
	MIMEType          string `json:"MIMEType"`
	CreateDate        string `json:"CreateDate"`
	Album             string `json:"Album"`
	Artist            string `json:"Artist"`
	Title             string `json:"Title"`
}

// DuplicateCollection represents a scanned file and also contains all its duplicates
type DuplicateCollection struct {
	Path       string
	Hash       string
	Duplicates []string
}

// ErrorFile represents a file that got an error
type ErrorFile struct {
	Path  string
	Error string
}

// GetFilesWithFileCount recursively counts all files in a directory
func GetFilesWithFileCount(sourcePath string) ([]string, error) {
	count := 0
	files := []string{}

	if err := filepath.Walk(sourcePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				count++
				PrintMessage("Getting files: "+strconv.Itoa(count), "info", true, true)
				files = append(files, path)
			}

			return nil
		}); err != nil {
		return nil, err
	}

	println()

	return files, nil
}

// GetFileMetadata
func GetFileMetadata(path string) (File, error) {
	// Get metadata with exiftool
	command := exec.Command("exiftool", "-json", path)
	var out bytes.Buffer
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		return File{}, err
	}

	// Unmarshal output into file
	var metaFiles []File
	err = json.Unmarshal(out.Bytes(), &metaFiles)
	if err != nil {
		return File{}, err
	}

	if len(metaFiles) == 0 {
		return File{}, fmt.Errorf("cannot get metadata for this file")
	}

	return metaFiles[0], nil
}

// CopyFile copies a file from source to destination
func CopyFile(src, dst string) (int64, error) {
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

// WriteDuplicateFiles
func WriteDuplicateFiles(destinationFolder, duplicatesFolder string, duplicates map[string]*DuplicateCollection) {
	// Create destination path
	destinationPath := path.Join(destinationFolder, "Errors")
	err := os.MkdirAll(destinationPath, 0750)
	if err != nil {
		PrintMessage(err.Error(), "error", false, false)
		os.Exit(1)
	}

	// Write html file
	htmlString := "<!DOCTYPE html><html lang='en'><head><meta charset='UTF-8'/><title>Duplicates</title></head><body><h1>Duplicates</h1><ul>"
	sortedDuplicates := []*DuplicateCollection{}

	for _, file := range duplicates {
		if len(file.Duplicates) > 0 {
			sortedDuplicates = append(sortedDuplicates, file)
			htmlString = htmlString + "<li><p><a href='file:///" + path.Join(destinationFolder, file.Path) + "' target='_blank'>" + file.Path + " (" + file.Hash + ")</a></p><ul>"
			for _, duplicate := range file.Duplicates {
				htmlString = htmlString + "<li><p><a href='file:///" + path.Join(duplicatesFolder, duplicate) + "' target='_blank'>" + duplicate + "</a></p></li>"
			}
			htmlString = htmlString + "</ul></li>"
		}
	}

	htmlString = htmlString + "</ul></body></html>"

	err = ioutil.WriteFile(path.Join(destinationPath, "sort-it_duplicates.html"), []byte(htmlString), 0750)
	if err != nil {
		PrintMessage(err.Error(), "error", false, false)
		os.Exit(1)
	}

	// Write json file
	jsonString, err := json.MarshalIndent(sortedDuplicates, "", "    ")
	if err != nil {
		PrintMessage(err.Error(), "error", false, false)
		os.Exit(1)
	}

	err = ioutil.WriteFile(path.Join(destinationPath, "sort-it_duplicates.json"), jsonString, 0750)
	if err != nil {
		PrintMessage(err.Error(), "error", false, false)
		os.Exit(1)
	}
}

// WriteErrorFiles
func WriteErrorFiles(destinationFolder string, errors []ErrorFile) {
	// Create destination path
	destinationPath := path.Join(destinationFolder, "Errors")
	err := os.MkdirAll(destinationPath, 0750)
	if err != nil {
		PrintMessage(err.Error(), "error", false, false)
		os.Exit(1)
	}

	// Write json file
	jsonString, err := json.MarshalIndent(errors, "", "    ")
	if err != nil {
		PrintMessage(err.Error(), "error", false, false)
		os.Exit(1)
	}

	err = ioutil.WriteFile(path.Join(destinationPath, "sort-it_errors.json"), jsonString, 0750)
	if err != nil {
		PrintMessage(err.Error(), "error", false, false)
		os.Exit(1)
	}
}
