package main

import (
	"log"
	"os"

	"github.com/UCLALibrary/pt-tools/cmd/ptcp"
	"github.com/UCLALibrary/pt-tools/cmd/ptls"
	"github.com/UCLALibrary/pt-tools/cmd/ptmv"
	"github.com/UCLALibrary/pt-tools/cmd/ptnew"
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
	case "ls":
		err := ptls.Run(args, writer)
		if err != nil {
			os.Exit(1)
		}
	case "rm":
		err := ptrm.Run(args, writer)
		if err != nil {
			os.Exit(1)
		}
	case "cp":
		err := ptcp.Run(args, writer)
		if err != nil {
			os.Exit(1)
		}
	case "mv":
		err := ptmv.Run(args, writer)
		if err != nil {
			os.Exit(1)
		}
	case "new":
		err := ptnew.Run(args, writer)
		if err != nil {
			os.Exit(1)
		}
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
