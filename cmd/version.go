package cmd

import (
	"fmt"
	"os"
)

var Version = "dev v2.0.0."

func printVersionAndExit() {
	fmt.Printf("renderctl %s\n", Version)
	os.Exit(0)
}
