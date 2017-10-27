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

	path:="."
	all:=false
	dirs:=[]string{"managed_cache","stack_cache","cache",`html_pages`}
	done:=make(chan struct{},len(dirs))

	flag.StringVar(&path, "path", ".", "Path to bitrix root")
	flag.BoolVar(&all, "all", false,"all files (if not provided then only expired will be processed)")
	flag.Parse()
	path = strings.TrimSuffix(path, "/")

	fileInfo, err := os.Stat(path + "/bitrix")
	if err != nil {
		if os.IsNotExist(err) {
			if path == "" {
				path = "Current directory"
			}
			fmt.Println(path + " is not bitrix root. Use -h for help.")
			os.Exit(1)
		} else if os.IsPermission(err) {
			fmt.Println("Permission error")
			os.Exit(1)
		}

	} else if !fileInfo.IsDir() {
		fmt.Println(path + "/bitrix is not directory.")
		os.Exit(1)
	}

	for _,dir := range dirs {
		go processDir(path+"/bitrix/"+dir, all, done)
	}

	fmt.Println(all)

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
		fmt.Println("Error processing "+dir)
	} else {
		if all {
			os.RemoveAll(dir)
			os.Mkdir(dir,0755)
		} else {
			filepath.Walk(dir,processExpiredFile)
		}


		fmt.Println("Done processing "+dir)
	}
	done <- struct{}{}
}

/* func processFile(path string, info os.FileInfo, err error) error {
	fmt.Println(path)
	fmt.Println(info.Size())
	fmt.Println(info.IsDir())
	return err
} */

func processExpiredFile(path string, info os.FileInfo, err error) error {
	fmt.Println(path)
	fmt.Println(info.Size())
	fmt.Println(info.IsDir())
	return err
}