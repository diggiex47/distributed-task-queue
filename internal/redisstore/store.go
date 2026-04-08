package redisstore

import (
	"context"
	"time"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/diggiex47/distributed-task-queue/internal/job"
)


type Store struct {
	client *redis.Client
	queueName string
}

func New(addr, password, queueName string) (*Store, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		Password: password,
		DB: 0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()


	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cannot connect to redis at %s: %w", addr, err)
	}

	return &Store{client: client, queueName: queueName}, nil
}



func (s *Store) EnqueueJob(ctx context.Context, j *job.Job) error {
	pipe := s.client.Pipeline()

	jobKey := "job:" + j.ID
	pipe.HSet(ctx, jobKey, map[string]interface{}{
		"id": j.ID,
		"status": string(j.Status),
		"payload": j.Payload,
		"result": "",
		"error": "",
		"created_at": j.CreatedAt.Format(time.RFC3339),
		"updated_at": j.UpdatedAt.Format(time.RFC3339),
	})

	pipe.Expire(ctx, jobKey, 24*time.Hour)

	pipe.LPush(ctx, s.queueName, j.ID)
	_, err := pipe.Exec(ctx)
	return err
}


func (s *Store) DequeueJob(ctx context.Context) (*job.Job, error) {

	result, err := s.client.BRPop(ctx, 0, s.queueName).Result()
	if err != nil {

		if err == context.Canceled {
			return nil, nil

		}
		return nil, fmt.Errorf("BRPOP failed: %w", err)
	}

	jobID := result[1]
	return s.GetJob(ctx, jobID)
}

func (s *Store) UpdateJobStatus(
	ctx context.Context,
	jobID string,
	status job.Status,
	result string,
	errMsg string,
) error {
	return s.client.HSet(ctx, "job:"+jobID, map[string]interface{} {
		"status": string(status),
		"result": result,
		"error": errMsg,
		"updated_at": time.Now().Format(time.RFC3339),
			}).Err()
}


func (s *Store) GetJob(ctx context.Context, jobID string)  (*job.Job, error) {
	data, err := s.client.HGetAll(ctx, "job:"+jobID).Result()
	if err != nil {
		return nil, fmt.Errorf("HGETALL failed: %w", err)
	}

	if len(data) == 0 {
		return nil, nil 
	}

	createdAt, _ := time.Parse(time.RFC3339, data["created_at"])
	updatedAt, _ := time.Parse(time.RFC3339, data["updated_at"])


	return &job.Job {
		ID: data["id"],
		Status: job.Status(data["status"]),
		Payload: data["payload"],
		Result: data["result"],
		Error: data["error"],
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil 
}


func (s *Store) Close() error {
	return s.client.Close()
}
