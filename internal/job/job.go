package job

import "time"

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed" 
)

type Job struct {
	ID  string `json:"id"`
	Status Status `json:"status"`
	Payload string `json:"payload"`
	Result string `json:"result"`
	Error string `json:"error"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

//Is complete returns true if this job finished successfully.
func (j Job) IsComplete() bool {
	return j.Status == StatusCompleted
}

//Is failed returns true if this job encountered an error.
func (j Job) IsFailed() bool {
	return j.Status == StatusFailed
}

//Age returns how long ago this job was created
//time.Since() is standard library function that calculatethe duration between past time and right now.
func (j Job) Age() time.Duration {
	return time.Since(j.CreatedAt)
}