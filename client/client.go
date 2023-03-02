package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func main() {
	fileClientLog, err := openLogFile("client/clientLog.log")
	if err != nil {
		log.Fatal(err)
	}
	clientLog := log.New(fileClientLog, "[client]", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)

	requestURL := fmt.Sprintf("http://localhost:%d/home", 3333)
	reg, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		clientLog.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	for true {
		res, err := client.Do(reg)
		if err != nil {
			clientLog.Printf("client: error making http request: %s\n", err)
			os.Exit(1)
		}
		clientLog.Printf("client: got response!\n")
		clientLog.Printf("client: status code: %d\n", res.StatusCode)
		time.Sleep(20 * time.Millisecond)
	}
}
