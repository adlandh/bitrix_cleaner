package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var all = false
var test = false
var done chan struct{}


func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	path, err := os.Getwd()
	all = false
	test = false
	dirs := []string{"managed_cache", "stack_cache", "cache", "html_pages"}
	done = make(chan struct{}, len(dirs))

	flag.StringVar(&path, "path", path, "Path to bitrix root")
	flag.BoolVar(&all, "all", all, "Process all files (if not provided then the expired files will be processed only)")
	flag.BoolVar(&test, "donotremove", test, "Do not remove files. Run for testing.")
	flag.Parse()
	path = strings.TrimSuffix(path, string(os.PathSeparator))

	fileInfo, err := os.Stat(path + string(os.PathSeparator) + "bitrix")
	if err != nil {
		if os.IsNotExist(err) {
			if path == "" {
				path = "Current directory"
			}
			fmt.Fprint(os.Stderr, path+" is not bitrix root. Use -h for help.")
		} else if os.IsPermission(err) {
			fmt.Fprint(os.Stderr, "Permission denied.")
		}

	} else if !fileInfo.IsDir() {
		fmt.Fprint(os.Stderr, path+string(os.PathSeparator)+"bitrix is not directory.")
	} else {
		for _, dir := range dirs {
			go processDir(path+string(os.PathSeparator)+"bitrix"+string(os.PathSeparator)+dir)
		}
		waitUntil(done, len(dirs))
	}
}

func waitUntil(done <-chan struct{}, len int) {
	for i := 0; i < len; i++ {
		<-done
	}
}

func processDir(dir string) {
	_, err := os.Stat(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error processing "+dir)
	} else {
		fmt.Println("Start processing " + dir)
		if all {
			err = filepath.Walk(dir, processFiles)
		} else {
			err = filepath.Walk(dir, processExpiredFiles)
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		fmt.Println("Done processing " + dir)

		done <- struct{}{}
	}
}

func processFiles(path string, info os.FileInfo, err error) error {
	if err == nil && !info.IsDir() && strings.HasSuffix(path, ".php") {
		if test {
			fmt.Println("Removing "+path)
		} else {
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func processExpiredFiles(path string, info os.FileInfo, err error) error {
	tmnow := time.Now().Unix()
	if err == nil && !info.IsDir() && strings.HasSuffix(path, ".php") {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		reader := bufio.NewReader(file)

		regs := regexp.MustCompile(`dateexpire = '(\d+)'`)

		for {
			line, err := reader.ReadString('\n')

			if err != nil {
				if err != io.EOF {
					fmt.Fprintln(os.Stderr, "failed to finish reading the file:", err)
				}
				break
			}

			match := regs.FindStringSubmatch(line)

			if match != nil {
				tm, err := strconv.Atoi(match[1])
				if err == nil {
					if int64(tm) < tmnow {
						if test {
							fmt.Println("Removing "+path)
						} else {
							err := os.Remove(path)
							if err != nil {
								return err
							}
						}
					}
				}
				break
			}

		}
	}
	return err
}
