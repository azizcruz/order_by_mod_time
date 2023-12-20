package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"sync"
)

type File struct {
	Entry       os.DirEntry
	UnixModTime int64
}

func main() {
	wg := &sync.WaitGroup{}
	filesChan := make(chan os.DirEntry, 100)
	listOfFiles := []File{}
	mutex := &sync.Mutex{}

	wg.Add(2)

	go filesLister(filesChan, wg)
	go filesCollector(filesChan, &listOfFiles, mutex, wg)

	fmt.Println("Number of Goroutines:", runtime.NumGoroutine())

	wg.Wait()

	// Sort by timestamp
	sort.Slice(listOfFiles, func(i, j int) bool {
		return listOfFiles[i].UnixModTime < listOfFiles[j].UnixModTime
	})

	var counter int = 1
	for _, file := range listOfFiles {
		defer fmt.Printf("Done renaming %s file\n", file.Entry.Name())
		if file.Entry.Name() != "main.go" && file.Entry.Name() != "go.mod" {
			newFileName := file.Entry.Name() + "-" + generateNumber(counter)
			counter++
			os.Rename(file.Entry.Name(), newFileName)
			fmt.Println(file.Entry.Name(), "->", newFileName)
		}

	}
}

func filesLister(ch chan os.DirEntry, wg *sync.WaitGroup) {
	files, err := os.ReadDir(".")
	if err != nil {
		panic(err)
	} else {
		for _, file := range files {
			ch <- file
		}
	}
	close(ch) // Close the channel when done sending files
	wg.Done() // Signal completion of filesLister
}

func filesCollector(ch chan os.DirEntry, listOfFiles *[]File, m *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	for file := range ch {
		wg.Add(1)
		unixModTime, _ := os.Stat(filepath.Join(file.Name()))
		go func(file os.DirEntry, timestamp fs.FileInfo) {
			defer wg.Done()
			m.Lock()
			*listOfFiles = append(*listOfFiles, File{
				Entry:       file,
				UnixModTime: timestamp.ModTime().Unix(),
			})
			m.Unlock()
		}(file, unixModTime)
	}
}

func generateNumber(num int) string {
	if num > 9 {
		return strconv.FormatInt(int64(num), 10)
	} else {
		return strconv.FormatInt(int64(0), 10) + strconv.FormatInt(int64(num), 10)
	}
}
