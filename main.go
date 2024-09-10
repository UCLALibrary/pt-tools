package main

import (
	"log"
	"os"

	"github.com/UCLALibrary/pt-tools/cmd/ptls"
	"github.com/UCLALibrary/pt-tools/cmd/ptrm"
)

func main() {
	// Basic command-line argument parsing
	if len(os.Args) < 2 {
		log.Fatal("No command specified")
	}

	command := os.Args[1]
	// Pass in os.Args excluding the general and specifc program name
	args := os.Args[2:]

	// Use os.Stdout for standard output
	writer := os.Stdout

	switch command {
	case "ptls":
		err := ptls.Run(args, writer)
		if err != nil {
			os.Exit(1)
		}
	case "ptrm":
		err := ptrm.Run(args, writer)
		if err != nil {
			os.Exit(1)
		}
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
