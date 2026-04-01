package main

import (
	"fmt"
	"time"

	"github.com/diggiex47/distributed-task-queue/internal/config"
	"github.com/diggiex47/distributed-task-queue/internal/job"
)

func main() {

	cfg := config.Load()


	fmt.Println("=== Configuration loaded ===")
	fmt.Printf("Redis Address: %s\n", cfg.RedisAddr)
	fmt.Printf("Queue Name: %s\n", cfg.QueueName)
	fmt.Printf("Worker Count: %d\n", cfg.WorkerCount)
	fmt.Printf("API Port: %s\n", cfg.APIPort)


	fmt.Println("\n=== Simulating a job lifecycle ===")
	j := job.Job{
		ID: "test-1",
		Status: job.StatusPending,
		Payload: "send-welcome-email",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}


	fmt.Printf("Job created -> status: %s\n", j.Status)

	j.Status = job.StatusProcessing
	fmt.Printf("Worker picked up -> Status: %s\n", j.Status)

	// Simulate the worker finished successfully
	j.Status = job.StatusCompleted
	j.Result = " Email sent to user@example.com"
	j.UpdatedAt = time.Now()


	fmt.Println("\n=== job completed ===")
	fmt.Printf("staus: %s\n", j.Status)
	fmt.Printf("Result: %s\n", j.Result)
	fmt.Printf("Is complete? %v\n", j.IsComplete())
}