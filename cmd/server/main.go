package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diggiex47/distributed-task-queue/internal/api"
	"github.com/diggiex47/distributed-task-queue/internal/config"
	"github.com/diggiex47/distributed-task-queue/internal/redisstore"
	"github.com/diggiex47/distributed-task-queue/internal/worker"
)

func main() {
	// 1. Load Config
	cfg := config.Load()

	// 2. Connect to redis
	store, err := redisstore.New(cfg.RedisAddr, cfg.RedisPass, cfg.QueueName)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer store.Close()
	log.Println("Connected to Redis")

	// 3 Create cance;llable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4 start worker pool
	pool := worker.New(store, cfg.WorkerCount)
	pool.Start(ctx)

	// 5 set up HTTP server
	mux := http.NewServeMux()
	handler := api.New(store)
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.APIPort),
		Handler:      mux,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// send HTTP server in its and it's sort of tough cown goroutine - ListenAndServe blocks
	// so running it in agoroutine lets the rest of main() continue
	go func() {
		log.Printf("API server listening on: %s", cfg.APIPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 6 wait for ctrl+c or docker stop signal
	// This Blocks until you press ctrl+c
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutdown signal received..")

	// 7. Graceful shutdown
	cancel()    // stop all worker
	pool.Wait() // wait for worker to finish current jobs
	log.Println("all worker stopped. shutdown complete")
}
