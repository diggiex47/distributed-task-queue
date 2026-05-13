package config

import ( 
	"fmt"
	"os"
)

type Config struct {

	// RedisAddr is the host:port of our Redis server.
	RedisAddr string

	// RedisPass is the Redis password. Empty string means no password.
	RedisPass string

	// QueueName is the name of the Redis list we use as our task queue.
	QueueName string

	// WorkerCount controls how many goroutines process jobs simultaneously.
	// More workers = more parallelism, but also more Redis connections.
	WorkerCount int

	// APIPort is the port the HTTP server listens on.
	APIPort string

}



// Load reads all configuration from environment variables and returns
// a fully populated Config struct.
// This function is the ONLY place in the entire codebase that calls os.Getenv.
// Everything else just receives a *Config and reads from it.

func Load() *Config {
	return &Config{
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass: getEnv("REDIS_PASS", ""),
		QueueName: getEnv("QUEUE_NAME", "task_queue"),
		WorkerCount: getEnvInt("WORKER_COUNT", 5),
		APIPort: getEnv("API_PORT", "8080"),
	
	}
}


// getEnv reads an environment variable by key. If the variable is not set
// or is empty, it returns the defaultVal instead.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val !="" {
		return val
	}
	return defaultVal
}



// getEnvInt is like getEnv but converts the result to an integer.
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var result int 
		if _, err := fmt.Sscan(val, &result); err == nil {
			return result
		}
	}
	return defaultVal
}