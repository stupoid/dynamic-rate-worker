package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

var logger *slog.Logger
var taskCount int

func process(job int64) {
	logger.Info("start", "job", job, "tasks", taskCount)
	time.Sleep(3 * time.Second)
	logger.Info("completed", "job", job, "tasks", taskCount)
}

func main() {
	logger = slog.New(slog.Default().Handler())
	rate := 1 * time.Second
	ticker := time.NewTicker(rate)
	wg := sync.WaitGroup{}
	var job int64

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		interrupt <- syscall.SIGINT
	})
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		payload, err := json.Marshal(map[string]interface{}{
			"running_tasks": taskCount,
			"goroutines":    runtime.NumGoroutine(),
			"handled_jobs":  job,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(payload)
	})
	http.HandleFunc("/up", func(w http.ResponseWriter, r *http.Request) {
		rate /= 2
		ticker.Reset(rate)
		payload, err := json.Marshal(map[string]interface{}{
			"rate": fmt.Sprintf("%s", rate),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(payload)
	})
	http.HandleFunc("/down", func(w http.ResponseWriter, r *http.Request) {
		rate *= 2
		ticker.Reset(rate)
		payload, err := json.Marshal(map[string]interface{}{
			"rate": fmt.Sprintf("%s", rate),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(payload)
	})
	go func() {
		logger.Info("Waiting for interrupt on http://localhost:8090/stop")
		http.ListenAndServe(":8090", nil)
	}()

	go func() {
		for {
			select {
			case <-ticker.C:
				wg.Add(1)
				taskCount++
				job++
				go func() {
					defer wg.Done()
					defer func() { taskCount-- }()
					process(job)
				}()
			}
		}
	}()

	<-interrupt
	logger.Info("Stopping ticker")
	ticker.Stop()
	logger.Info("Draining", "tasks", taskCount)
	wg.Wait()
}
