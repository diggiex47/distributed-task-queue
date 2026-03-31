package main

import (
	"fmt"
	"time"
	"github.com/diggiex47/distributed-task-queue/internal/job"
)

func main() {
	j := job.Job{
		ID: "test-1",
		Status: job.StatusPending,
		Payload: "send-welcome-email",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fmt.Println("=== Initial State ===")
	fmt.Printf("ID: %s\n", j.ID)
	fmt.Printf("Status; %s\n", j.Status)
	fmt.Printf("payload: %s\n", j.Payload)
	fmt.Printf("Is complete? %v\n", j.IsComplete())
	fmt.Printf("Age: %s\n", j.Age())


	j.Status = job.StatusProcessing
	fmt.Println("\n=== Worker picked it up ===")
	fmt.Printf("Status: %s\n", j.Status)

	// Simulate the worker finished successfully
	j.Status = job.StatusCompleted
	j.Result = " Email sent to user@example.com"
	j.UpdatedAt = time.Now()


	fmt.Println("\n=== job completed ===")
	fmt.Printf("staus: %s\n", j.Status)
	fmt.Printf("Result: %s\n", j.Result)
	fmt.Printf("Is complete? %v\n", j.IsComplete())
}