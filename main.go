package main

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {

	stat := flag.Bool("s", false, "stat file")
	recurse := flag.Bool("r", false, "recurse dirs")
	flag.Parse()

	// setup watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
	}(watcher)

	done := make(chan bool)
	printer := make(chan fsnotify.Event)
	// use goroutine to start the watcher
	go func() {
		ffs := os.DirFS(".")

		for {
			event := <-printer
			fmt.Printf("%-56s %v => %s", time.Now(), event.Name, event.Op.String())
			if *stat || *recurse {
				fi, err := fs.Stat(ffs, event.Name)
				if err != nil {
					fmt.Println()
					continue
				}
				if *recurse && fi.IsDir() {
					err = watcher.Add(event.Name)
					if err != nil {
						fmt.Println()
						continue
					}

				}
				if *stat {
					fmt.Printf(" %v %v %v", fi.Mode().String(), fi.Size(), fi.IsDir())
				}
			}
			fmt.Println("")
		}
	}()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// monitor only for write events
				go func() {
					printer <- event
				}()

			case err := <-watcher.Errors:
				log.Println("Error:", err)
			}
		}
	}()

	// provide the file name along with path to be watched
	if flag.NArg() == 0 {
		err = watcher.Add(".")
		if err != nil {
			log.Fatal(err)
		}

		err = filepath.WalkDir(".",
			func(path string, file fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if !file.IsDir() {
					return nil
				}

				return watcher.Add(path)
			})
		if err != nil {
			return
		}
	} else {
		for _, dir := range flag.Args() {
			err = watcher.Add(dir)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	<-done

}
