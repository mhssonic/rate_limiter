package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

var mu sync.Mutex

func sendHttp(clientLog *log.Logger) {
	requestURL := fmt.Sprintf("http://localhost:%d/home", 3333)
	//var wg sync.WaitGroup
	//wg.Add(4)
	//req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	//if err != nil {
	//	fmt.Printf("client: could not create request: %s\n", err)
	//	os.Exit(1)
	//}

	//clientLog.Println("hey")
	//clients := make([]http.Client, 20)

	for i := 0; i < 1000; i++ {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConns = 100
		t.MaxConnsPerHost = 100
		t.MaxIdleConnsPerHost = 100

		httpClient := &http.Client{
			Timeout:   10 * time.Second,
			Transport: t,
		}
		go func() {
			//defer wg.Done()
			for {
				httpClient.Get(requestURL)
				//time.Sleep(1 * time.Microsecond)

				//client.Do(req)
				//clients[i].Do(req)
				//mu.Lock()
				//_, err := client.Do(req)
				//if err != nil {
				//	clientLog.Printf("client: error making http request: %s\n", err)
				//	os.Exit(1)
				//} /**else {
				//clientLog.Println("Good")
				//}**/
				//mu.Unlock()
			}
			//runtime.Gosched()
		}()
	}
	//wg.Wait()
	time.Sleep(2 * time.Minute)
}

func main() {
	fileClientLog, err := openLogFile("client/clientLog.log")
	if err != nil {
		log.Fatal(err)
	}
	clientLog := log.New(fileClientLog, "[client]", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
	sendHttp(clientLog)
	//sendHttp2(clientLog)

}
