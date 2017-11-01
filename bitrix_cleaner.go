package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/stretchr/powerwalk"
	"io"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cleaningPath struct {
	dir     string
	counter int
	mutex   *sync.RWMutex
	regs    *regexp.Regexp
	done    chan<- int
	all     bool
	test    bool
	ignore  bool
	tmNow   int64
}

func NewCleaningPath(dir string, done chan<- int, all bool, test bool, ignore bool) *cleaningPath {
	return &cleaningPath{dir, 0, new(sync.RWMutex),
		regexp.MustCompile(`dateexpire = '(\d+)'`), done, all, test, ignore, time.Now().Unix()}
}

func (cp *cleaningPath) incCounter() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	cp.counter++
}

func (cp *cleaningPath) GetCounter() int {
	return cp.counter
}

func (cp *cleaningPath) ProcessDir() {
	go func() {
		_, err := os.Stat(cp.dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error processing "+cp.dir)
		} else {
			fmt.Println("Start processing " + cp.dir)
			if cp.all {
				err = powerwalk.Walk(cp.dir, cp.processFiles)
			} else {
				err = powerwalk.Walk(cp.dir, cp.processExpiredFiles)
			}

			if err != nil && !cp.ignore {
				fmt.Fprintln(os.Stderr, err)
			}
			fmt.Printf("Done processing %s. Removed %d files.\n", cp.dir, cp.counter)
		}

		cp.done <- cp.counter
	}()
}

func (cp *cleaningPath) processFiles(path string, info os.FileInfo, err error) error {
	if err == nil && (info.Mode()&os.ModeType == 0) && (strings.HasSuffix(path, ".php") ||
		strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js")) {
		if cp.test {
			fmt.Println("Removing " + path)
		} else {
			err := os.Remove(path)
			if err != nil {
				if cp.ignore {
					fmt.Fprintln(os.Stderr, err)
					return nil
				} else {
					return err
				}
			}
		}
		cp.incCounter()
	} else if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (cp *cleaningPath) processExpiredFiles(path string, info os.FileInfo, err error) error {
	if err == nil && (info.Mode()&os.ModeType == 0) && info.Size() > 100 && strings.HasSuffix(path, ".php") {
		return cp.processExpiredFile(path)
	} else if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (cp *cleaningPath) processExpiredFile(path string) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	lineNo := 0

	for {
		line, err := reader.ReadString('\n')
		lineNo++

		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, "failed to finish reading the file:", err)
			}
			break
		}

		if lineNo == 4 {
			match := cp.regs.FindStringSubmatch(line)

			if match != nil {
				tm, err := strconv.ParseInt(match[1], 10, 0)
				if err == nil {
					if tm < cp.tmNow {
						if cp.test {
							fmt.Println("Removing " + path)
						} else {
							err := os.Remove(path)
							if err != nil {
								if cp.ignore {
									fmt.Fprintln(os.Stderr, err)
									return nil
								} else {
									return err
								}
							}
						}
						cp.incCounter()
					}
				}
				break
			}
		}
	}
	return nil
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	path, err := os.Getwd()
	all := false
	test := false
	ignore := false
	dirsExp := []string{"managed_cache", "stack_cache", "cache"}
	dirsAll := []string{"managed_cache", "stack_cache", "cache", "html_pages"}

	flag.StringVar(&path, "path", path, "Path to bitrix root")
	flag.BoolVar(&all, "all", all, "Process all files (if not provided then the expired files will be processed only)")
	flag.BoolVar(&test, "test", test, "Do not remove files. Run for testing")
	flag.BoolVar(&ignore, "ignore", ignore, "Do not stop processing on errors")
	flag.Parse()
	path = strings.TrimSuffix(path, string(os.PathSeparator))

	fileInfo, err := os.Stat(path + string(os.PathSeparator) + "bitrix")
	if err != nil {
		if os.IsNotExist(err) {
			if path == "" {
				path = "Current directory"
			}
			fmt.Fprintln(os.Stderr, path+" is not bitrix root. Use -h for help.")
		} else if os.IsPermission(err) {
			fmt.Fprintln(os.Stderr, "Permission denied.")
		}

	} else if !fileInfo.IsDir() {
		fmt.Fprintln(os.Stderr, path+string(os.PathSeparator)+"bitrix is not directory.")
	} else {
		dirs := dirsExp
		if all {
			dirs = dirsAll
		}

		done := make(chan int, len(dirs))

		for _, dir := range dirs {
			prc := NewCleaningPath(path+string(os.PathSeparator)+"bitrix"+string(os.PathSeparator)+dir, done, all, test, ignore)
			prc.ProcessDir()
		}
		waitUntil(done, len(dirs))
	}
}

func waitUntil(done <-chan int, len int) {
	realCounter := 0
	for i := 0; i < len; i++ {
		realCounter = realCounter + (<-done)
	}
	fmt.Printf("Total removed %d files.\n", realCounter)
}
