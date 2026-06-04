# Distributed Task Queue

A production-style distributed task queue built with **Go**, **Redis**, and **Docker**.

Clients submit long-running tasks via a REST API. Tasks are queued in Redis
and processed asynchronously by a pool of worker goroutines. Results are
stored in Redis and retrievable by job ID.

---

## Architecture

Client → REST API → Redis Queue → Worker Pool → Redis Result Store

**Pattern**: Producer-Consumer  
**Queue**: Redis List (LPUSH / BRPOP)  
**Job Store**: Redis Hash (HSET / HGETALL)  
**Concurrency**: Go goroutines with WaitGroup and context cancellation
**Deployment**: Docker Compose

---

## Quick Start

```bash
docker-compose up --build
```

Server runs on http://localhost:8080

---


## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| POST | /jobs | Submit a new job |
| GET | /jobs/{id} | Check job status |
| GET | /health | Health check |

### Submit a Job
```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"payload": "your-task-here"}'
```

Response (202 Accepted):
```json
{
  "id": "a8f3d2c1-...",
  "status": "pending",
  "payload": "your-task-here",
  "created_at": "2026-06-04T16:20:00Z"
}
```

### Check Job Status
```bash
curl http://localhost:8080/jobs/{id}
```

Response:
```json
{
  "id": "a8f3d2c1-...",
  "status": "completed",
  "result": "processed 'your-task-here' successfully"
}
```

---

## Tech Stack

| Technology | Role |
|---|---|
| **Go 1.22** | REST API, worker pool, graceful shutdown |
| **Redis 7** | Message queue + job result store |
| **Docker** | Containerization |
| **Docker Compose** | Multi-container orchestration |

---

## Project Status

| Phase | Status |
|---|---|
| Job Model | ✅ Complete |
| Config Layer | ✅ Complete |
| Redis Store | ✅ Complete |
| Worker Pool | ✅ Complete |
| REST API | ✅ Complete |
| Graceful Shutdown | ✅ Complete |
| Docker + Compose | ✅ Complete |

---

## How To Run
## Local Development (Without Docker)

```bash
# Start Redis manually
docker start redis-local

# Run server
go run ./cmd/server/main.go

# In another terminal, test
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"payload": "test"}'
```

---

## How It Works (Step by Step)

1. **Client submits job** via `POST /jobs`
2. **API creates UUID**, enqueues to Redis, returns 202 Accepted
3. **Worker wakes up** (BRPOP unblocks), picks up job ID
4. **Worker fetches job details** from Redis hash
5. **Worker processes job** (calls business logic)
6. **Worker updates status** to "completed" or "failed"
7. **Client polls** `GET /jobs/{id}` to check result

All while remaining decoupled — no worker talks to the API, no API talks to workers. Redis is the coordinator.

---

## Interview Talking Points

**Architecture & Design**
- Producer-Consumer pattern decouples job submission from processing
- Repository pattern isolates all Redis logic — easy to swap backends
- Single Responsibility — each package does one thing well

**Concurrency**
- Goroutines cost ~2KB vs ~1MB for OS threads — enables high concurrency
- BRPOP provides zero-CPU blocking — workers sleep until work arrives
- Context cancellation propagates shutdown signal through entire call stack

**Production Readiness**
- 202 Accepted response signals async processing to clients
- Graceful shutdown ensures no job is lost mid-task
- Health check endpoint for monitoring and load balancers
- Environment-based configuration — no hardcoded values
- Docker Compose proves the system works in a containerized environment

**Potential Enhancements**
- Heartbeat-based reaper for at-least-once delivery guarantees
- Job retries with exponential backoff
- Prometheus metrics for monitoring worker throughput and latency
- Job priorities using multiple Redis queues
- Idempotency keys to prevent duplicate processing

---