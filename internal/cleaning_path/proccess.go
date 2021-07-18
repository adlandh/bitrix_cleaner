package cleaning_path

import (
	"log"
	"os"

	"github.com/stretchr/powerwalk"
)

func (cp *cleaningPath) Run() error {
	if err := cp.checkBitrixDir(); err != nil {
		return err
	}

	dirs := dirsExp
	if cp.processAllFiles {
		dirs = dirsAll
	}

	for _, dir := range dirs {
		fullPath := cp.pathToBitrixRoot + string(os.PathSeparator) + "bitrix" + string(os.PathSeparator) + dir
		cp.wg.Add(1)
		go cp.ProcessDir(fullPath)
	}

	cp.wg.Wait()
	return nil
}

func (cp *cleaningPath) ProcessDir(path string) {
	var err error
	defer cp.wg.Done()
	if _, err = os.Stat(path); err != nil {
		log.Println(err)
		return
	}
	log.Println("Start processing " + path)
	if cp.processAllFiles {
		err = powerwalk.Walk(cp.pathToBitrixRoot, cp.processFiles)
	} else {
		err = powerwalk.Walk(cp.pathToBitrixRoot, cp.processExpiredFiles)
	}
	if err != nil && !cp.ignoreErrors {
		log.Println(err)
	}
	log.Println("Done processing " + path)
}
