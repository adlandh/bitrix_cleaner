package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"runtime"
	"path/filepath"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	path,err:=os.Getwd()
	all:=false
	dirs:=[]string{"managed_cache","stack_cache","cache","html_pages"}
	done:=make(chan struct{},len(dirs))

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
			fmt.Fprintln(os.Stderr,path + " is not bitrix root. Use -h for help.")
			os.Exit(1)
		} else if os.IsPermission(err) {
			fmt.Println("Permission denied.")
			os.Exit(1)
		}

	} else if !fileInfo.IsDir() {
		fmt.Fprintln(os.Stderr,path + string(os.PathSeparator)+"bitrix is not directory.")
		os.Exit(1)
	}

	for _,dir := range dirs {
		go processDir(path+string(os.PathSeparator)+"bitrix"+string(os.PathSeparator)+dir, all, done)
	}

	waitUntil(done,len(dirs))

}


func waitUntil(done <-chan struct{}, len int) {
	for i := 0; i < len; i++ {
		<-done
	}
}

func processDir(dir string, all bool, done chan<- struct{}) {
	_,err:=os.Stat(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr,"Error processing "+dir)
	} else {
		fmt.Println("Start processing "+dir)

		if all {
			err=filepath.Walk(dir,processFile)
		} else {
			err=filepath.Walk(dir,processExpiredFile)
		}

		if err != nil {
			fmt.Fprintln(os.Stderr,err)
			fmt.Fprintln(os.Stderr,"Error processing "+dir)
		} else {
			fmt.Println("Done processing "+dir)
		}
	}
	done <- struct{}{}
}

func processFile(path string, info os.FileInfo, err error) error {
	if !info.IsDir() && strings.HasSuffix(path,".php") && err == nil {
		return os.Remove(path)
	} else {
		return err
	}
}

func processExpiredFile(path string, info os.FileInfo, err error) error {
	fmt.Println(path)
	fmt.Println(info.Size())
	fmt.Println(info.IsDir())
	return err
}