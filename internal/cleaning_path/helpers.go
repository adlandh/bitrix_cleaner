package cleaning_path

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func (cp *cleaningPath) incCounter() {
	atomic.AddUint32(&cp.fileCounter, 1)
}

func (cp cleaningPath) checkBitrixDir() error {
	fileInfo, err := os.Stat(cp.pathToBitrixRoot + string(os.PathSeparator) + "bitrix")
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return errors.New("path is not a directory: " + cp.pathToBitrixRoot)
	}

	return nil
}

func (cp *cleaningPath) processFiles(path string, info os.FileInfo, err error) error {
	if os.IsNotExist(err) || cp.ignoreErrors {
		return nil
	}
	if err != nil {
		return err
	}

	isFileToProcess := strings.HasSuffix(path, ".php") || strings.HasSuffix(path, ".html") ||
		strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js")

	if (info.Mode()&os.ModeType == 0) && isFileToProcess {
		err = cp.processFile(path)
		cp.incCounter()
	}
	return err
}

func (cp cleaningPath) processFile(path string) error {
	if cp.dryRun {
		log.Println("Removing " + path)
		return nil
	}
	err := os.Remove(path)
	if err != nil && cp.ignoreErrors {
		log.Println(err)
		return nil
	}

	return err
}

func (cp *cleaningPath) processExpiredFiles(path string, info os.FileInfo, err error) error {
	if os.IsNotExist(err) || cp.ignoreErrors {
		return nil
	}

	if err != nil {
		return err
	}

	if info.Mode()&os.ModeType == 0 && info.Size() > 100 && strings.HasSuffix(path, ".php") {
		err = cp.processExpiredFile(path)
		cp.incCounter()
	}

	return err
}

func (cp *cleaningPath) processExpiredFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	reader := bufio.NewReader(file)
	var line string

	for lineNo := int8(0); lineNo < 5; lineNo++ {
		line, err = reader.ReadString('\n')
		if err != nil {
			return err
		}
	}

	match := regexForExpired.FindStringSubmatch(line)
	if match == nil {
		return nil
	}

	tm, err := strconv.ParseInt(match[1], 10, 0)
	if err == nil {
		if tm < time.Now().Unix() {
			if cp.dryRun {
				fmt.Println("Removing " + path)
				return nil
			}
			err = os.Remove(path)
			if err != nil {
				if cp.ignoreErrors {
					log.Println(err)
					return nil
				}
				return err
			}
		}
	}
	return nil
}
