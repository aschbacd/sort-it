package app

import (
	"path"
	"strings"
	"time"

	"github.com/aschbacd/sort-it/pkg/logger"
	"github.com/aschbacd/sort-it/pkg/utils"
)

var baseParts []string

// Sort uses its supplied parameters to sort files appropriately
func Sort(sourceFolder, destinationFolder string, copyDuplicates, duplicatesOnly, multimediaOnly bool) {
	// File lists
	hashList := []string{}
	errorList := []File{}
	sortedList := []File{}
	duplicatesList := []File{}

	// Channels (get files)
	hashFiles := make(chan File, 100)
	errorFiles := make(chan File, 100)
	fileCount := make(chan int, 100)

	logger.Info("Getting files ...")

	go func() {
		if err := GetFilesWithHash(sourceFolder, hashFiles, errorFiles, fileCount); err != nil {
			logger.Fatal(err.Error())
		}
	}()

	baseParts = strings.Split(sourceFolder, "/")

	// Channels (sort files)
	sortedFiles := make(chan File, 100)
	duplicateFiles := make(chan File, 100)
	errorFilesValid := make(chan File, 100)
	fileCountValid := make(chan int, 100)

	go func() {
		for range fileCount {
			select {
			case hashFile := <-hashFiles:
				// Check for duplicate
				if utils.SliceContainsString(hashList, hashFile.Hash) {
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
				logger.Error(errorFile.Error + " - " + errorFile.Path)
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
			logger.Info("successfully sorted " + hashFileSorted.Path)
		case hashFileDuplicate := <-duplicateFiles:
			duplicatesList = append(duplicatesList, hashFileDuplicate)
			logger.Info("successfully sorted " + hashFileDuplicate.Path)
		case errorFile := <-errorFilesValid:
			errorList = append(errorList, errorFile)
			logger.Error(errorFile.Error + " - " + errorFile.Path)
		}
	}

	// Close channels (sort files)
	close(sortedFiles)
	close(duplicateFiles)
	close(errorFilesValid)

	logger.Info("Writing duplicate files ...")
	WriteFileLogs(destinationFolder, sortedList, duplicatesList)
	logger.Info("Writing error files ...")
	WriteErrorFiles(destinationFolder, errorList)
	logger.Info("Program finished successfully")
}

// CopyDuplicate copies a duplicate to its destination folder
func CopyDuplicate(sourceFile File, destinationFolder string, duplicateFiles chan<- File, errorFiles chan<- File) {
	// Get relative file path / destination path
	sourcePathRelative := strings.Join(strings.Split(sourceFile.Path, "/")[len(baseParts):], "/")
	destinationPath := path.Join(destinationFolder, "Errors", "Duplicates", sourcePathRelative)

	// Copy duplicate
	err := CopyFile(sourceFile.Path, destinationPath)
	if err != nil {
		errorFiles <- File{Path: sourceFile.Path, Error: err.Error()}
		return
	}

	// Add to channel
	duplicateFiles <- File{Path: destinationPath, RelativePath: sourcePathRelative, Hash: sourceFile.Hash}
}

// SortFile sorts a folder
func SortFile(sourceFile File, destinationFolder string, duplicatesOnly, multimediaOnly bool, sortedFiles chan<- File, errorFiles chan<- File) {
	// Get relative file path / destination path
	sourcePathRelative := strings.Join(strings.Split(sourceFile.Path, "/")[len(baseParts):], "/")
	destinationPath := path.Join(destinationFolder, "Data", sourcePathRelative)

	if !duplicatesOnly {
		// Multimedia
		metaFile, err := GetFileMetadata(sourceFile.Path)
		if err != nil && multimediaOnly {
			errorFiles <- File{Path: sourceFile.Path, Error: err.Error()}
			return
		}

		// Check if create date available
		fileDate, err := time.Parse("2006:01:02 15:04:05", metaFile.CreateDate)
		if err != nil {
			// Invalid creation date / time
			fileDate, _ = time.Parse("2006:01:02 15:04:05-07:00", metaFile.CreateDate)
		}

		mimeTypeParts := strings.Split(metaFile.MIMEType, "/")

		if !fileDate.IsZero() {
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
		} else if mimeTypeParts[0] == "audio" {
			// Audio (Music)
			if metaFile.Artist != "" && metaFile.Album != "" && metaFile.Title != "" {
				destinationPath = path.Join(destinationFolder, "Multimedia", "Audio", "Music", metaFile.Artist, metaFile.Album, metaFile.Title+"."+metaFile.FileTypeExtension)
			}
		}
	}

	if destinationPath != path.Join(destinationFolder, "Data", sourcePathRelative) || !multimediaOnly {
		// Copy file
		err := CopyFile(sourceFile.Path, destinationPath)
		if err != nil {
			errorFiles <- File{Path: sourceFile.Path, Error: err.Error()}

			// Copy to errors
			destinationPath = path.Join(destinationFolder, "Errors", "Files", sourcePathRelative)
			err := CopyFile(sourceFile.Path, destinationPath)
			if err != nil {
				errorFiles <- File{Path: sourceFile.Path, Error: err.Error()}
			}

			return
		}

		// Get relative destination path
		destinationPathRelative := strings.Join(strings.Split(destinationPath, "/")[len(strings.Split(destinationFolder, "/")):], "/")

		// Add to channel
		sortedFiles <- File{Path: destinationPath, RelativePath: destinationPathRelative, Hash: sourceFile.Hash}
	}
}
