package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"runtime"
	"path/filepath"
	//"bufio"
)

var  files chan string

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	path,err:=os.Getwd()
	all:=false
	dirs:=[]string{"managed_cache","stack_cache","cache","html_pages"}
	done:=make(chan error,len(dirs))

	flag.StringVar(&path, "path", path, "Path to bitrix root")
	flag.BoolVar(&all, "all", false,"Process all files (if not provided then the expired files will be processed only)")
	flag.Parse()
	path = strings.TrimSuffix(path, string(os.PathSeparator))

	fileInfo, err := os.Stat(path + string(os.PathSeparator)+"bitrix")
	if err != nil {
		if os.IsNotExist(err) {
			if path == "" {
				path = "Current directory"
			}
			done<-fmt.Errorf(path + " is not bitrix root. Use -h for help.")
		} else if os.IsPermission(err) {
			done<-fmt.Errorf("Permission denied.")
		}

	} else if !fileInfo.IsDir() {
		done<-fmt.Errorf(path + string(os.PathSeparator)+"bitrix is not directory.")
	} else {
		for _,dir := range dirs {
			go processDir(path+string(os.PathSeparator)+"bitrix"+string(os.PathSeparator)+dir, all, done)
		}
	}

	waitUntil(done,len(dirs))
}


func waitUntil(done <-chan error, len int) {
	for i:=0; i < len; i++ {
		err:=<-done
		if err!= nil {
			fmt.Fprintln(os.Stderr,err)
			i = len
		}
	}
}

func processDir(dir string, all bool, done chan<- error) {
	_,err:=os.Stat(dir)
	if err != nil {
		done<-fmt.Errorf("Error processing "+dir)
	} else {
		files = make(chan string)
		fmt.Println("Start processing "+dir)
		go processFiles(all)
		err=filepath.Walk(dir,listFiles)
		if err != nil {
			done<-err
		} else {
			fmt.Println("Done processing "+dir)
			done <- nil
		}
	}
}

func listFiles(path string, info os.FileInfo, err error) error {
	if err == nil && !info.IsDir() && strings.HasSuffix(path,".php") {
		files <- path
	}
	return err
}

func processFiles(all bool) {
	if all {
		for file := range files {
			os.Remove(file)
		}
	} else {
		for file := range files {
			/* file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		reader := bufio.NewReader(file) */
		fmt.Println(file)
		}

	}

}