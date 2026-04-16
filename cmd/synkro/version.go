package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "2.1.0"
	Commit    = "unknown"
	BuildTime = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show Synkro version",
	Long:  "Display version information including build details",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Synkro v%s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Built: %s\n", BuildTime)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
