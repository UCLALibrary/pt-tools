package main

import (
	"log"
	"os"

	"github.com/UCLALibrary/pt-tools/cmd/ptls"
	// Import other tools as needed
)

func main() {
	// Basic command-line argument parsing
	if len(os.Args) < 2 {
		log.Fatal("No command specified")
	}

	command := os.Args[1]

	switch command {
	case "ptls":
		err := ptls.Run()
		if err != nil {
			log.Fatal("Error running ptls:", err)
		}
		// Add cases for other tools as needed
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
