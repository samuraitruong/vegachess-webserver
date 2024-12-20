package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Repository URL - This should be set from environment variable or passed as input
var REPO_URL = os.Getenv("REPO_URL")
var DRYRUN = os.Getenv("DRYRUN") == "true"
// Replace with actual URL or use env variable
var REPO_FOLDER = "tournaments"

func main() {
	// Start the HTTP server in a separate goroutine
	go startHTTPServer()

	// Read DELAY_TIME from environment variable
	DELAY_TIME, err := strconv.Atoi(os.Getenv("DELAY_TIME"))
	if err != nil {
		// Log the error if DELAY_TIME is not set or invalid
		fmt.Println("Invalid DELAY_TIME value, using default 10 seconds.")
		DELAY_TIME = 10
	}
	var debounceTime = time.Duration(DELAY_TIME) * time.Second

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Set the FTP root folder
	ftpRoot := "/data"
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
			// Ensure the tournaments directory exists, clone repo if needed
			ensureRepoClone(REPO_FOLDER)
			// Copy files from /data to /www/tournaments/www
			copyFiles(ftpRoot, REPO_FOLDER+"/www")
			// Perform git push
			gitPush(DRYRUN)
		})
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// fmt.Println("Event:", event)
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

// Starts a simple HTTP server
func startHTTPServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Watcher is running")
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	port := 8080 // Default HTTP port
	fmt.Printf("HTTP server is running on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

// ensureRepoClone checks if the repo exists, clones it if not
func ensureRepoClone(directory string) {
	// Check if the tournaments directory exists
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		// Directory doesn't exist, clone the repository
		fmt.Println("Cloning repository...")
		cmd := exec.Command("git", "clone", "--depth", "1", REPO_URL, directory)
		output, err := cmd.CombinedOutput() // Capture stdout and stderr
		fmt.Printf("Running command: git clone --depth 1 %s %s\n", REPO_URL, directory)
		fmt.Printf("Command output: %s\n", output)
		if err != nil {
			log.Printf("Error cloning repository: %v\nOutput: %s\n", err, output)
			return
		}
		fmt.Println("Repository cloned.")
	} else {
		// Directory exists, pulling changes if necessary
		fmt.Println("Repository already exists, pulling changes...")
		cmd := exec.Command("git", "pull")
		cmd.Dir = directory // Set working directory to /www/tournaments
		output, err := cmd.CombinedOutput()
		fmt.Printf("Running command: git pull\n")
		fmt.Printf("Command output: %s\n", output)
		if err != nil {
			log.Printf("Error pulling repository changes: %v\nOutput: %s\n", err, output)
			return
		}
		fmt.Println("Repository updated.")
	}
}

// copyFiles copies all files and directories from source to destination
func copyFiles(srcDir, dstDir string) {
	fmt.Println("Starting file copy from", srcDir, "to", dstDir)

	// Use sh -c to allow wildcard expansion through the shell
	cmd := exec.Command("sh", "-c", fmt.Sprintf("cp -R %s/* %s", srcDir, dstDir))
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("Error copying files from %s to %s: %v\nOutput: %s\n", srcDir, dstDir, err, output)
	} else {
		fmt.Println("Files copied successfully from", srcDir, "to", dstDir)
	}
}

// gitPush runs a git push command to push changes to the remote repository
func gitPush(dryrun bool) {
	cmd := exec.Command("git", "status")
	cmd.Dir = REPO_FOLDER
	output, err := cmd.CombinedOutput()
	fmt.Printf("Running command: git status\n")
	fmt.Printf("Command output: %s\n", output)
	if err != nil {
		log.Printf("Error running git status: %v\nOutput: %s\n", err, output)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = REPO_FOLDER
	output, err = cmd.CombinedOutput()
	fmt.Printf("Running command: git add .\n")
	fmt.Printf("Command output: %s\n", output)
	if err != nil {
		log.Printf("Error running git add: %v\nOutput: %s\n", err, output)
	}

	commitMessage := "Vega Publish: Update vega publish html from clients"
	cmd = exec.Command("git", "commit", "-m", commitMessage)
	cmd.Dir = REPO_FOLDER
	output, err = cmd.CombinedOutput()
	fmt.Printf("Running command: git commit -m \"%s\"\n", commitMessage)
	fmt.Printf("Command output: %s\n", output)
	if err != nil {
		log.Printf("Error running git commit: %v\nOutput: %s\n", err, output)
	}

	if dryrun {
		fmt.Println("Dry run enabled. Skipping git push.")
	} else {
		cmd = exec.Command("git", "push")
		cmd.Dir = REPO_FOLDER
		output, err = cmd.CombinedOutput()
		fmt.Printf("Running command: git push\n")
		fmt.Printf("Command output: %s\n", output)
		if err != nil {
			log.Printf("Error running git push: %v\nOutput: %s\n", err, output)
		}
	}

	fmt.Println("Git commit (and push if not dryrun) completed.")
}
