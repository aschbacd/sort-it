package utils

import (
	"fmt"

	"github.com/fatih/color"
)

// PrintBanner prints the banner for sort-it
func PrintBanner() {
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

// PrintMessage prints a message
func PrintMessage(message, mode string) {
	// Colors
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Set options
	messageTemplate := "[ %s ] %s\n"

	// Print message
	switch mode {
	case "success":
		fmt.Printf(messageTemplate, green("SUCCESS"), message)
	case "info":
		fmt.Printf(messageTemplate, cyan(" INFO "), message)
	case "warning":
		fmt.Printf(messageTemplate, yellow("WARNING"), message)
	case "error":
		fmt.Printf(messageTemplate, red(" ERROR "), message)
	}
}
