package app

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aschbacd/sort-it/internal/utils"
)

var duplicates map[string]*utils.DuplicateCollection
var errors []utils.ErrorFile
var baseParts []string

// Sort uses its supplied parameters to sort files appropriately
func Sort(sourceFolder, destinationFolder string, copyDuplicates, duplicatesOnly, multimediaOnly bool) {
	// Check if exiftool is available
	if !duplicatesOnly {
		if !utils.CommandAvailable("exiftool") {
			utils.PrintMessage("exiftool not available", "error", false, false)
			os.Exit(1)
		}
	}

	// Get files and file count
	files, err := utils.GetFilesWithFileCount(sourceFolder)
	if err != nil {
		utils.PrintMessage(err.Error(), "error", false, false)
	}

	// Initialize fields
	duplicates = map[string]*utils.DuplicateCollection{}
	errors = []utils.ErrorFile{}

	fileCount := float64(len(files))
	baseParts = strings.Split(sourceFolder, "/")

	// Sort files
	for index, filePath := range files {
		if err := SortFile(filePath, destinationFolder, copyDuplicates, duplicatesOnly, multimediaOnly); err != nil {
			errors = append(errors, utils.ErrorFile{Path: filePath, Error: err.Error()})
		}

		// Increase counter
		progress := strconv.FormatFloat(float64(index+1)/fileCount*100, 'f', 2, 64)
		utils.PrintMessage("Progress: "+progress+"%", "info", true, true)
	}

	// Write duplicate and error files
	duplicatesFolder := sourceFolder
	if copyDuplicates {
		duplicatesFolder = path.Join(destinationFolder, "Errors", "Duplicates")
	}

	println()

	utils.PrintMessage("Writing duplicate files", "info", false, false)
	utils.WriteDuplicateFiles(destinationFolder, duplicatesFolder, duplicates)
	utils.PrintMessage("Writing error files", "info", false, false)
	utils.WriteErrorFiles(destinationFolder, errors)
	utils.PrintMessage("Program finished successfully", "info", false, false)
}

// SortFile sorts a folder
func SortFile(filePath, destinationFolder string, copyDuplicates, duplicatesOnly, multimediaOnly bool) error {
	// File -> get md5 hash and check if file is duplicate
	file, err := os.Open(filePath)
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

	// Get md5 sum
	sum := hex.EncodeToString(hash.Sum(nil))

	fileParts := strings.Split(filePath, "/")
	filePathRelative := strings.Join(fileParts[len(baseParts):], "/")

	// Check if duplicate
	if duplicates[sum] != nil {
		// Copy duplicate
		if copyDuplicates {
			destinationPath := path.Join(destinationFolder, "Errors", "Duplicates", filePathRelative)

			err := os.MkdirAll(path.Dir(destinationPath), 0750)
			if err != nil {
				return err
			}

			// Copy duplicate
			_, err = utils.CopyFile(filePath, destinationPath)
			if err != nil {
				return err
			}
		}

		// Save duplicate
		duplicatesList := duplicates[sum].Duplicates
		duplicatesList = append(duplicatesList, filePathRelative)

		duplicates[sum] = &utils.DuplicateCollection{Path: duplicates[sum].Path, Hash: sum, Duplicates: duplicatesList}
	} else {
		// New file
		destinationPath := path.Join(destinationFolder, "Data", filePathRelative)
		valid := true

		if !duplicatesOnly {
			// Multimedia
			metaFile, err := utils.GetFileMetadata(filePath)
			if err != nil {
				return err
			}

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
					if multimediaOnly {
						valid = false
					}
				}
			} else {
				// Only copy multimedia files
				if multimediaOnly {
					valid = false
				}
			}
		}

		if valid {
			// Copy file
			err = os.MkdirAll(path.Dir(destinationPath), 0750)
			if err != nil {
				return err
			}

			_, err = utils.CopyFile(filePath, destinationPath)
			if err != nil {
				return err
			}

			destinationFolderParts := strings.Split(destinationFolder, "/")
			destinationFileParts := strings.Split(destinationPath, "/")
			destinationFilePathRelative := strings.Join(destinationFileParts[len(destinationFolderParts):], "/")

			duplicates[sum] = &utils.DuplicateCollection{Path: destinationFilePathRelative, Hash: sum}
		}
	}

	return nil
}
