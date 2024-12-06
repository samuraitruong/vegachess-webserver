package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounceTime = 2 * time.Second

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Set FTP root folder
	ftpRoot := "/home/ftpuser/ftp"
	err = watcher.Add(ftpRoot)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Watching for changes in:", ftpRoot)

	var timer *time.Timer
	debounce := func(event fsnotify.Event) {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounceTime, func() {
			fmt.Printf("Changes detected: %s\n", event)
			// Add your logic here
		})
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			fmt.Println("Event:", event)
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
				debounce(event)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Error:", err)
		}
	}
}
