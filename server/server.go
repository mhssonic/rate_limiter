package main

import (
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// variables of counter in metric
var okStatusCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "ok_request_count",
		Help: "Number of 200",
	},
)
var tooManyRequestCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "too_many_request_count",
		Help: "Number of 429",
	},
)

var mu sync.Mutex
var redis = make(map[string][]int) //last element of every array is sum of the req of a min

func listener(serverLog *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//serverLog.Println("Got a request!")
		clientAddr := r.RemoteAddr
		second := time.Now().Second()
		mu.Lock()
		el, ok := redis[clientAddr]
		if ok {
			if el[60] > 60 { //todo database rule
				//serverLog.Println("not ok request")
				tooManyRequestCounter.Inc()
				w.WriteHeader(http.StatusTooManyRequests)
			} else {
				//serverLog.Println("ok request")
				okStatusCounter.Inc()
				redis[clientAddr][second]++
				redis[clientAddr][60]++
			}
		} else {
			redis[clientAddr] = make([]int, 61)
			//serverLog.Println("ok request")
			okStatusCounter.Inc()
			redis[clientAddr][second] = 1
			redis[clientAddr][60] = 1
		}
		mu.Unlock()
		//serverLog.Printf("%v\n", redis[clientAddr])
		//fmt.Printf("server: %s /\n", r.Method)
		//fmt.Printf("server: query id: %s\n", r.URL.Query().Get("id"))
		//fmt.Printf("server: content-type: %s\n", r.Header.Get("content-type"))
		//fmt.Printf("server: ip: %s\n", r.RemoteAddr)
		//fmt.Printf("server: headers:\n")
		//for headerName, headerValue := range r.Header {
		//	fmt.Printf("\t%s = %s\n", headerName, strings.Join(headerValue, ", "))
		//}
		//reqBody, err := ioutil.ReadAll(r.Body)
		//if err != nil {
		//	fmt.Printf("server: could not read request body")
		//}
		//fmt.Printf("server: request body %s\n", reqBody)
		//fmt.Fprintf(w, `{"message": "hello!"}`)
	}
}

func main() {
	//metric
	prometheus.MustRegister(tooManyRequestCounter, okStatusCounter)

	//log handling
	fileServerLog, err := openLogFile("server/serverLog.log")
	if err != nil {
		log.Fatal(err)
	}
	serverLog := log.New(fileServerLog, "[server]", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)

	var wg sync.WaitGroup
	wg.Add(1)

	//renew redis every second
	flagSecond := time.Now().Second()
	go func() {
		for true {
			if flagSecond != time.Now().Second() {
				mu.Lock()
				flagSecond = time.Now().Second()
				for key, el := range redis {
					redis[key][60] -= el[flagSecond]
					redis[key][flagSecond] = 0
				}
				mu.Unlock()
			}
		}
	}()

	//server:
	go func() {
		defer wg.Done()
		mux := http.NewServeMux()
		mux.HandleFunc("/home", listener(serverLog))
		mux.Handle("/metrics", promhttp.Handler())
		server := http.Server{
			Addr:    fmt.Sprintf(":%d", 3333),
			Handler: mux,
		}
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				serverLog.Printf("error running http server: %s\n", err)
			}
		}
	}()
	wg.Wait()
}
