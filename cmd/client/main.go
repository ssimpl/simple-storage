package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	filePath := flag.String("file", "./testdata/funny_cats.mp4", "file path")
	serverAddr := flag.String("server", "http://localhost:8080", "server address")

	flag.Parse()

	log.Println("Client started")

	file, err := os.Open(*filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	log.Println("File size:", stat.Size())

	fileURL := fmt.Sprintf("%s/%s", *serverAddr, stat.Name())

	log.Println("File URL:", fileURL)

	req, err := http.NewRequest(http.MethodPut, fileURL, file)
	if err != nil {
		return err
	}

	startTime := time.Now()
	log.Println("Uploading file...", "timeStart", startTime.Format(time.RFC3339Nano))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println("Upload finished", "duration", time.Since(startTime).String())

	if resp.StatusCode != http.StatusOK {
		return err
	}

	req, err = http.NewRequest(http.MethodGet, fileURL, nil)
	if err != nil {
		return err
	}

	startTime = time.Now()
	log.Println("Downloading file...", "timeStart", startTime.Format(time.RFC3339Nano))

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Println("Download finished", "duration", time.Since(startTime).String())

	log.Println("Comparing files...")

	if md5.Sum(fileData) != md5.Sum(body) {
		log.Println("Files are not equal")
	} else {
		log.Println("Files are equal")
	}

	log.Println("Client finished")

	return nil
}
