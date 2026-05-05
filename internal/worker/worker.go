package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/diggiex47/distributed-task-queue/internal/job"
	"github.com/diggiex47/distributed-task-queue/internal/redisstore"
)

// pool managess a fixed number of worker goroutines
// In Go: a struct  with a method to start worker and a method to wait for them all to finish

type Pool struct {
	store   *redisstore.Store
	workerCount int
	wg  sync.WaitGroup // scoreboard  - tracks running goroutines 
}


// new creates a pool. dosen't start any goroutines yet - just stores the setting 
func New(store *redisstore.Store, workerCount int) *Pool {
	return &Pool {
		store: store,
		workerCount: workerCount,
	}
}


// start launcehs workerCount goroutines. each one runs independently and concutrrently - they dont't wait for each other

func (p *Pool) Start (ctx context.Context) {
	log.Printf("Starting %d workers...", p.workerCount)

	for i := 0; i< p.workerCount; i++ {
		p.wg.Add(1) // tell the scoreboard; one more goroutine is starting

		workerID := i + 1 //give each worker a human-redable ID for logs



		//"go func() launches this function as goroutine - it runs concurrently alongside all the other immediately"
		// imp : we pass workerID as a parrameter to the goroutines function - if u used workerID directly inside the goroutines without passing it, all gorou tines would share the same variabvle and likely printr the same ID.
		go func (id int) {
			defer p.wg.Done() // when this goroutines exit, decrement scoreboard 
			p.runWorker(ctx, id)
		} (workerID)
	}
}

//wait block until every worker goroutines has finished.
//call this during shutdown - it ensure no worker is mid-task when the progeam exit
func (p *Pool) Wait() {
	p.wg.Wait()
}

// runWorker is the main loop for a single worker goroutines
// it runs forever, pickinh up one job at a time, until the context is cancelled 
func ( p *Pool) runWorker(ctx context.Context, workerID int) {
	log.Printf("worker %d: started aand waiting for jobs", workerID)
	for {
		// before trying to dequeue check if we've been told to stop
		//select in go if like a switch satatement but for channels.
		//check - this specific pattern with a default case - is non blocking. it check ctx.Done() instantly and moves on if there no shutdown signal.

		select {
		case <-ctx.Done():
			log.Printf("worker %d: shutdown signal received, stopping",workerID)
			return 
		default:
			// no shutdown signal - continue to dequeue
		}

		//BRPOP blocks here until a job arrrives in the queue.
		//The gotroutines is [arker at the network level - zero CPU used while waitin. Redis wakes it up the instant a job id pushed.
		

		// If ctx is cancelled while blocking jere BRPOP returns 
		// immediattly with a nil job - which we handle below.
		j, err := p.store.DequeueJob(ctx)
		if err != nil {
			log.Printf("worker %d: error deqeuing job: %v", workerID, err) 
			continue  // log the error and try again 	
	}
	if j ==  nil {
		 // this happens when contxt was cancelld while waiting 
		 // it means we're shutting down. Exit the loop cleanly.
		 return 
}

 //we have a job - process it 
 p.processJob(ctx, workerID, j)

}

}


// processJob handles a single job from start to finish:
// mark it processing -> do the work -> mark it completed or failed 

func (p *Pool) processJob(ctx context.Context, workerID int, j *job.Job) {
	log.Printf("worker %d: picked up job %s (payload: %s)", workerID, j.ID, j.Payload)


	// step 1: mark as processing so that the API can report it to clients
	err := p.store.UpdateJobStatus(ctx, j.ID, job.StatusProcessing, "", "")
	if err != nil {
		log.Printf("worker %d: failed to mark job %s as processing: %v", workerID, j.ID, err)
		return
	}

	// step 2: do the actual work 
	// In real system this would parse the payload and call teal business
	// Logic - send email, resize an image etx
	// for now we simulate work with a 2 second sleep
	result, err := doWork(ctx, j)
	
	// step 3: update the final status based on outcome
	if err != nil {
		log.Printf("worker %d: job %s failed: %v", workerID, j.ID, err) 
		_ = p.store.UpdateJobStatus(ctx, j.ID, job.StatusFailed, "", err.Error())
		return
	}


	log.Printf("worker %d: job %s Completed - result: %s", workerID, j.ID, result) 
	_ = p.store.UpdateJobStatus(ctx, j.ID, job.StatusCompleted, result, "")
}


// doWork simulates processsing a task 
// replace this in a real system with actual business logic
// notice it accepts ctx - if shutdown is requested mid-work, the select case <-ctx.Done() triggers and we return an error immediately instead oof finishing the fake work. this is how go handles cancellation deep inside a call chain.

func doWork(ctx context.Context, j *job.Job) (string, error) {
	select {
	case <-time.After(2 * time.Second): // simulate 2 sec of work 
	return fmt.Sprintf("processed '%s' successfully", j.Payload), nil
	case <-ctx.Done():
		return "", fmt.Errorf("job cancelled due to shutdown")
	}
}
