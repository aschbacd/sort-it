package app

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"

	"github.com/aschbacd/sort-it/pkg/logger"
)

// File represents a scanned file
type File struct {
	Path         string `json:"path,omitempty"`
	RelativePath string `json:"relative_path,omitempty"`
	Hash         string `json:"hash,omitempty"`
	Error        string `json:"error,omitempty"`
	Duplicates   []File `json:"duplicates,omitempty"`
}

// MetaFile represents a file including only necessary meta data
type MetaFile struct {
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

// GetFilesWithHash recursively scans all files in a directory and calculates its hash sum
func GetFilesWithHash(sourcePath string, hashFiles chan<- File, errorFiles chan<- File, fileCount chan<- int) error {
	err := filepath.Walk(sourcePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				// Add file count and calculate hash sum
				fileCount <- 0
				go GetFileHash(path, hashFiles, errorFiles)
			}

			return nil
		})

	// Close channel
	close(fileCount)

	return err
}

// GetFileHash returns the hash sum for a file
func GetFileHash(path string, hashFiles chan<- File, errorFiles chan<- File) {
	// Get file
	file, err := os.Open(path)
	if err != nil {
		errorFiles <- File{Path: path, Error: err.Error()}
		return
	}
	defer file.Close()

	// Get md5 hash
	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		errorFiles <- File{Path: path, Error: err.Error()}
		return
	}

	// Get md5 sum
	sum := hex.EncodeToString(hash.Sum(nil))
	hashFiles <- File{Path: path, Hash: sum}
}

// GetFileMetadata returns the metadata for a file
func GetFileMetadata(path string) (MetaFile, error) {
	// Get metadata with exiftool
	command := exec.Command("exiftool", "-json", path)
	var out bytes.Buffer
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		return MetaFile{}, err
	}

	// Unmarshal output into file
	var metaFiles []MetaFile
	err = json.Unmarshal(out.Bytes(), &metaFiles)
	if err != nil {
		return MetaFile{}, err
	}

	// Check output
	if len(metaFiles) != 1 {
		return MetaFile{}, fmt.Errorf("cannot get metadata for this file")
	}

	return metaFiles[0], nil
}

// GetDuplicates returns all duplicates for a given hash
func GetDuplicates(files []File, hash string) []File {
	duplicates := []File{}
	for _, file := range files {
		if file.Hash == hash {
			duplicates = append(duplicates, file)
		}
	}
	return duplicates
}

// CopyFile copies a file from source to destination
func CopyFile(sourcePath, destinationPath string) error {
	// Check source file
	sourceFileStat, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", sourcePath)
	}

	// Check if destination file already exists
	if _, err := os.Stat(destinationPath); os.IsNotExist(err) {
		// Open file
		source, err := os.Open(sourcePath)
		if err != nil {
			return err
		}
		defer source.Close()

		// Create destination folder
		err = os.MkdirAll(path.Dir(destinationPath), 0750)
		if err != nil {
			return err
		}

		// Copy file
		destination, err := os.Create(destinationPath)
		if err != nil {
			return err
		}
		defer destination.Close()
		_, err = io.Copy(destination, source)
		return err
	} else {
		// Destination file already exists
		return fmt.Errorf("destination file already exists (" + destinationPath + ")")
	}
}

// WriteFileLogs creates html/json files for file logs
func WriteFileLogs(destinationFolder string, sortedFiles, duplicateFiles []File) {
	// Sort lists
	sort.SliceStable(sortedFiles, func(i, j int) bool {
		return sortedFiles[i].Path < sortedFiles[j].Path
	})

	sort.SliceStable(duplicateFiles, func(i, j int) bool {
		return duplicateFiles[i].Path < duplicateFiles[j].Path
	})

	// Merge sorted files and duplicates
	sortedFilesWithDuplicates := []File{}
	for _, file := range sortedFiles {
		file.Duplicates = GetDuplicates(duplicateFiles, file.Hash)
		if len(file.Duplicates) > 0 {
			sortedFilesWithDuplicates = append(sortedFilesWithDuplicates, file)
		}
	}

	// Create destination path
	destinationPath := path.Join(destinationFolder, "Errors")
	err := os.MkdirAll(destinationPath, 0750)
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Write html file
	htmlString := "<!DOCTYPE html><html lang='en'><head><meta charset='UTF-8'/><title>Duplicates</title></head><body><h1>Duplicates</h1><ul>"

	for _, file := range sortedFilesWithDuplicates {
		htmlString = htmlString + "<li><p><a href='file://" + file.Path + "' target='_blank'>" + file.RelativePath + " (" + file.Hash + ")</a></p><ul>"
		for _, duplicate := range file.Duplicates {
			htmlString = htmlString + "<li><p><a href='file://" + duplicate.Path + "' target='_blank'>" + duplicate.RelativePath + "</a></p></li>"
		}
		htmlString = htmlString + "</ul></li>"
	}

	htmlString = htmlString + "</ul></body></html>"

	err = ioutil.WriteFile(path.Join(destinationPath, "sort-it_duplicates.html"), []byte(htmlString), 0750)
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Write json file
	jsonString, err := json.MarshalIndent(sortedFilesWithDuplicates, "", "    ")
	if err != nil {
		logger.Fatal(err.Error())
	}

	err = ioutil.WriteFile(path.Join(destinationPath, "sort-it_duplicates.json"), jsonString, 0750)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

// WriteErrorFiles
func WriteErrorFiles(destinationFolder string, errors []File) {
	// Sort errors
	sort.SliceStable(errors, func(i, j int) bool {
		return errors[i].Path < errors[j].Path
	})

	// Create destination path
	destinationPath := path.Join(destinationFolder, "Errors")
	err := os.MkdirAll(destinationPath, 0750)
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Write json file
	jsonString, err := json.MarshalIndent(errors, "", "    ")
	if err != nil {
		logger.Fatal(err.Error())
	}

	err = ioutil.WriteFile(path.Join(destinationPath, "sort-it_errors.json"), jsonString, 0750)
	if err != nil {
		logger.Fatal(err.Error())
	}
}
