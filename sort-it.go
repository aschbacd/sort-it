package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

// DuplicateCollection represents a scanned file and also contains all its duplicates
type DuplicateCollection struct {
	Path       string
	Hash       string
	Duplicates []string
}

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

var duplicates map[string]*DuplicateCollection
var currentFile float64
var fileCount float64
var basePath string

// SortIt uses its supplied parameters to sort files appropriately
func SortIt(sourceFolder, destinationFolder, mode string, copyDuplicates bool) error {
	// Check folders
	if !checkFolder(sourceFolder) {
		printMessage("invalid source folder", "error", false, false)
		os.Exit(1)
	} else if !checkFolder(destinationFolder) {
		printMessage("invalid destination folder", "error", false, false)
		os.Exit(1)
	}

	// Get file count
	fileCount = float64(getFileCount(sourceFolder))

	// Set duplicates, basePath, and current file to their start value
	duplicates = map[string]*DuplicateCollection{}
	basePath = sourceFolder
	currentFile = 0

	// Sort folder
	err := SortFolder(sourceFolder, destinationFolder, mode, copyDuplicates)
	println()
	return err
}

// SortFolder sorts a folder
func SortFolder(sourceFolder, destinationFolder, mode string, copyDuplicates bool) error {
	// Get all files and directories
	content, err := ioutil.ReadDir(sourceFolder)
	if err != nil {
		return err
	}

	for _, f := range content {
		if f.IsDir() {
			// Directory -> sort sub directory
			err := SortFolder(path.Join(sourceFolder, f.Name()), destinationFolder, mode, copyDuplicates)
			if err != nil {
				return err
			}
		} else {
			// File -> get md5 hash and check if file is duplicate
			file, err := os.Open(path.Join(sourceFolder, f.Name()))
			defer file.Close()
			if err != nil {
				return err
			}

			// Get md5 hash
			hash := md5.New()
			_, err = io.Copy(hash, file)
			if err != nil {
				return err
			}

			// Name and sum
			filePath := file.Name()
			sum := hex.EncodeToString(hash.Sum(nil))

			baseParts := strings.Split(basePath, "/")
			fileParts := strings.Split(filePath, "/")
			filePathRelative := strings.Join(fileParts[len(baseParts):], "/")

			// Check if duplicate
			if duplicates[sum] != nil {
				// Copy duplicate
				if copyDuplicates {
					destinationPath := path.Join(destinationFolder, "Duplicates", "Files", filePathRelative)

					if copyDuplicates {
						err := os.MkdirAll(path.Dir(destinationPath), 0777)
						if err != nil {
							return err
						}

						// Copy duplicate
						_, err = copyFile(filePath, destinationPath)
						if err != nil {
							return err
						}
					}
				}

				// Save duplicate
				duplicatesList := duplicates[sum].Duplicates
				duplicatesList = append(duplicatesList, filePathRelative)

				duplicates[sum] = &DuplicateCollection{Path: duplicates[sum].Path, Hash: sum, Duplicates: duplicatesList}
			} else {
				// New file
				destinationPath := path.Join(destinationFolder, "Data", filePathRelative)
				valid := true

				if mode == "normal" || mode == "multimedia-only" {
					// Multimedia
					command := exec.Command("./exiftool.exe", "-json", filePath)
					var out bytes.Buffer

					// set the output to our variable
					command.Stdout = &out
					err = command.Run()
					if err != nil {
						log.Println(err)
					}

					var metaFiles []File
					err = json.Unmarshal(out.Bytes(), &metaFiles)
					if err != nil {
						return err
					}

					metaFile := metaFiles[0]

					// Check if create date available
					fileDate, err := time.Parse("2006:01:02 15:04:05", metaFile.CreateDate)
					if err != nil {
						// Invalid creation date / time
						fileDate, _ = time.Parse("2006:01:02 15:04:05-07:00", metaFile.CreateDate)
					}

					mimeTypeParts := strings.Split(metaFile.MIMEType, "/")

					if !fileDate.IsZero() || mimeTypeParts[0] == "audio" {
						switch mimeTypeParts[0] {
						case "image":
							// Image
							destinationPath = path.Join(destinationFolder, "Multimedia", "Pictures", fileDate.Format("2006"), fileDate.Format("01 January"), fileDate.Format("IMG_20060102_150405.")+metaFile.FileTypeExtension)
						case "video":
							// Video
							destinationPath = path.Join(destinationFolder, "Multimedia", "Videos", fileDate.Format("2006"), fileDate.Format("01 January"), fileDate.Format("VID_20060102_150405.")+metaFile.FileTypeExtension)
						case "audio":
							// Audio
							if metaFile.Artist != "" && metaFile.Album != "" && metaFile.Title != "" {
								destinationPath = path.Join(destinationFolder, "Multimedia", "Audio", "Music", metaFile.Artist, metaFile.Album, metaFile.Title+"."+metaFile.FileTypeExtension)
							} else {
								destinationPath = path.Join(destinationFolder, "Multimedia", "Audio", "Sounds", fileDate.Format("2006"), fileDate.Format("01 January"), fileDate.Format("AUD_20060102_150405.")+metaFile.FileTypeExtension)
							}
						default:
							// Only copy multimedia files
							if mode == "multimedia-only" {
								valid = false
							}
						}
					} else {
						// Only copy multimedia files
						if mode == "multimedia-only" {
							valid = false
						}
					}
				}

				if valid {
					// Copy file
					err = os.MkdirAll(path.Dir(destinationPath), 0777)
					if err != nil {
						return err
					}

					_, err = copyFile(filePath, destinationPath)
					if err != nil {
						return err
					}

					duplicates[sum] = &DuplicateCollection{Path: filePathRelative, Hash: sum}
				}
			}

			// Increase counter
			currentFile++
			progress := strconv.FormatFloat(currentFile/fileCount*100, 'f', 2, 64)
			printMessage("Progress: "+progress+"%", "info", true, true)
		}
	}
	return nil
}
