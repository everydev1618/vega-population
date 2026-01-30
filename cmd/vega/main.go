package main

import (
	"fmt"
	"os"

	"github.com/martellcode/vega-population/population"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "population", "pop":
		if err := population.RunCLI(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	case "version", "-v", "--version":
		fmt.Println("vega version 0.1.0")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`vega - AI agent orchestration toolkit

Usage: vega <command> [options]

Commands:
  population, pop    Manage skills, personas, and profiles
  help               Show this help message
  version            Show version information

Run 'vega <command> help' for more information about a command.`)
}
