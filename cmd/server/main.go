package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/diggiex47/distributed-task-queue/internal/config"
	"github.com/diggiex47/distributed-task-queue/internal/job"
	"github.com/diggiex47/distributed-task-queue/internal/redisstore"
	"github.com/diggiex47/distributed-task-queue/internal/worker"
)



func main() {

	//load config and connect to redis
	cfg := config.Load()
	store, err:= redisstore.New(cfg.RedisAddr, cfg.RedisPass, cfg.QueueName)
	if err != nil {
		 log.Fatalf("Failed to connect to redis : %v", err)
	}

	defer store.Close()
	fmt.Println("connected to redis")

	// create a cancellable context.
	// when we call cancel(), ecery worker goroutine will see it and exit cleanly after finishing its current job.
	ctx, cancel := context.WithCancel(context.Background())

	// start the worker pool - this lanches 5 goroutines immediately, all them blocking BRPOP, waiting for jobs to appear
	pool := worker.New(store, cfg.WorkerCount)
	pool.Start(ctx)

	//give worker a moment to start up and log their "ready" messages 
	time.Sleep(500 * time.Millisecond)


	// push 5 jobs into the queue - watch the worker race to grab them 
	fmt.Println("\n --- pushing 5 jobs into the queue ---")
	for i := 1; i<= 5; i++ {
		now := time.Now()
		j := &job.Job {
			ID:         fmt.Sprintf("job-%d", i),
			Status:     job.StatusPending,
			Payload:    fmt.Sprintf("task-number-%d", i),
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		if err := store.EnqueueJob(ctx, j); err != nil {
			log.Printf("failed to enqueue job %d: %v", i, err)
			continue		
		}
		fmt.Printf("Enqueued: %s\n", j.ID)
	}


	// wait for all jobs to be processed
	// each job tadkles 3 seconds. with 5 workers running in parallel,
	// all 5 jobvs should complere in ~2 seconds total - not 10 sec
	// that the power of concurrent processing.
	fmt.Println("\n --- waiting for worker to process all jobs ---")
	time.Sleep(5 * time.Second)

	// check the final status of every job 
	fmt.Println("\n --- final job status ---")
	for i := 1; i <= 5; i++ {
		jobID := fmt.Sprintf("job-%d", i)
		j, err := store.GetJob(ctx, jobID)
		if err != nil || j == nil {
			fmt.Printf("%s: not found\n", jobID)
			continue
		}
		fmt.Printf("%s -> status: %s | result : %s\n", j.ID , j.Status, j.Result)
	}


	// signal all worker to stop and wait for them to finish cleanly
	fmt.Println("\n --- shutting down worker ---")
	cancel()      // send stop signal to all worker goroutines 
	pool.Wait()   // block until every goroutine has exited
	fmt.Println("All worker stopped. shutdown complete.")

}