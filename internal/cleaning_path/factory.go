package cleaning_path

import (
	"regexp"
	"sync"
)

var regexForExpired = regexp.MustCompile(`dateexpire = '(\d+)'`)
var dirsExp = []string{"managed_cache", "stack_cache", "cache"}
var dirsAll = []string{"managed_cache", "stack_cache", "cache", "html_pages"}

type cleaningPath struct {
	pathToBitrixRoot string
	fileCounter      uint32
	processAllFiles  bool
	dryRun           bool
	ignoreErrors     bool
	wg               *sync.WaitGroup
}

func NewCleaningPath(pathToBitrixRoot string, processAllFiles bool, dryRun bool, ignoreErrors bool) *cleaningPath {
	return &cleaningPath{
		pathToBitrixRoot: pathToBitrixRoot,
		processAllFiles:  processAllFiles,
		dryRun:           dryRun,
		ignoreErrors:     ignoreErrors,
		wg:               &sync.WaitGroup{},
	}
}

func (cp *cleaningPath) GetCounter() uint32 {
	return cp.fileCounter
}
