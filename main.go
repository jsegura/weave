package main

import (
	"callumj.com/weave/core"
	"callumj.com/weave/upload"
	"callumj.com/weave/upload/uptypes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	args := os.Args

	checkArgs(args)

	if strings.HasSuffix(args[1], ".enc") || len(args) >= 3 {
		performExtraction(args)
		return
	}

	abs, err := filepath.Abs(args[1])
	if err != nil {
		log.Printf("Unable to expand %v\r\n", args[1])
		panicQuit()
	}

	performCompilation(abs)
}

func checkArgs(args []string) {
	if len(args) == 1 {
		log.Printf("Usage: %v CONFIG_FILE\r\n", args[0])
		log.Printf("Usage: %v ENCRYPTED_FILE KEY_FILE [OUT_FILE]\r\n", args[0])
		panicQuit()
	}
}

func panicQuit() {
	os.Exit(1)
}
