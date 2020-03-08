package main

import (
	"os"
	"path"
	"path/filepath"
)

func main() {
	// Show banner
	printBanner()

	// Get arguments
	args := os.Args[1:]

	// Check help
	if len(args) >= 1 && args[0] == "--help" || args[0] == "-h" {
		printHelp()
		os.Exit(0)
	}

	// Check if arguments supplied
	if len(args) >= 2 {
		sourceFolder := filepath.ToSlash(args[len(args)-2])
		destinationFolder := filepath.ToSlash(args[len(args)-1])

		var err error

		// Arguments
		switch args[0] {
		case "--copy-duplicates":
			// Normal mode + copy duplicates -> sort everything (including multimedia), copy duplicates
			err = SortIt(sourceFolder, destinationFolder, "normal", true)
		case "--duplicates-only":
			// Duplicates only -> only check for duplicates and generate info
			if args[1] == "--copy-duplicates" {
				// Extra case -> copy duplicates
				err = SortIt(sourceFolder, destinationFolder, "duplicates-only", true)
			} else {
				// Don't copy duplicates
				err = SortIt(sourceFolder, destinationFolder, "duplicates-only", false)
			}
		case "--multimedia-only":
			// Multimedia only -> only copy sorted multimedia files
			if args[1] == "--copy-duplicates" {
				// Extra case -> copy duplicates
				err = SortIt(sourceFolder, destinationFolder, "multimedia-only", true)
			} else {
				// Don't copy duplicates
				err = SortIt(sourceFolder, destinationFolder, "multimedia-only", false)
			}
		default:
			// Normal mode -> sort everything (including multimedia), don't copy duplicates
			err = SortIt(sourceFolder, destinationFolder, "normal", false)
		}

		// Check for error
		if err != nil {
			// Error -> show error message
			printMessage(err.Error(), "error", false, false)
		} else {
			// Set duplicates html link parent
			duplicatesFolder := sourceFolder
			if args[0] == "--copy-duplicates" || args[1] == "--copy-duplicates" {
				duplicatesFolder = path.Join(destinationFolder, "Duplicates", "Files")
			}

			// Write duplicate files
			printMessage("Writing duplicate files", "info", false, false)
			writeDuplicateFiles(sourceFolder, destinationFolder, duplicatesFolder, duplicates)

			printMessage("Program finished successfully", "info", false, false)
		}
	} else {
		printHelp()
		os.Exit(1)
	}
}
