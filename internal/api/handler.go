package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/diggiex47/distributed-task-queue/internal/job"
	"github.com/diggiex47/distributed-task-queue/internal/redisstore"
	"github.com/google/uuid"
)

// Handler holds the dependencies our HTTP handlers need
// Notice it holds a *redisstore.Store - this dependency Injection.
// Instead of creating a Redis connection inside each handler function,
// we pass it in from outside.

type Handler struct {
	store *redisstore.Store
}

// New creates a Handler with ecerything it needs to do its job
func New(store *redisstore.Store) *Handler {
	return &Handler{store: store}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

// RegisterRoutes wire URL paths to handler functions.
// mux is short for "multiplexer" - it's Go's built in URL router.
// Think  of it like a switchboard - incoming request hits a URL, mux looks up who handles it, calls that function.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /jobs", h.SubmitJob)
	mux.HandleFunc("GET /jobs/{id}", h.GetJobStatus)
	mux.HandleFunc("GET /health", h.Health)
}

// SubmitJob handles POST/jobs
// Client sends: { "payload": "some-task"}
// we respod: { "id": "uuid", "status": "pending", ...}
func (h *Handler) SubmitJob(w http.ResponseWriter, r *http.Request) {
	// step 1: Read and parse the JSON body from the request
	var req struct {
		Payload string `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Payload == "" {
		writeError(w, http.StatusBadRequest, "payload is required")
		return
	}

	// step 2: create a new job with a unique ID
	// uuid.New().String() generates a universally wnique ID like:
	//"a8f2h4n2-1c2b-4d3e-9f5a-6b7c8d9e0f1g"
	// no 2 uuid are ever the same

	now := time.Now().UTC()
	j := &job.Job{
		ID:        uuid.New().String(),
		Status:    job.StatusPending,
		Payload:   req.Payload,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// step 3: enqueue the job into Redis
	if err := h.store.EnqueueJob(r.Context(), j); err != nil {
		log.Printf("failed to enqueue job: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to enqueue job")
		return
	}

	log.Printf("API: enqueue job %s (payload: %s)", j.ID, j.Payload)

	// step 4: respond with 202 Accepted - not 200 ok
	// 202 means "I received your request, woek will happen asynchronously"
	// 200 would mean " I alreasy did the work" - which isn't true yet
	writeJSON(w, http.StatusAccepted, j)
}

// GetJobStatus handles GET /jobs/{id}
// Client sends: GET /jobs/a8f2h4n2-1c2b-4d3e-9f5a-6b7c8d9e0f1g
// we respond with the current job state from redis

func (h *Handler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	// r.PathValue("id") extracts the {id} part from the URL

	jobID := r.PathValue("id")
	if jobID == "" {
		writeError(w, http.StatusBadRequest, "job ID required")
		return
	}

	j, err := h.store.GetJob(r.Context(), jobID)
	if err != nil {
		log.Printf("failed to get job %s: %v", jobID, err)
		writeError(w, http.StatusInternalServerError, "failed to retrieve job")
		return
	}

	// nil means the job diesn't exist in Redis
	if j == nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	writeJSON(w, http.StatusOK, j)
}

// health handles GET /health
// used by Docker, load balancers, and kubernetes to check if app is alive
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// writeJSON is a helper that converts any value to JSON
// aaas an HTTP response. used by every handler above.

// writeError is ahelprer for sending consistent wrror response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
