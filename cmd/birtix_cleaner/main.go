package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/adlandh/bitrix_cleaner/internal/cleaning_path"
)

const ver = "1.3"

var pathToBitrixRoot string
var processAllFiles bool
var dryRun bool
var ignoreErrors bool
var showVersionAndExit bool

func init() {
	currentDir, err := os.Getwd()
	if err != nil {
		_ = fmt.Errorf("error getting current directory: %e", err)
		os.Exit(1)
	}

	flag.StringVar(&pathToBitrixRoot, "path", currentDir, "Path to bitrix root")
	flag.BoolVar(&processAllFiles, "all", false, "Process all files (if not provided then the expired files will be processed only)")
	flag.BoolVar(&dryRun, "dry-run", false, "Do not remove files. Run for testing")
	flag.BoolVar(&ignoreErrors, "ignore", false, "Do not stop processing on errors")
	flag.BoolVar(&showVersionAndExit, "version", false, "Show version number and exit")
	flag.Parse()
}

func main() {
	if showVersionAndExit {
		fmt.Println("Version: " + ver)
		os.Exit(0)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	pathToBitrixRoot = strings.TrimSuffix(pathToBitrixRoot, string(os.PathSeparator))

	prc := cleaning_path.NewCleaningPath(pathToBitrixRoot, processAllFiles, dryRun, ignoreErrors)
	if err := prc.Run(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Total removed %d files.\n", prc.GetCounter())
}
