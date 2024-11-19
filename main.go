package main

import (
	"fmt"
	"log"
	"os"

	"github.com/UCLALibrary/pt-tools/cmd/ptcp"
	"github.com/UCLALibrary/pt-tools/cmd/ptls"
	"github.com/UCLALibrary/pt-tools/cmd/ptmv"
	"github.com/UCLALibrary/pt-tools/cmd/ptnew"
	"github.com/UCLALibrary/pt-tools/cmd/ptrm"
)

const help = `pt facilitates interactions with a Pairtree without the user needing to know about the Pairtree’s internal structure. 

Please refer to the README(https://github.com/UCLALibrary/pt-tools) for more detailed instructions

	Usage: pt [command] [options]
	Commands:
	  ls     List directories and files
	  rm     Remove files or directories
	  cp     Copy files or directories
	  mv     Move files or directories
	  new    Create a new pairtree object
	
	For more information on a specific command, run 'pt [command] --help'.`

func main() {
	// Basic command-line argument parsing
	if len(os.Args) < 2 {
		fmt.Println(help)
		os.Exit(1)
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
			os.Exit(2)
		}
	case "rm":
		err := ptrm.Run(args, writer)
		if err != nil {
			os.Exit(3)
		}
	case "cp":
		err := ptcp.Run(args, writer)
		if err != nil {
			os.Exit(4)
		}
	case "mv":
		err := ptmv.Run(args, writer)
		if err != nil {
			os.Exit(5)
		}
	case "new":
		err := ptnew.Run(args, writer)
		if err != nil {
			os.Exit(6)
		}
	default:
		fmt.Println(help)
		log.Fatalf("Unknown command: %s", command)
	}
}
