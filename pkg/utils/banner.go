package utils

import "github.com/common-nighthawk/go-figure"

// PrintBanner prints the application name in ascii art
func PrintBanner() {
	banner := figure.NewFigure("sort-it", "larry3d", true)
	banner.Print()
	println()
}
