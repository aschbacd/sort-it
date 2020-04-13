package app

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/aschbacd/sort-it/internal/utils"
)

var baseParts []string

// Sort uses its supplied parameters to sort files appropriately
func Sort(sourceFolder, destinationFolder string, copyDuplicates, duplicatesOnly, multimediaOnly bool) {
	// Check if exiftool is available
	if !duplicatesOnly {
		if !utils.CommandAvailable("exiftool") {
			utils.PrintMessage("exiftool not available", "error")
			os.Exit(1)
		}
	}

	// File lists
	hashList := []string{}
	errorList := []utils.File{}
	sortedList := []utils.File{}
	duplicatesList := []utils.File{}

	// Channels (get files)
	hashFiles := make(chan utils.File, 100)
	errorFiles := make(chan utils.File, 100)
	fileCount := make(chan int, 100)

	utils.PrintMessage("Getting files ...", "info")

	go func() {
		if err := utils.GetFilesWithHash(sourceFolder, hashFiles, errorFiles, fileCount); err != nil {
			utils.PrintMessage(err.Error(), "error")
			os.Exit(1)
		}
	}()

	baseParts = strings.Split(sourceFolder, "/")

	// Channels (sort files)
	sortedFiles := make(chan utils.File, 100)
	duplicateFiles := make(chan utils.File, 100)
	errorFilesValid := make(chan utils.File, 100)
	fileCountValid := make(chan int, 100)

	go func() {
		for range fileCount {
			select {
			case hashFile := <-hashFiles:
				// Check for duplicate
				if utils.Contains(hashList, hashFile.Hash) {
					// Duplicate
					if copyDuplicates {
						go CopyDuplicate(hashFile, destinationFolder, duplicateFiles, errorFilesValid)
					} else {
						hashFile.RelativePath = strings.Join(strings.Split(hashFile.Path, "/")[len(baseParts):], "/")
						duplicateFiles <- hashFile
					}
				} else {
					// New file
					hashList = append(hashList, hashFile.Hash)
					go SortFile(hashFile, destinationFolder, duplicatesOnly, multimediaOnly, sortedFiles, errorFilesValid)
				}

				// Increase count
				fileCountValid <- 0
			case errorFile := <-errorFiles:
				// Error
				errorList = append(errorList, errorFile)
				utils.PrintMessage(errorFile.Error+" - "+errorFile.Path, "error")
			}
		}

		// Close channels (get files)
		close(hashFiles)
		close(errorFiles)
		close(fileCountValid)
	}()

	// Add files to slices
	for range fileCountValid {
		select {
		case hashFileSorted := <-sortedFiles:
			sortedList = append(sortedList, hashFileSorted)
			utils.PrintMessage("successfully sorted "+hashFileSorted.Path, "success")
		case hashFileDuplicate := <-duplicateFiles:
			duplicatesList = append(duplicatesList, hashFileDuplicate)
			utils.PrintMessage("successfully sorted "+hashFileDuplicate.Path, "success")
		case errorFile := <-errorFilesValid:
			errorList = append(errorList, errorFile)
			utils.PrintMessage(errorFile.Error+" - "+errorFile.Path, "error")
		}
	}

	// Close channels (sort files)
	close(sortedFiles)
	close(duplicateFiles)
	close(errorFilesValid)

	utils.PrintMessage("Writing duplicate files ...", "info")
	utils.WriteFileLogs(destinationFolder, sortedList, duplicatesList)
	utils.PrintMessage("Writing error files ...", "info")
	utils.WriteErrorFiles(destinationFolder, errorList)
	utils.PrintMessage("Program finished successfully", "info")
}

// CopyDuplicate copies a duplicate to its destination folder
func CopyDuplicate(sourceFile utils.File, destinationFolder string, duplicateFiles chan<- utils.File, errorFiles chan<- utils.File) {
	// Get relative file path / destination path
	sourcePathRelative := strings.Join(strings.Split(sourceFile.Path, "/")[len(baseParts):], "/")
	destinationPath := path.Join(destinationFolder, "Errors", "Duplicates", sourcePathRelative)

	// Copy duplicate
	err := utils.CopyFile(sourceFile.Path, destinationPath)
	if err != nil {
		errorFiles <- utils.File{Path: sourceFile.Path, Error: err.Error()}
		return
	}

	// Add to channel
	duplicateFiles <- utils.File{Path: destinationPath, RelativePath: sourcePathRelative, Hash: sourceFile.Hash}
}

// SortFile sorts a folder
func SortFile(sourceFile utils.File, destinationFolder string, duplicatesOnly, multimediaOnly bool, sortedFiles chan<- utils.File, errorFiles chan<- utils.File) {
	// Get relative file path / destination path
	sourcePathRelative := strings.Join(strings.Split(sourceFile.Path, "/")[len(baseParts):], "/")
	destinationPath := path.Join(destinationFolder, "Data", sourcePathRelative)

	if !duplicatesOnly {
		// Multimedia
		metaFile, err := utils.GetFileMetadata(sourceFile.Path)
		if err != nil && multimediaOnly {
			errorFiles <- utils.File{Path: sourceFile.Path, Error: err.Error()}
			return
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
			}
		}
	}

	if destinationPath != path.Join(destinationFolder, "Data", sourcePathRelative) || !multimediaOnly {
		// Copy file
		err := utils.CopyFile(sourceFile.Path, destinationPath)
		if err != nil {
			errorFiles <- utils.File{Path: sourceFile.Path, Error: err.Error()}
			return
		}

		// Get relative destination path
		destinationPathRelative := strings.Join(strings.Split(destinationPath, "/")[len(strings.Split(destinationFolder, "/")):], "/")

		// Add to channel
		sortedFiles <- utils.File{Path: destinationPath, RelativePath: destinationPathRelative, Hash: sourceFile.Hash}
	}
}
