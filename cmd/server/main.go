package main

import (
	"fmt"
	"time"
	"log"
	"context"

	"github.com/diggiex47/distributed-task-queue/internal/config"
	"github.com/diggiex47/distributed-task-queue/internal/job"
	"github.com/diggiex47/distributed-task-queue/internal/redisstore"
)



func main() {

	cfg := config.Load()
	fmt.Println("=== Configuration loaded ===")
	fmt.Printf("Connecting to redis at %s\n", cfg.RedisAddr)

	store, err := redisstore.New(cfg.RedisAddr, cfg.RedisPass, cfg.QueueName)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer store.Close()
	fmt.Println("connected to redis successfully")

	ctx := context.Background()




	// create and enqueue a job
	now := time.Now()
	j := &job.Job{
		ID: "test-job-001",
		Status: job.StatusPending,
		Payload: "send-welcome-email",
		CreatedAt: now,
		UpdatedAt: now,
	}

	fmt.Printf("\n === Enquering job: %s ===\n", j.ID)
	if err := store.EnqueueJob(ctx, j); err != nil {
		log.Fatalf("Failed to enqueue job: %v", err)
	}
	fmt.Println("Job enquered successfully")




	//Reading job from redis

	fmt.Printf("\n=== Reading job back from redis ===\n")
	retrieved, err := store.GetJob(ctx, j.ID)
	if err != nil {
		log.Fatalf("failed to get job: %v", err)
	}
	fmt.Printf("ID: %s\n", retrieved.ID)
	fmt.Printf("Status: %s\n", retrieved.Status)
	fmt.Printf("Payload: %s\n", retrieved.Payload)




	//simulate worker updating the status 
	fmt.Printf("\n=== simulating worker picking up the job ===\n")
	err = store.UpdateJobStatus(ctx, j.ID, job.StatusProcessing, "", "")
	if err != nil {
		log.Fatalf("failed to update status: %v", err)
	}

	// Read it again to confirm the status changed in redis
	updated, _ := store.GetJob(ctx, j.ID)
	fmt.Printf("status is now: %s\n", updated.Status)



	//simulate worker completing the job 
	fmt.Printf("\n=== simulating job completion ===\n")
err = store.UpdateJobStatus(ctx, j.ID, job.StatusCompleted, "Emial sent", "")
if err != nil {
	log.Fatalf("Failed to update status: %v", err)
}

completed, _ := store.GetJob(ctx, j.ID)
fmt.Printf("Status: %s\n", completed.Status)
fmt.Printf("Result: %s\n", completed.Result)
fmt.Println("\n Phase 3 completed successfully")
}
