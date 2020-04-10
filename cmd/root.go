package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aschbacd/sort-it/internal/app"
	"github.com/aschbacd/sort-it/internal/utils"
	"github.com/spf13/cobra"
)

var copyDuplicates bool
var duplicatesOnly bool
var multimediaOnly bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "sort-it [source folder] [destination folder]",
	Version: "1.0.0",
	Short:   "Sort your unorganized files with one command.",
	Long: `Sort your unorganized files with sort-it using only one command. This utility
is able to find duplicates, sort multimedia files like photos, videos, and
audio and also to create summary files in json as well as html where all
duplicates are listed.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("source and destination folder required")
		}

		if !utils.CheckFolder(args[0]) {
			return errors.New("source folder does not exist")
		}

		if !utils.CheckFolder(args[1]) {
			return errors.New("destination folder does not exist")
		}

		if duplicatesOnly && multimediaOnly {
			return errors.New("duplicates-only and multimedia-only cannot be used together")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Show banner
		utils.PrintBanner()

		// Convert path to slash (windows)
		sourceFolder := filepath.Clean(filepath.ToSlash(args[0]))
		destinationFolder := filepath.Clean(filepath.ToSlash(args[1]))

		// Sort folder
		app.Sort(sourceFolder, destinationFolder, copyDuplicates, duplicatesOnly, multimediaOnly)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&copyDuplicates, "copy-duplicates", false, "copy duplicates to destination folder")
	rootCmd.PersistentFlags().BoolVar(&duplicatesOnly, "duplicates-only", false, "only look for duplicate files, do not take account of file type")
	rootCmd.PersistentFlags().BoolVar(&multimediaOnly, "multimedia-only", false, "only sort photos, videos, and audio files, ignore all other file types")
}
