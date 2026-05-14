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

### Check Job Status
```bash
curl http://localhost:8080/jobs/{id}
```

---

## Tech Stack

| Technology | Role |
|---|---|
| **Go** | REST API, worker pool, graceful shutdown |
| **Redis** | Message queue + job result store |
| **Docker & Compose** | Containerized deployment |

---

## Project Status

| Phase | Status |
|---|---|
| Job Model | ✅ Complete |
| Config Layer | ✅ Complete |
| Redis Store | ✅ Complete |
| Worker Pool | ✅ Complete |
| REST API | ✅ Complete |
| Graceful Shutdown | 🚧 In Progress |
| Docker + Compose | ⏳ Pending |

---

## How To Run

```bash
# Start Redis
docker start redis-local

# Run the server
go run ./cmd/server/main.go
```

---

## Interview Talking Points

- Producer-Consumer pattern decouples job submission from processing
- BRPOP provides zero-CPU blocking — workers sleep until work arrives
- Goroutines cost ~2KB vs ~1MB for OS threads — enables high concurrency
- Repository pattern isolates all Redis logic in one place
- 202 Accepted response signals async processing to clients
- Graceful shutdown via context cancellation ensures no job is lost on exit