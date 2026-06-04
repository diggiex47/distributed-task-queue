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

	// 3 Create cancellable context for graceful shutdown
	// this is the master switch for the entire system.
	// Calling cancel() sends a stop signal to ecery goroutines watching ctx.
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
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// send HTTP server in its and it's sort of tough cown goroutine - ListenAndServe blocks
	// so running it in agoroutine lets the rest of main() continue to the shutdown logic below.
	go func() {
		log.Printf("API server listening on: %s", cfg.APIPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 6 wait for ctrl+c or docker stop signal
	// make(chan os.Signal, 1) creates a channel that carries OS signal
	// signal.Notify twlls Go: "when you receive SIGINT or SIGTERM, send it to the quit channel instead of killing the program"
	// SIGINT = ctrl+c 
	// SIGTERM = docker stop 
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)


	sig := <-quit // Blocks here until a signal arrivves
	log.Println("shutdown signal received..", sig)



	// 7. Graceful shutdown
	// step 1: stop the HTTP sercer gracefullly.
	// no new request will be accepted. existing requests get 30 sec
	// to complete beffore being forcefully closed
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to close: %v", err)
	}
	log.Println("HTTP server stopped")

	// step 2: signal all worker to stop.
	//cabncel() close the ctx.Donr() channel, which every worker is watching.
	// Their BRPOP calls return immediately.
	cancel()
	log.Println("waiting for the worker to finish current job...")

	//step 3: wiat fot every worker to finish its current job.
	// pool.Wait() blocks until the WaitGroup counter reaches zero.
	// no worker will be kulled nid-task.
	pool.Wait()
	log.Println("all workers stopped. exiting.")
}
