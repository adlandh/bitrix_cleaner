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
	"io/ioutil"
)

var all = false
var test = false
var done chan struct{}
var regs *regexp.Regexp
var tmNow int64
var removed int

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	path, err := os.Getwd()
	all = false
	test = false
	dirsExp := []string{"managed_cache", "stack_cache", "cache"}
	dirsAll := []string{"managed_cache", "stack_cache", "cache", "html_pages"}
	regs = regexp.MustCompile(`dateexpire = '(\d+)'`)
	tmNow = time.Now().Unix()

	defer fmt.Printf("Removed %d files.\n",removed)

	flag.StringVar(&path, "path", path, "Path to bitrix root")
	flag.BoolVar(&all, "all", all, "Process all files (if not provided then the expired files will be processed only)")
	flag.BoolVar(&test, "test", test, "Do not remove files. Run for testing.")
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

		done = make(chan struct{}, len(dirs))

		for _, dir := range dirs {
			go processDir(path + string(os.PathSeparator) + "bitrix" + string(os.PathSeparator) + dir)
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
	}

	done <- struct{}{}

}

func processFiles(path string, info os.FileInfo, err error) error {
	if err == nil && !info.IsDir() && (strings.HasSuffix(path, ".php") || strings.HasSuffix(path, ".html") ||
		strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js")) {
		if test {
			fmt.Println("Removing " + path)
		} else {
			err := os.Remove(path)
			if err != nil {
				return err
			}
			removed++
		}
	} else if os.IsNotExist(err){
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return err
}

func processExpiredFiles(path string, info os.FileInfo, err error) error {
	if err == nil && !info.IsDir() && strings.HasSuffix(path, ".php") {
		return processExpiredFile(path)
	} else if os.IsNotExist(err){
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return err
}

func processExpiredFile(path string) error {

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

		if lineNo==4 {
			match := regs.FindStringSubmatch(line)

			if match != nil {
				tm, err := strconv.ParseInt(match[1], 10, 0)
				if err == nil {
					if tm < tmNow {
						if test {
							fmt.Println("Removing " + path)
						} else {
							err := os.Remove(path)
							if err != nil {
								return err
							}
							removed++
						}
					}
				}
				break
			}
		}
	}
	return nil
}
