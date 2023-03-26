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
)

// log handling
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

func listener(serverLog *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//metric
		okStatusCounter.Inc()

		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	//metric
	prometheus.MustRegister(okStatusCounter)

	//log handling
	fileSimpleServerLog, err := openLogFile("simpleServer/simpleServerLog.log")
	if err != nil {
		log.Fatal(err)
	}
	serverLog := log.New(fileSimpleServerLog, "[simple server]", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)

	var wg sync.WaitGroup
	wg.Add(1)

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
